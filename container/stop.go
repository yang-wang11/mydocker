package container

import (
	log "github.com/sirupsen/logrus"
	"github.com/yang-wang11/mydocker/common"
	"strconv"
	"syscall"
)

func StopContainer(containerName string) {

	con := GetContinerInfoByName(containerName)

	// get pid from config file
	ConPid, err := strconv.Atoi(con.Pid)
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
	if err := common.CheckPidIsRunning(childPid); err != nil {
		UpdateContainerStatus(&con, common.ContainerStopped)
		return
	}

	// kill process  syscall.SIGKILL
	if err := syscall.Kill(ConPid, syscall.SIGTERM); err != nil {
		log.Errorf("kill pid %d failed", ConPid)
		return
	}

	UpdateContainerStatus(&con, common.ContainerStopped)
	log.Debugf("kill pid %d successfully", ConPid)
}
