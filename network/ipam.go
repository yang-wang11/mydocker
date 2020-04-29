package network

import (
	. "docker/mydocker/util"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
)

type IPAM struct {
	SubnetPath string
	Mux        *sync.RWMutex
	Subnets    *map[string]string // map[subnet]bits
}

func NewIPAM() *IPAM {

	subs := make(map[string]string, 1)

	return &IPAM{
		Mux:        &sync.RWMutex{},
		SubnetPath: defIPAMConf,
		Subnets:    &subs ,
	}

}

func (ipam *IPAM) init() error {

	// create if not exist
	if !PathExists(defIPAMPath) {
		if err := os.MkdirAll(defIPAMPath, 0644); err != nil {
			return NewError("create folder %s failed, %v", defIPAMPath, err)
		}
	}

  return nil

}

// load ipam from config file
func (ipam *IPAM) load() error {

	if !PathExists(ipam.SubnetPath) {
		return NewError("ipam load: %s didn't exist", ipam.SubnetPath)
	}

	ipam.Mux.Lock()
	// read data from subnet.json
	subnetJson, err := ioutil.ReadFile(ipam.SubnetPath)
	if err != nil {
		return NewError("read data from subnet.json failed, err %v", err)
	} else {
		//log.Debugf("read from subnet.json: %s", string(subnetJson))
	}
	ipam.Mux.Unlock()

	// unmarshal subnet
	err = json.Unmarshal(subnetJson, ipam.Subnets)
	if err != nil {
		return NewError("load: Unmarshal failed, %v", err)
	}
	return nil
}

// store ipam
func (ipam *IPAM) Dump(iSubnet *map[string]string) error {

	// save data to subnet.json
	ipam.Mux.Lock()
	ipamFile, err := os.OpenFile(ipam.SubnetPath, os.O_CREATE| os.O_TRUNC| os.O_WRONLY, 0644)
	defer ipamFile.Close()
	if err != nil {
		return err
	}

	ipamJson, err := json.Marshal(iSubnet)
	if err != nil {
		return err
	} else {
		//log.Debugf("iSubnet: %v", iSubnet)
	}

	_, err = ipamFile.Write(ipamJson)
	ipam.Mux.Unlock()
	if err != nil {
		return err
	}

	return nil

}

func (ipam *IPAM) InitSubnet(subnetStr string) *net.IPNet {

	// translate subnet to IP.Net format
	_, subnet, _ := net.ParseCIDR(subnetStr)

	//log.Infof("subnetStr: %s, after translate subnet %v", subnetStr, subnet)

	// for instance if data is 127.0.0.0/8 then it will return 8, 24
	actualSize, totalSize := subnet.Mask.Size()

	if err := ipam.load(); err != nil {
		//log.Debugf("init subnet, load ipam config failed")
	}

	// store new subnet config if not in ipam.Subnets
	if _, exist := (*ipam.Subnets)[subnetStr]; !exist {
		// init subnet
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(totalSize-actualSize))  // 2^
		// store
		ipam.Dump(ipam.Subnets)
		return subnet
	} else {
		log.Debugf("not need to init ipam")
	}

	return nil

}

func (ipam *IPAM) Allocate(SubNet *net.IPNet) (ip net.IP, err error) {

	_, subnet, err := net.ParseCIDR(SubNet.String())
	if err != nil {
		return nil, err
	}

	// load subnets from config file
	err = ipam.load()
	if err != nil {
		log.Errorf("Allocate: ipam load failed, %v", err)
	} else {
		//log.Debugf("load ipam.Subnets: %v", ipam.Subnets)
	}

	for c := range (*ipam.Subnets)[subnet.String()] {
		// 0 mean available, get first element
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			// book this bit.
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			// first ip, for instance if data is 127.0.0.0/8 then it will return 127.0.0.0
			ip = subnet.IP
			// translate bytes(baseip + changed pos) to real ip
			for t := uint(4); t > 0; t -= 1 {
				// example:
				// t=0,  ip[0] = ip[0] + changePosition shift 3 * 2^3
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			// start from 1
			ip[3] += 1
			break
		}
	}

	ipam.Dump(ipam.Subnets)
	return
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {

	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		log.Errorf("ipam load failed, %v", err)
	}

	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1  // ip start from xx.xx.xx.1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)  // current ip - base ip -> bits string
	}

	// set bit to available
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	ipam.Dump(ipam.Subnets)
	return nil
}
