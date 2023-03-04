package container

import (
	log "github.com/sirupsen/logrus"
	"github.com/yang-wang11/mydocker/cgroups"
	"github.com/yang-wang11/mydocker/common"
	"github.com/yang-wang11/mydocker/network"
	"strconv"
	"strings"
	"time"
)

func Run(con *common.Container) {

	// prepare aufs, namespace ..
	child, writePipe := NewProcess(con)
	if child == nil {
		UpdateContainerStatus(con, common.ContainerFailed)
		log.Warnf("New process init failed. ")
		return
	}

	// start to run command
	if err := child.Start(); err != nil {
		UpdateContainerStatus(con, common.ContainerFailed)
		return
	} else {
		con.Pid = strconv.Itoa(child.Process.Pid)
		log.Debugf("get pid: %v. ", child.Process.Pid)
	}

	// setup cgroup for child
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(con.ResConfig)
	cgroupManager.Apply(child.Process.Pid)

	// setup network
	if err := network.Connect(con); err != nil {
		log.Errorf("connect to network %s failed, %v", con.Network, err)
		return
	}

	// send command
	writePipe.WriteString(con.Command)
	writePipe.Close()

	// persistent container info
	PersistContainerInfo(con)

	// check child thread
	if err := common.CheckPidIsRunning(child.Process.Pid); err != nil {
		UpdateContainerStatus(con, common.ContainerStopped)
	} else {
		UpdateContainerStatus(con, common.ContainerRunning)
	}

	if con.TtyMode {
		child.Wait()
		UpdateContainerStatus(con, common.ContainerStopped)
	} else {
		time.Sleep(1 * time.Second)
	}

	log.Debugln("main process exit.")
}

func volumeMapSpliter(volumeMap string) (match bool, ParentPath, ContainerPath string) {
	// init
	ParentPath, ContainerPath, match = "", "", false
	// split
	volumeArr := strings.Split(volumeMap, ":")
	if len(volumeArr) == 2 {
		ParentPath = volumeArr[0]
		ContainerPath = volumeArr[1]
		if ParentPath != "" && ContainerPath != "" {
			match = true
		}
	}
	return
}

func CheckContainer(ContainerName string) bool {

	// check container name
	foundContainer := false

	for _, con := range ListContainers(false) {
		if con == ContainerName {
			foundContainer = true
			break
		}
	}

	return foundContainer
}
