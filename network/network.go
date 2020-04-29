package network

import (
	. "docker/mydocker/util"
	"fmt"
	"net"
	"os/exec"

	"encoding/json"
	"os"
	"path"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	log "github.com/Sirupsen/logrus"
)

type Network struct {
	Name    string
	IpRange *net.IPNet
	Driver  string
}

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	Network     *Network
	PortMapping []string
}

func (nw *Network) dump(dumpPath string) error {

	if !PathExists(dumpPath) {
		if err := os.MkdirAll(dumpPath, 0644); err != nil {
			return NewError("Network dump: create folder %s failed, %v", dumpPath, err)
		}
	}

	// get persistent filename of nw
	nwPath := path.Join(dumpPath, nw.Name)

	// read network setting
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return NewError("Network dump: read file %s failed, %v", nwPath, err)
	}
	defer nwFile.Close()

	// marshal content
	nwJson, err := json.Marshal(nw)
	if err != nil {
		return NewError("Network dump: marshal %s failed, %v", nw.Name, err)
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		return NewError("Network dump: save content to file %s failed, %v", nwPath, err)
	}
	return nil

}

func (nw *Network) remove(dumpPath string) error {

	if !PathExists(path.Join(dumpPath, nw.Name)) {
		return nil
	} else {
		return os.RemoveAll(path.Join(dumpPath, nw.Name))
	}

}

func (nw *Network) load(dumpPath string) error {

	nwConf, err := os.Open(dumpPath)
	if err != nil {
		return NewError("Network load: open file %s failed, %v", dumpPath, err)
	}
	defer nwConf.Close()

	// read network setting
	nwJson := make([]byte, 2000)
	fLen, err := nwConf.Read(nwJson)
	if err != nil {
		return NewError("Network load: read file %s failed, %v", dumpPath, err)
	}

	err = json.Unmarshal(nwJson[:fLen], nw)
	if err != nil {
		return NewError("Network load: unmarshal network failed, %v", err)
	}

	return nil

}

func CreateNetwork(driverType, subnet, name string) error {

	var cidrStr string

	// translate from subnet to format: cidr
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	} else {
		cidrStr = cidr.String()
		log.Debugf("cidr set to %s", cidrStr)
	}

	IpAllocator.InitSubnet(cidrStr)
	// assign ip for gateway
	gatewayip, err := IpAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayip

	// use "network driver" to create network
	drive, ok := Drivers[driverType]
	if !ok {
		return NewError("drive type %s not found", driverType)
	}

	log.Debugf("cidr.String(): ", cidr.String())

	nw, err := drive.Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return nw.dump(defNetworkPath)

}

func ListNetwork() {

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)

	fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, nw := range Networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver)
	}
	if err := w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
		return
	}
}

func DeleteNetwork(networkName string) error {

	nw, ok := Networks[networkName]
	if !ok {
		return NewError("No Such Network: %s", networkName)
	}

	// release ip
	if err := IpAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return NewError("Error Remove Network gateway ip: %s", err)
	}

	// update ipam
	IpAllocator.load()
	_, cidr, _ := net.ParseCIDR(nw.IpRange.String())
	cidrStr := cidr.String()
	delete(*IpAllocator.Subnets, cidrStr)
	IpAllocator.Dump(IpAllocator.Subnets)

	// delete iptables setting
	iptablesCmd := fmt.Sprintf("-t nat -D POSTROUTING -s %s ! -o %s -j MASQUERADE", cidrStr, networkName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	_, err := cmd.Output()
	if err != nil {
		log.Errorf("clean iptables of %s failed, %v", networkName, err)
	}

	// delete network from device
	driver, ok := Drivers[nw.Driver]
	if !ok {
		return NewError("drive type %s not found", nw.Driver)
	}
	if err := driver.Delete(*nw); err != nil {
		return NewError("Error Remove Network Driver failed, %s", err)
	}

	// delete network setting
	return nw.remove(defNetworkPath)

}

func enterContainerNetns(enLink *netlink.Link, con *Container) func() {

	// get network-namespace info
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", con.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("get container net namespace failed, %v, Pid: '%s'", err, con.Pid)
	}

	// get file descriptor for network-namespace, lock it!!!
	netFd := f.Fd()
	runtime.LockOSThread()

	// puts veth peer into container's network namespace ( ip link set $link netns $ns )
	if err = netlink.LinkSetNsFd(*enLink, int(netFd)); err != nil {
		log.Errorf("set link netns failed, %v", err)
	}

	// get current process's network namespace
	origns, err := netns.Get()
	if err != nil {
		log.Errorf("get current netns failed, %v", err)
	}

	// set current process's to new network namespace(container)
	if err = netns.Set(netns.NsHandle(netFd)); err != nil {
		log.Errorf("set netns failed, %v", err)
	}

	// recover to original network namespace
	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}

}

func configEndpointIpAddressAndRoute(ep *Endpoint, con *Container) error {

	// get other side of veth
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return NewError("config endpoint failed, %v", err)
	}

	// when enterContainerNetns executed, the netns become from host to container
	// after defer function execute, netns recover to original
	defer enterContainerNetns(&peerLink, con)()
	// following code all execute in container's netns

	// net.IPNet
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress
	log.Debugf("interfaceIP %v, %v", interfaceIP.String(), ep.Network.IpRange.String())

	// set ip for veth peer
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return NewError("setInterfaceIP failed, %v,%s", ep.Network, err)
	}

	// set veth peer up
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return NewError("setInterfaceUP failed, %s", err)
	}

	// set lo up
	if err = setInterfaceUP("lo"); err != nil {
		return NewError("setInterfaceUP failed, %s", err)
	}

	// add route: all network through veth peer
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		log.Errorf("route add %v failed, %v", defaultRoute, err)
	}

	return nil
}

func configPortMapping(ep *Endpoint) error {

	var containerPort, hostPort string

	for _, pm := range ep.PortMapping {

		// valid port mapping
		portMapArray := strings.Split(pm, ":")
		if len(portMapArray) != 2 {
			log.Errorf("port mapping format failed, %v", pm)
			continue
		} else {
			hostPort, containerPort = portMapArray[0], portMapArray[1]
		}

		// get host ip
		//hostIp := ""
		//addrs, err := net.InterfaceAddrs()
		//if err != nil {
		//	return  NewError("get host ip failed")
		//}
		//for _, address := range addrs {
		//
		//	// 检查ip地址判断是否回环地址
		//	if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
		//		if ipnet.IP.To4() != nil {
		//			hostIp = ipnet.IP.String()
		//		}
		//
		//	}
		//}

		// execute DNAT
		iptableDNAT := fmt.Sprintf("-t nat -I PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s",
			hostPort, ep.IPAddress.String(), containerPort)
		if !Cmder("iptables", true, nil, strings.Split(iptableDNAT, " ")...) {
			return NewError("invoke iptables DNAT failed")
		}

	}
	return nil
}

func Connect(con *Container) error {

	// get network setting
	netName := con.Network

	network, ok := Networks[netName]
	if !ok {
		return NewError("No Such Network %s", netName)
	} else {
		log.Debugf("connect to network %v", network)
	}

	// assign IP for Endpoint
	ip, err := IpAllocator.Allocate(network.IpRange)
	if err != nil {
		return NewError("IpAllocator.Allocate failed, %v", err)
	} else {
		con.IP = ip.String()
		log.Debugf("get ip %s for ep, ip range %v, container ip %v", ip, network.IpRange, con.IP)
	}

	// create endpoint (container - network)
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", con.Id, netName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: con.PortMapping,
	}

	con.NetworkDevice  = GetVethName(ep.ID)

	// through driver connect&config to endpoint
	driver, ok := Drivers[network.Driver]
	if !ok {
		return NewError("drive type %s not found", network.Driver)
	}
	// connect one side of ep to network (in host)
	if err = driver.Connect(network, ep); err != nil {
		return NewError("driver.Connect failed, %v", err)
	}

	// connect other side of ep to container
	// configure container's network IP&route with self network namespace
	if err = configEndpointIpAddressAndRoute(ep, con); err != nil {
		return NewError("configEndpointIpAddressAndRoute failed, %v", err)
	}

	// config port mapping between container and host
	return configPortMapping(ep)

}

func Disconnect(con *Container) error {
	return nil
}
