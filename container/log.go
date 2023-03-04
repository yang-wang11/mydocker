package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/yang-wang11/mydocker/common"
	"io"
	"os"
	"path"
)

func GrabContainerLog(containerName string) error {

	ContainerRuntimeLogFolder := path.Join(ContainerRuntimeBaseFolder, containerName)
	ContainerRuntimeLogPath := path.Join(ContainerRuntimeLogFolder, containerLogName)
	fp, err := os.Open(ContainerRuntimeLogPath)
	if err != nil {
		return common.NewError("open file %s failed, %v", ContainerRuntimeLogPath, err)
	}

	logs, err := io.ReadAll(fp)
	if err != nil {
		return common.NewError("read file %s failed, %v", ContainerRuntimeLogPath, err)
	}

	if _, err := fmt.Fprint(os.Stdout, string(logs)); err != nil {
		return common.NewError("print log failed")
	}

	return nil

}

func FeedContainerLogPointer(con *common.Container) (fp *os.File) {

	var err error

	// create container folder if not exist
	ContainerRuntimePath := path.Join(ContainerRuntimeBaseFolder, con.ContainerName)
	if !common.PathExists(ContainerRuntimePath) {
		if err := os.MkdirAll(ContainerRuntimePath, 0644); err != nil {
			log.Errorf("create folder %s failed, %v", ContainerRuntimePath, err)
			return
		}
	} else {
		log.Debugf("folder %s exist.", ContainerRuntimePath)
	}

	// open/create file
	ContainerLogPath := path.Join(ContainerRuntimePath, containerLogName)
	if !common.PathExists(ContainerLogPath) {
		if fp, err = os.Create(ContainerLogPath); err != nil {
			log.Errorf("Log: create file %s failed, %v", ContainerRuntimePath, err)
			return
		} else {
			log.Debugf("Log: create file %s successfully. ", ContainerRuntimePath)
		}
	} else {
		if fp, err = os.OpenFile(ContainerLogPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend); err != nil {
			log.Errorf("Log: open file %s failed, %v", ContainerRuntimePath, err)
			return
		} else {
			log.Debugf("Log: open file %s successfully. ", ContainerRuntimePath)
		}
	}

	return fp

}
