package container

import (
	. "docker/mydocker/util"
	log "github.com/Sirupsen/logrus"
	"os"
)

// mydocker location setting
const (
	ImageBaseFolder     string = RootDir + "/images"
	ContainerBaseFolder string = RootDir + "/containers"
	MntLayer            string = "mnt"
	WriteLayer          string = "write"
	ReadLayer           string = "read"
)

const (
	// container runtime
	ContainerRuntimeBaseFolder string = "/var/run/mydocker"
	ContainerConfigName        string = "config.json"
	containerLogName           string = "container.log"
)

func init(){

	// container runtime
	if err := os.MkdirAll(ContainerRuntimeBaseFolder, os.ModeDir); err != nil {
		log.Errorf("create folder %s failed, %v", ContainerRuntimeBaseFolder, err)
	}

	// mydocker config
	InitSystemFolder()

}