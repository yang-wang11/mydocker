package common

import (
	log "github.com/sirupsen/logrus"
)

const (
	RootDir = "/root/mydockerspace"

	// network settings
	DefNetworkName   = "mybridge"
	DefNetworkSubnet = "192.168.0.1/24"
	DefDriveType     = BridgeDrive

	// location settings
	ImageBaseFolder     string = RootDir + "/images"
	ContainerBaseFolder string = RootDir + "/containers"
	MntLayer            string = "mnt"
	WriteLayer          string = "write"
	ReadLayer           string = "read"

	// container runtime
	ContainerRuntimeBaseFolder string = "/var/run/mydocker"
	ContainerConfigName        string = "config.json"
	ContainerLogName           string = "container.log"
)

// network drive
type DriveType string

const (
	BridgeDrive DriveType = "bridge"
)

type ContainerStatus string

const (
	ContainerCreating ContainerStatus = "CREATING"
	ContainerRunning  ContainerStatus = "RUNNING"
	ContainerStopped  ContainerStatus = "STOP"
	ContainerFailed   ContainerStatus = "FAILED"
)

// supported subsystem
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Container struct {
	Id            string          `json:"id"`
	Pid           string          `json:"pid"`
	Command       string          `json:"command"`
	Env           []string        `json:"env"`
	Network       string          `json:"network"`
	NetworkDevice string          `json:"network_device"`
	ImageName     string          `json:"image_name"`
	ContainerName string          `json:"container_name"`
	CreateTime    string          `json:"create_time"`
	ModifyTime    string          `json:"modify_time"`
	StopTime      string          `json:"stop_time"`
	Status        ContainerStatus `json:"status"`
	VolumeMap     string          `json:"volume_map"`
	ContainerPath string          `json:"container_path"`
	TtyMode       bool            `json:"tty_mode"`
	ResConfig     *ResourceConfig `json:"res_config"`
	PortMapping   []string        `json:"portmapping"`
	IP            string          `json:"ip"`
}

var RuntimeLogger *log.Logger
