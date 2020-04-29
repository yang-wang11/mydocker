package container

import (
	. "docker/mydocker/util"
	"fmt"
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"os"
	"path"
)

func GrabContainerLog(containerName string) error {

	ContainerRuntimeLogFolder := path.Join(ContainerRuntimeBaseFolder, containerName)
	ContainerRuntimeLogPath := path.Join(ContainerRuntimeLogFolder, containerLogName)
	fp, err := os.Open(ContainerRuntimeLogPath)
	if err != nil {
		return NewError("open file %s failed, %v", ContainerRuntimeLogPath, err)
	}

	logs, err := ioutil.ReadAll(fp)
	if err != nil {
		return NewError("read file %s failed, %v", ContainerRuntimeLogPath, err)
	}

	if _, err := fmt.Fprint(os.Stdout, string(logs)); err != nil {
		return NewError("print log failed");
	}

	return nil

}


func FeedContainerLogPointer(con *Container) (fp *os.File) {

	var err error

	// create container folder if not exist
	ContainerRuntimePath := path.Join(ContainerRuntimeBaseFolder, con.ContainerName)
	if !PathExists(ContainerRuntimePath) {
		if err := os.MkdirAll(ContainerRuntimePath, 0644); err != nil {
			log.Errorf("create folder %s failed, %v", ContainerRuntimePath, err)
			return
		}
	} else {
		log.Debugf("folder %s exist.", ContainerRuntimePath)
	}

	// open/create file
	ContainerLogPath := path.Join(ContainerRuntimePath, containerLogName)
	if !PathExists(ContainerLogPath) {
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
