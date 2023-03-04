package container

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {

	// container runtime
	if err := os.MkdirAll(ContainerRuntimeBaseFolder, os.ModeDir); err != nil {
		log.Errorf("create folder %s failed, %v", ContainerRuntimeBaseFolder, err)
	}

	// mydocker config
	InitSystemFolder()
}
