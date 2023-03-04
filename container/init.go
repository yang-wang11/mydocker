package container

import (
	"fmt"
	"github.com/yang-wang11/mydocker/common"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

//const ShPath string = "/bin/sh"

func RunContainerInitProcess(containerName string) error {

	cmdArray := readUserCommand()
	if len(cmdArray) == 0 {
		return common.NewError("no command found!!!")
	} else {
		RuntimeLogger.Debugf("user command: %v", cmdArray)
	}

	if err := setUpMount(containerName); err != nil {
		return common.NewError("container init mount failed!")
	}

	SetEnvPATH()

	// check command
	Command, err := exec.LookPath(cmdArray[0])
	if err != nil {
		return common.NewError("command '%s' not found", cmdArray[0])
	} else {
		RuntimeLogger.Debugf("command %s found", Command)
	}

	if err := syscall.Exec(Command, cmdArray, os.Environ()); err != nil {
		RuntimeLogger.Errorf("call %s failed", cmdArray[0])
	} else {
		RuntimeLogger.Debugf("call %s successfully", cmdArray[0])
	}

	return nil

}

func NewProcess(con *common.Container) (*exec.Cmd, *os.File) {

	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", []string{"containers", "init", "--name", con.ContainerName}...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	// set hostname
	hostname := "mydocker"
	if err := assignHost(hostname); err != nil {
		log.Errorf("set hostname failed, %v ", err.Error())
	} else {
		log.Debugf("set hostname to %s. ", hostname)
	}

	// if enable -it
	var fp *os.File

	if con.TtyMode {

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Debugf("redirect log to standard .")

	} else {

		cmd.Stdin = os.Stdin

		// redirect stdout to log file
		fp = FeedContainerLogPointer(con)
		if fp != nil {
			log.Debugf("redirect log to file pointer successfully. %v", fp)
			//_, err = io.WriteString(fp, "tester")
			//if err != nil {
			//  log.Errorf("write failed %v", err)
			//}
			cmd.Stdin = os.Stdin
			cmd.Stdout = fp
		} else {
			log.Errorf("redirect log to file pointer failed.")
		}

	}

	// feed readpipe to child thread
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), con.Env...)

	// prepare container root path
	NewWorkSpace(con)

	// set union layer as container's root path
	cmd.Dir = path.Join(con.ContainerPath, MntLayer)

	log.Debug("prepare ParentProcess finished.")

	return cmd, writePipe

}

func assignHost(name string) error {

	if err := syscall.Sethostname([]byte(name)); err != nil {
		return fmt.Errorf("set hostname failed, %s", err)
	}

	return nil

}

func SetEnvPATH() {
	// set env variable PATH
	os.Setenv("PATH", "/bin:/sbin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin")
}

func readUserCommand() []string {
	// Stdin = 0  Stdout = 1  Stderr = 2   3(append): readpipe
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := io.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func setUpMount(containerName string) error {

	var err error

	// get current path
	pwd, err := os.Getwd()
	if err != nil {
		return common.NewError("Get current location error %v", err)
	} else {
		//RuntimeLogger.Debugf("Current location is %s", pwd)
	}

	// pivot_root命令用于将根目录替换为指定目录
	if err = pivotRoot(pwd, containerName); err != nil {
		return common.NewError("pivotRoot failed, err: %v", err)
	}

	return nil

}

func pivotRoot(root, containerName string) error {

	var err error

	/**
	  为了使当前root的老 root 和新 root 不在同一个文件系统下，我们把root重新mount了一次
	  bind mount是把相同的内容换了一个挂载点的挂载方法
	   mount -o bind root root
	*/
	if err = syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		log.Errorf("Mount rootfs to itself failed, %v. ", err)
		return err
	}

	// create rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if _, err := os.Stat(pivotDir); os.IsNotExist(err) {
		if err = os.MkdirAll(pivotDir, 0700); err != nil {
			log.Errorf("create folder %s failed. ", pivotDir)
			return err
		}
	}
	log.Debugf("pivot Dir: %s", pivotDir)

	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示声明要这个新的mount namespace独立。
	// 首先下载busybox镜像，然后接下放到/root/busybox下
	if err = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return common.NewError("mount / failed, %v", err.Error())
	} else {
		log.Debug("mount / successfully.")
	}

	// pwd: /root/mydockerspace/containers/1642393019/mnt
	// container Path
	containerBasePath := path.Join(ContainerBaseFolder, containerName)
	containerBaseMnt := path.Join(containerBasePath, MntLayer)

	// mount host proc to container
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", path.Join(containerBaseMnt, "/proc"), "proc", uintptr(defaultMountFlags), ""); err != nil {
		return common.NewError("mount proc %s failed, %v", path.Join(containerBaseMnt, "/proc"), err.Error())
	} else {
		log.Debug("mount proc successfully.")
	}

	// tmpfs is one of ram filesystem, use ram/swap to store data
	if err := syscall.Mount("tmpfs", path.Join(containerBaseMnt, "/dev"), "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return common.NewError("mount tmpfs %s failed, %v", path.Join(containerBaseMnt, "/dev"), err.Error())
	} else {
		log.Debug("mount tmpfs successfully.")
	}

	if err = syscall.PivotRoot(root, pivotDir); err != nil {
		log.Debugf("call pivot_root failed, root:%s oldroot:%s. ", root, pivotDir)
		return err
	} else {
		log.Debug("call pivot_root successfully.")
	}

	// 修改当前的工作目录到根目录
	if err = syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	// 卸载老mount点
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}

	// 删除临时文件夹
	return os.Remove(pivotDir)
}

func InitSystemFolder() error {

	// init image folder
	if !common.PathExists(ImageBaseFolder) {
		if err := os.MkdirAll(ImageBaseFolder, 0644); err != nil {
			return common.NewError("failed to create the image's folder.", ImageBaseFolder)
		}
	}

	// init container folder
	if !common.PathExists(ContainerBaseFolder) {
		if err := os.MkdirAll(ContainerBaseFolder, 0644); err != nil {
			return common.NewError("failed to create the container's folder.", ContainerBaseFolder)
		}
	}

	return nil
}
