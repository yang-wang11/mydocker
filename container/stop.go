package container

import (
	. "docker/mydocker/util"
	log "github.com/Sirupsen/logrus"
	"strconv"
	"syscall"
)

func StopContainer(containerName string) {

	con := GetContinerInfoByName(containerName)

	// get pid from config file
	ConPid, err := strconv.Atoi(con.Pid);
	if err != nil {
		log.Errorf("pid %s illegal", ConPid)
		return
	}

	// check child thread
	childPid, err := strconv.Atoi(con.Pid)
	if err != nil {
		log.Errorf("parse pid %s to integer failed", con.Pid)
	}
	// process already stopped
	if err := CheckPidIsRunning(childPid); err != nil {
		UpdateContainerStatus(&con, ContainerStopped)
		return
	}

	// kill process  syscall.SIGKILL
	if err := syscall.Kill(ConPid, syscall.SIGTERM); err != nil {
		log.Errorf("kill pid %d failed", ConPid)
		return
	}

	UpdateContainerStatus(&con, ContainerStopped)
	log.Debugf("kill pid %d successfully", ConPid)

}
