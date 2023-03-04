package container

import (
	"github.com/yang-wang11/mydocker/common"
	"path"
	"strconv"
	"time"
)

// new container
func NewContainerInfo(
	pid int,
	imageName string,
	containerName *string,
	network string,
	volumeMap string,
	ttyMode bool,
	command string,
	env []string,
	portMap []string,
	resConf *common.ResourceConfig) common.Container {

	// prepare info of container
	NewId := GeneratorID(10)

	if *containerName == "" {
		*containerName = NewId
	}

	return common.Container{
		Id:            NewId,
		Pid:           strconv.Itoa(pid),
		Command:       command,
		Network:       network,
		ContainerName: *containerName,
		ContainerPath: path.Join(ContainerBaseFolder, *containerName),
		Env:           env,
		PortMapping:   portMap,
		ImageName:     imageName,
		CreateTime:    time.Now().Format("2006/1/2 15:04:05"),
		Status:        common.ContainerCreating,
		VolumeMap:     volumeMap,
		TtyMode:       ttyMode,
		ResConfig:     resConf,
	}

}

//
//// get persistent logger
//func GetPersistentLogger(containerName string) *log.Logger {
//
//	ContainerRuntimeLogFolder := path.Join(ContainerRuntimeBaseFolder, containerName)
//	ContainerRuntimeLog := path.Join(ContainerRuntimeLogFolder, containerLogName)
//	if !common.PathExists(ContainerRuntimeLogFolder) {
//		if err := os.MkdirAll(ContainerRuntimeLogFolder, 0644); err != nil {
//			log.Errorf("GetPersistentLogger: create folder %s failed. %v", ContainerRuntimeLogFolder, err)
//			return nil
//		}
//	}
//
//	logger, err := os.OpenFile(ContainerRuntimeLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
//	if err != nil {
//		log.Errorf("GetPersistentLogger: open file failed: %v\n", err)
//	}
//
//	return &log.Logger{
//		Out:       io.MultiWriter(logger, os.Stdout),
//		Level:     log.DebugLevel,
//		Formatter: &log.JSONFormatter{},
//	}
//
//}

func UpdateContainerStatusByName(containerName string, status common.ContainerStatus) {

	con := GetContinerInfoByName(containerName)

	UpdateContainerStatus(&con, status)

}

// UpdateContainerStatus : update the status of container
func UpdateContainerStatus(con *common.Container, status common.ContainerStatus) {

	con.Status = status
	con.ModifyTime = time.Now().Format("2006/1/2 15:04:05")

	if status == common.ContainerStopped || status == common.ContainerFailed {
		con.StopTime = con.ModifyTime
	}

	PersistContainerInfo(con)
}
