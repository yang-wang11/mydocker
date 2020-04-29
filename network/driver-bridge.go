package network

import (
	. "docker/mydocker/util"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"time"
)

type BridgeNetworkDriver struct{}

func NewBridgeNetworkDriver() *BridgeNetworkDriver {
	return &BridgeNetworkDriver{}
}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (d *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {

	ip, ipRange, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Errorf("ParseCIDR failed, subnet %s %v", subnet, err)
		return nil, err
	}
	ipRange.IP = ip
	log.Debugf("BridgeNetworkDriver Create ip %v, ipRange %v", ip, ipRange)

	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  d.Name(),
	}

	err = d.initBridge(n)

	return n, err

}

func (d *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}

func (d *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {

	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	// create veth device
	vethName := GetVethName(endpoint.ID)
	la := netlink.NewLinkAttrs()
	la.Name = vethName
	// set master attribute of veth interface,  and bind other side of veth to bridge network
	la.MasterIndex = br.Attrs().Index

	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + vethName,
	}

	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return NewError("add Endpoint Device failed, %v", err)
	}

	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return NewError("set Endpoint Device up failed, %v", err)
	}

	return nil

}

func (d *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	return nil
}

func (d *BridgeNetworkDriver) initBridge(n *Network) error {

	// try to get bridge by name, if it already exists then exit
	bridgeName := n.Name
	// ip link add xxxx
	if err := createBridgeInterface(bridgeName); err != nil {
		return NewError("add bridge %s failed, %v", bridgeName, err)
	} else {
		log.Debugf("receive network info %v", *n)
	}

	// set bridge IP(ip addr add xxx)
	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP
	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return NewError("assigning address %v on bridge %s failed, %v", gatewayIP, bridgeName, err)
	}

	// enable interface(ip link set xxx up)
	if err := setInterfaceUP(bridgeName); err != nil {
		return NewError("set bridge up %s failed, %v", bridgeName, err)
	}

	// Setup iptables(SNAT)
	log.Infof("start to set IPtables, bridgeName %s, IPRange %s \n", bridgeName, n.IpRange)
	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
		return NewError("set iptables for %s failed, %v", bridgeName, err)
	}

	return nil
}

func (d *BridgeNetworkDriver) deleteBridge(n *Network) error {
	bridgeName := n.Name

	// get the link
	l, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("Getting link with name %s failed: %v", bridgeName, err)
	}

	// delete the link
	if err := netlink.LinkDel(l); err != nil {
		return fmt.Errorf("Failed to remove bridge interface %s delete: %v", bridgeName, err)
	}

	return nil
}

// create bridge
func createBridgeInterface(bridgeName string) error {

	// return if bridge exist
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return NewError("bridge %s not found", bridgeName)
	}

	// create device(ip link add xxxx)
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return NewError("create bridge %s failed, %v", bridgeName, err)
	}
	return nil

}

// setup interface
func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("Error retrieving a link named [ %s ]: %v", iface.Attrs().Name, err)
	}

	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("Error enabling interface for %s: %v", interfaceName, err)
	}
	return nil
}

// set ip of interface
func setInterfaceIP(name string, rawIP string) error {

	var err error
	var iface netlink.Link

	// get device
	for i := 0; i < 3; i++ {
		iface, err = netlink.LinkByName(name)
		if err != nil {
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return NewError("Run [ ip link ] failed, %v", err)
	} else {
		log.Debugf("Run [ ip link ] successfully, ip %v", rawIP)
	}

	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return NewError("ParseIPNet %s failed, %v", rawIP, err)
	} else {
		log.Debugf("ParseIPNet %s successfully.", rawIP)
	}
	addr := &netlink.Addr{
		IPNet: ipNet,
		Label: "",
		Flags: 0,
		Scope: 0,
		Peer:  nil,
	}

	// ip addr add xxx
	return netlink.AddrAdd(iface, addr)

}

// set iptables
func setupIPTables(bridgeName string, subnet *net.IPNet) error {

	var err error

	//iptablesCmd := fmt.Sprintf("-t nat -I POSTROUTING -p tcp -d %s --dport %s -j SNAT --to 192.168.66.20", subnet.String(), bridgeName)
	iptablesCmd := fmt.Sprintf("-t nat -I POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)

	if ret := Cmder("iptables", true, nil, strings.Split(iptablesCmd, " ")...); !ret {
		err = fmt.Errorf("setup iptables failed")
	}

	return err

}

// due to length limit
func GetVethName(endpointId string) string {
	return endpointId[:5] ;
}
