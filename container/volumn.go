package container

import (
	log "github.com/sirupsen/logrus"
	"github.com/yang-wang11/mydocker/common"
	"os"
	"path"
)

// Create a AUFS filesystem as container root workspace
func NewWorkSpace(con *common.Container) {

	// build container filesystem
	CreateReadOnlyLayer(con.ContainerPath, con.ImageName)
	CreateWriteLayer(con.ContainerPath)
	CreateMountPoint(con.ContainerPath)

	// prepare -v
	match, ParentPath, ContainerPath := volumeMapSpliter(con.VolumeMap)
	log.Debugf("volume mapping: hostPath: %v, containerPath: %v \n", ParentPath, ContainerPath)
	if match {
		MountVolume(path.Join(con.ContainerPath, MntLayer), ParentPath, ContainerPath)
	}

}

// copy image to container's read layer
func CreateReadOnlyLayer(containerPath, imageName string) error {

	// init variables
	containerReadLayer := path.Join(containerPath, ReadLayer)

	// image path
	imagePath := path.Join(ImageBaseFolder, imageName)
	SetupFlag := false

	// judge ReadOnlyLayer is exist
	if !common.PathExists(containerReadLayer) {
		if err := os.MkdirAll(containerReadLayer, 0644); err != nil {
			return common.NewError("make dir %s error. %v", containerReadLayer, err)
		}
	}

	// prepare rootfs (extract the filesystem from image)
	if common.Cmder("tar", false, nil, "-xvf", imagePath, "-C", containerReadLayer) {
		SetupFlag = true
	} else {
		return common.NewError("untar %s failed.", imagePath)
	}

	// if setup failed
	if !SetupFlag {
		return common.NewError("setup image %s failed. ", imageName)
	}

	return nil

}

func CreateWriteLayer(containerPath string) error {
	// setup container write layer
	containerWriteLayer := path.Join(containerPath, WriteLayer)
	if err := os.MkdirAll(containerWriteLayer, 0777); err != nil {
		return common.NewError("make dir %s error. %v", containerWriteLayer, err)
	}
	return nil
}

func CreateMountPoint(containerPath string) error {
	// setup container union layer
	containerUnionLayer := path.Join(containerPath, MntLayer)
	if err := os.MkdirAll(containerUnionLayer, 0777); err != nil {
		return common.NewError("make dir %s error. %v", containerUnionLayer, err)
	}
	// combine writelayer & readerlayer to union layer
	dirs := "dirs=" + path.Join(containerPath, WriteLayer) + ":" + path.Join(containerPath, ReadLayer) + "=ro"
	mountCmdstr := []string{"mount", "-t", "aufs", "-o", dirs, "none", containerUnionLayer}
	//log.Debugln("Create Mount Command: ", strings.Join(mountCmdstr, " "))
	common.Cmder("mount", false, nil, mountCmdstr[1:]...)
	return nil
}

// Delete the AUFS filesystem while container exit
func DeleteWorkSpace(con *common.Container) {
	DeleteMountPoint(path.Join(con.ContainerPath, MntLayer), con.VolumeMap)
	DeleteLayer(con.ContainerPath)
}

func DeleteMountPoint(mntPath, volumeMap string) {

	// if enable volume, unmount container folder first!!
	match, _, ContainerVolumeMapPath := volumeMapSpliter(volumeMap)
	if match {
		ContainerFullVolumeMapPath := path.Join(mntPath, ContainerVolumeMapPath)
		common.Cmder("umount", false, nil, ContainerFullVolumeMapPath)
	}

	// unmount union layer
	common.Cmder("umount", false, nil, mntPath)

	// delete union folder
	if err := os.RemoveAll(mntPath); err != nil {
		log.Errorf("Remove dir %s error %v", mntPath, err)
	}

}

func DeleteLayer(rootPath string) error {

	if err := os.RemoveAll(rootPath); err != nil {
		return common.NewError("delete dir %s error %v", rootPath, err)
	}
	return nil
}

func MountVolume(mntPath string, hostPath, conPath string) error {

	// create parent folder(in host) if not exist
	if err := os.MkdirAll(hostPath, 0777); err != nil {
		return common.NewError("make host dir %s error. %v", hostPath, err)
	} else {
		log.Debugf("make host dir %s successfully. ", hostPath)
	}
	// create container folder if not exist
	containerVolumePath := path.Join(mntPath, conPath)
	if err := os.MkdirAll(containerVolumePath, 0777); err != nil {
		return common.NewError("make container dir %s error. %v", containerVolumePath, err)
	} else {
		log.Debugf("make container dir %s successfully. ", containerVolumePath)
	}
	dirs := "dirs=" + hostPath
	common.Cmder("mount", false, nil, "-t", "aufs", "-o", dirs, "none", containerVolumePath)

	return nil
}
