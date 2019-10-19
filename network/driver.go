package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
)

type Driver interface {
	Create(string, string) error
	Delete(string) error
	GetNetwork() *Network
	Connect(*Network, *Endpoint) error
	GetName() string
	load() error
	dump() error
}

func SetInterfaceUpByName(name string) error  {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("look link %s failed, error:%s", name, err.Error())
	}

	return SetInterfaceUp(link)
}

func SetInterfaceUp(link netlink.Link) error  {
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("up %s failed, error:%s", link.Attrs().Name, err.Error())
	}

	return nil
}

func ConfigInterfaceNetworkByName(name string, NW *Network) error  {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("look link %s failed, error:%s", name, err.Error())
	}

	return ConfigInterfaceNetwork(link, NW)
}

func ConfigInterfaceNetwork(link netlink.Link, NW *Network) error  {
	return ConfigInterfaceNetworkFromSubnetString(link, NW.IpRange.String())
}

func ConfigInterfaceNetworkFromSubnetString(link netlink.Link, subnet string) error  {
	addr, err := netlink.ParseAddr(subnet)
	if err != nil {
		return fmt.Errorf("parse %s failed; error:%s", NW.IpRange.String(), err.Error())
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("config %s to link %s failed, error:%s", addr.String(), NW.Name, err.Error())
	}

	return nil
}

// iptables -t nat -A POSTROUTING -o br0 -j MASQUERADE -s xxx.xxx.xx.xx
func ConfigMASQUERADE(NW *Network) error  {
	args := fmt.Sprintf("-t nat -A POSTROUTING -o %s -j MASQUERADE -s %s", NW.Name, NW.IpRange.String())
	cmd := exec.Command("iptables", strings.Split(args, " ")...)
	if output, err := cmd.Output(); err != nil {
		return fmt.Errorf("configMASQUERADE failed:%v; error:%s", output, err.Error())
	}

	return nil
}
