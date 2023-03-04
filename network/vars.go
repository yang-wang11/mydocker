package network

import (
	"github.com/yang-wang11/mydocker/common"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	defIPAMPath    = common.RootDir + "/network/ipam/"
	defIPAMConf    = defIPAMPath + "subnet.json"
	defNetworkPath = common.RootDir + "/network/network/"
)

var (
	Drivers     = map[string]NetworkDriver{}
	Networks    = map[string]*Network{}
	IpAllocator *IPAM
)

// load network drivers & network
func Initbridge() {

	log.Debugf("start to init network.")

	// load supported drives to global drive
	bridgeDrive := NewBridgeNetworkDriver() // bridge
	Drivers[bridgeDrive.Name()] = bridgeDrive

	// make sure default network exist
	if !common.PathExists(defNetworkPath) {
		if err := os.MkdirAll(defNetworkPath, os.ModeDir); err != nil {
			log.Errorf("create folder %s failed, %v", defNetworkPath, err)
		}
	}

	// init ipam module
	IpAllocator = NewIPAM()
	if !common.PathExists(IpAllocator.SubnetPath) {
		IpAllocator.init()
	}

	// load exist networks to networks(only name)
	reloadNetwork := func() {

		filepath.Walk(defNetworkPath, func(nwPath string, info os.FileInfo, err error) error {

			// skip folder
			if info.IsDir() {
				return nil
			}

			// get file name
			_, nwName := path.Split(nwPath)
			nw := &Network{
				Name: nwName,
			}

			// load all network setting
			if err := nw.load(nwPath); err != nil {
				log.Errorf("error load network: %s", err)
			}
			Networks[nwName] = nw

			return nil
		})

	}

	reloadNetwork()

	// if default bridge network is not exist, generate it
	if _, ok := Networks[common.DefNetworkName]; !ok {
		err := CreateNetwork(string(common.DefDriveType), common.DefNetworkSubnet, common.DefNetworkName)
		if err != nil {
			log.Errorf("init default network %s, %+v", common.DefNetworkName, err)
		}
		reloadNetwork()
	}

	log.Debugf("init default network %s successfully. networks: %v", common.DefNetworkName, Networks)

}
