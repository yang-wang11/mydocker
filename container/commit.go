package container

import (
	"github.com/yang-wang11/mydocker/common"
	"path"

	log "github.com/sirupsen/logrus"
)

// backup container to tar file
func CommitContainer(containerName, imageName string) {

	containerBasePath := path.Join(ContainerBaseFolder, containerName)
	containerBaseMnt := path.Join(containerBasePath, MntLayer)

	if !common.PathExists(containerBaseMnt) {
		log.Errorf("commit container failed. path %s not exist", containerBaseMnt)
		return
	}

	// image path
	targetTar := path.Join(ImageBaseFolder, imageName)

	if common.Cmder("tar", false, nil, "-czf", targetTar, "-C", containerBaseMnt, ".") {
		log.Debugf("tar %s successfully. ", targetTar)
	} else {
		log.Errorf("tar %s failed. ", targetTar)
	}

}
