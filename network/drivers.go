package network

import "github.com/yang-wang11/mydocker/common"

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network Network, endpoint *Endpoint) error
}

func ValidNetworkDriver(driverType string) bool {

	driver := common.DriveType(driverType)

	switch driver {

	case common.BridgeDrive:
		return true

	default:
		return false
	}
}
