package container

import (
	"docker/mydocker/network"
	. "docker/mydocker/util"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

func DeleteContainer(containerName string, delAll bool) {

	if containerName != "" {
		deleteContainer(containerName)
	} else if delAll{
		for _, conName := range ListContainers(false) {
			deleteContainer(conName)
		}
	}

}

func deleteContainer(containerName string){

	// delete persistent folder
	con := GetContinerInfoByName(containerName)
	DeleteWorkSpace(&con)

	// delete runtime folder
	time.Sleep(100 * time.Millisecond)
	delContainerRuntime(containerName)

	// release ip
	ip, cidr, err := net.ParseCIDR(con.IP+"/24")
	if err != nil {
		log.Errorf("parse ip %s failed", con.IP)
	}
	err = network.IpAllocator.Release(cidr, &ip)
	if err != nil {
		log.Errorf("release ip of container %s failed", con.Id)
	} else {
		log.Debugf("release ip %v of container %s successfully", con.IP, con.Id)
	}

	// release DNAT rule
	for _, portmap := range con.PortMapping {
		 portArray := strings.Split(portmap, ":")
		 hostPort, containerPort := portArray[0], portArray[1]
		 iptableDNAT := fmt.Sprintf("-t nat -D PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s",  hostPort, con.IP, containerPort)
		 if !Cmder("iptables", true, nil, strings.Split(iptableDNAT, " ")...) {
			 log.Errorf("release DNAT rule of container %s failed", con.Id)
		 }
	}

	log.Debugf("Container %s deleted. ", containerName)

}

func delContainerRuntime(containerName string){

	ContainerRuntimeFullPath := path.Join(ContainerRuntimeBaseFolder, containerName)
	if PathExists(ContainerRuntimeFullPath) {
		if err := os.RemoveAll(ContainerRuntimeFullPath); err != nil {
			log.Errorf("Container %s runtime delete failed. %v", containerName, err)
			return
		}
	}

}