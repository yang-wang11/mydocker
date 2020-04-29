package network

import (
	. "docker/mydocker/util"
)

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network Network, endpoint *Endpoint) error
}


func ValidNetworkDriver(driverType string) bool {

	driver := DriveType(driverType)

	switch driver {

		case BridgeDrive :

			return true

		default:

			return false

	}

}
