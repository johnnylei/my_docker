package network

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
)

type Bridge struct {}

func (bridge *Bridge) Name() string {
	return "bridge"
}

func (bridge *Bridge) Delete(name string) error  {
	bridgeLink, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("delete bridge %s failed,error:%s", name, err.Error())
	}

	if err := netlink.LinkDel(bridgeLink); err != nil {
		return fmt.Errorf("delete bridge %s failed, error:%s", name, err.Error())
	}

	return nil
}

func (bridge *Bridge) Create(subnet string, name string) (*Network, error)  {
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("parse %s failed, error:%s", subnet, err.Error())
	}
	ipnet.IP = ip
	createdNet := &Network{
		Name: name,
		IpRange: ipnet,
	}

	if err := bridge.initBridgeInterface(createdNet); err != nil {
		return nil, err
	}

	return createdNet, nil
}

/**
1. 创建网桥设备
2. 设置IP， route等信息
3. ip link set bridge up
4. 配置iptables SNAT
 */
func (bridge *Bridge) initBridgeInterface(network *Network) error {
	if err := bridge.createBridgetInterface(network); err != nil {
		return err
	}

	if err := bridge.configBridgeInterface(network); err != nil {
		return err
	}

	if err := bridge.upBridgeInterface(network); err != nil {
		return err
	}

	if err := bridge.configMASQUERADE(network); err != nil {
		return err
	}

	return nil
}

func (bridge *Bridge) createBridgetInterface(network *Network) error  {
	device, _ := net.InterfaceByName(network.Name)
	if device != nil {
		return fmt.Errorf("interface %s exist", network.Name)
	}

	la := netlink.NewLinkAttrs()
	la.Name = network.Name
	if err := netlink.LinkAdd(&netlink.Bridge{LinkAttrs: la}); err != nil {
		return fmt.Errorf("add bridge %s failed, error:%s", la.Name, err.Error())
	}

	return nil
}

func (bridge *Bridge) configBridgeInterface(network *Network) error  {
	bridgeLink, err := netlink.LinkByName(network.Name)
	if err != nil {
		return fmt.Errorf("link %s failed, error: %s", network.Name, err.Error())
	}

	addr, err := netlink.ParseAddr(network.IpRange.String())
	if err != nil {
		return fmt.Errorf("parse %s failed; error:%s", network.IpRange.String(), err.Error())
	}

	if err := netlink.AddrAdd(bridgeLink, addr); err != nil {
		return fmt.Errorf("config %s to link %s failed, error:%s", addr.String(), network.Name, err.Error())
	}

	return nil
}

func (bridge *Bridge) upBridgeInterface(network *Network) error  {
	bridgeLink, err := netlink.LinkByName(network.Name)
	if err != nil {
		return fmt.Errorf("link %s failed, error: %s", network.Name, err.Error())
	}
	if err := netlink.LinkSetUp(bridgeLink); err != nil {
		return fmt.Errorf("link %s set up failed, error:%s", network.Name, err.Error())
	}

	return nil
}

// iptables -t nat -A POSTROUTING -o br0 -j MASQUERADE -s xxx.xxx.xx.xx
func (bridge *Bridge) configMASQUERADE(network *Network) error  {
	args := fmt.Sprintf("-t nat -A POSTROUTING -o %s -j MASQUERADE -s %s", network.Name, network.IpRange.String())
	cmd := exec.Command("iptables", strings.Split(args, " ")...)
	if output, err := cmd.Output(); err != nil {
		return fmt.Errorf("configMASQUERADE failed:%v; error:%s", output, err.Error())
	}

	return nil
}

func CreateBridgeInterface(context *cli.Context) error {
	bridge := &Bridge{}
	if _, err := bridge.Create(context.String("subnet"), context.String("name")); err != nil {
		return err
	}

	return nil
}

func DeleteBridgeInterface(context *cli.Context) error  {
	bridge := &Bridge{}
	if err := bridge.Delete(context.String("name")); err != nil {
		return err
	}

	return nil
}