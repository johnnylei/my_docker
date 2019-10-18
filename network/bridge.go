package network

import (
	"encoding/json"
	"fmt"
	"github.com/johnnylei/my_docker/common"
	"github.com/urfave/cli"
	"github.com/vishvananda/netlink"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

type Bridge struct {
	Name string `json:"name"`
	nw *Network `json:"nw"`
	Loaded bool
}

func (bridge *Bridge) GetNetwork() *Network  {
	return bridge.nw
}

func (bridge *Bridge) dump() error  {
	if _, err := os.Stat(common.NETWORK_DRIVER_DIRECTORY); err != nil {
		if err := os.MkdirAll(common.NETWORK_DRIVER_DIRECTORY, 0644); err != nil {
			return fmt.Errorf("bridge dump failed, mkdir %s failed, error:%s", common.NETWORK_DRIVER_DIRECTORY, err.Error())
		}
	}

	dumpPath := path.Join(common.NETWORK_DRIVER_DIRECTORY, bridge.Name + ".json")
	if _, err := os.Stat(dumpPath); err != nil {
		if _, err := os.Create(dumpPath); err != nil {
			return fmt.Errorf("bridge dump failed, create %s failed, error:%s", dumpPath, err.Error())
		}
	}

	bridgeBytes, err := json.Marshal(bridge)
	if err != nil {
		return fmt.Errorf("bridge dump failed, json marshal failed, error:%s", err.Error())
	}

	if err := ioutil.WriteFile(dumpPath, bridgeBytes, syscall.O_CREAT | syscall.O_TRUNC | syscall.O_WRONLY); err != nil {
		return fmt.Errorf("bridge dump failed, write %s failed, error:%s", dumpPath, err.Error())
	}

	return nil
}

func (bridge *Bridge) load() error  {
	dumpPath := path.Join(common.NETWORK_DRIVER_DIRECTORY, bridge.Name + ".json")
	bridgeBytes, err := ioutil.ReadFile(dumpPath)
	if err != nil {
		return fmt.Errorf("laod %s failed, error:%s", dumpPath, err.Error())
	}

	if err := json.Unmarshal(bridgeBytes, bridge); err != nil {
		return fmt.Errorf("laod failed, json unmarshal failed, error:%s", err.Error())
	}

	bridge.nw = &Network{
		Loaded: false,
		Name: bridge.Name,
	}
	if err := bridge.nw.load();err != nil {
		return fmt.Errorf("bridge load:%s", err.Error())
	}

	bridge.Loaded = true
	return nil
}

func (bridge *Bridge) Delete(name string) error  {
	bridge.Name = name
	if !bridge.Loaded {
		if err := bridge.load(); err != nil {
			return fmt.Errorf("delete %s failed, load failed, error:%s", bridge.Name, err.Error())
		}
	}

	bridgeLink, err := netlink.LinkByName(bridge.Name)
	if err != nil {
		return fmt.Errorf("delete bridge %s failed,error:%s", bridge.Name, err.Error())
	}

	if err := netlink.LinkDel(bridgeLink); err != nil {
		return fmt.Errorf("delete bridge %s failed, error:%s", bridge.Name, err.Error())
	}

	// 删除iptables nat规则
	// iptables -t nat -D POSTROUTING -s 172.17.0.0/16 -o br0 -j MASQUERADE
	args := fmt.Sprintf("-t nat -D POSTROUTING -s %s -o %s -j MASQUERADE", bridge.nw.IpRange.String(), bridge.Name)
	if err := exec.Command("iptables", strings.Split(args, " ")...).Run(); err != nil {
		return fmt.Errorf("delete iptables failed, err:%s", err.Error())
	}

	// 删除路由
	// route del -net 172.17.0.0 netmask 255.255.0.0
	args = fmt.Sprintf("del -net %s netmask %s", GetSubnet(bridge.nw.IpRange), MaskToCIDRFormat(bridge.nw.IpRange.Mask))
	if err := exec.Command("route", strings.Split(args, " ")...).Run(); err != nil {
		return fmt.Errorf("delete route failed, route %s; err:%s", args, err.Error())
	}
	return nil
}

func (bridge *Bridge) Create(subnet string, name string) error  {
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("parse %s failed, error:%s", subnet, err.Error())
	}
	ipnet.IP = ip
	bridge.nw = &Network{
		Name: name,
		IpRange: ipnet,
		Driver: name,
		DriverType: "bridge",
	}
	bridge.Name = name
	if err := bridge.initBridgeInterface(); err != nil {
		return err
	}

	if err := bridge.dump(); err != nil {
		return err
	}

	return nil
}

/**
1. 创建网桥设备
2. 设置IP， route等信息
3. ip link set bridge up
4. 配置iptables SNAT
 */
func (bridge *Bridge) initBridgeInterface() error {
	if err := bridge.createBridgetInterface(); err != nil {
		return err
	}

	if err := bridge.configBridgeInterface(); err != nil {
		return err
	}

	if err := bridge.upBridgeInterface(); err != nil {
		return err
	}

	if err := bridge.configMASQUERADE(); err != nil {
		return err
	}

	return nil
}

func (bridge *Bridge) createBridgetInterface() error  {
	device, _ := net.InterfaceByName(bridge.nw.Name)
	if device != nil {
		return fmt.Errorf("interface %s exist", bridge.nw.Name)
	}

	la := netlink.NewLinkAttrs()
	la.Name = bridge.nw.Name
	if err := netlink.LinkAdd(&netlink.Bridge{LinkAttrs: la}); err != nil {
		return fmt.Errorf("add bridge %s failed, error:%s", la.Name, err.Error())
	}

	return nil
}

func (bridge *Bridge) configBridgeInterface() error  {
	bridgeLink, err := netlink.LinkByName(bridge.nw.Name)
	if err != nil {
		return fmt.Errorf("link %s failed, error: %s", bridge.nw.Name, err.Error())
	}

	addr, err := netlink.ParseAddr(bridge.nw.IpRange.String())
	if err != nil {
		return fmt.Errorf("parse %s failed; error:%s", bridge.nw.IpRange.String(), err.Error())
	}

	if err := netlink.AddrAdd(bridgeLink, addr); err != nil {
		return fmt.Errorf("config %s to link %s failed, error:%s", addr.String(), bridge.nw.Name, err.Error())
	}

	return nil
}

func (bridge *Bridge) upBridgeInterface() error  {
	bridgeLink, err := netlink.LinkByName(bridge.nw.Name)
	if err != nil {
		return fmt.Errorf("link %s failed, error: %s", bridge.nw.Name, err.Error())
	}
	if err := netlink.LinkSetUp(bridgeLink); err != nil {
		return fmt.Errorf("link %s set up failed, error:%s", bridge.nw.Name, err.Error())
	}

	return nil
}

// iptables -t nat -A POSTROUTING -o br0 -j MASQUERADE -s xxx.xxx.xx.xx
func (bridge *Bridge) configMASQUERADE() error  {
	args := fmt.Sprintf("-t nat -A POSTROUTING -o %s -j MASQUERADE -s %s", bridge.nw.Name, bridge.nw.IpRange.String())
	cmd := exec.Command("iptables", strings.Split(args, " ")...)
	if output, err := cmd.Output(); err != nil {
		return fmt.Errorf("configMASQUERADE failed:%v; error:%s", output, err.Error())
	}

	return nil
}

func CreateBridgeInterface(context *cli.Context) error {
	bridge := &Bridge{
		Loaded:false,
	}
	if err := bridge.Create(context.String("subnet"), context.String("name")); err != nil {
		return err
	}

	return nil
}

func DeleteBridgeInterface(context *cli.Context) error  {
	bridge := &Bridge{
		Loaded:false,
	}
	if err := bridge.Delete(context.String("name")); err != nil {
		return err
	}

	return nil
}

func MaskToCIDRFormat(mask net.IPMask) string {
	ones, _ := mask.Size()
	cidr := [4]int{0, 0, 0, 0}

	if ones <= 8 {
		for i := 0; i < ones; i++ {
			cidr[0] += 1 << uint8(ones - i)
		}
	} else if ones <= 16 {
		cidr[0] = 255
		ones -= 8
		for i := 0; i < ones; i++ {
			cidr[1] += 1 << uint8(7 - i)
		}
	} else if ones <= 24 {
		cidr[0] = 255
		cidr[1] = 255
		ones -= 16
		for i := 0; i < ones; i++ {
			cidr[2] += 1 << uint8(7 - i)
		}
	} else if ones <= 24 {
		cidr[0] = 255
		cidr[1] = 255
		cidr[2] = 255
		ones -= 24
		for i := 0; i < ones; i++ {
			cidr[3] += 1 << uint8(7 - i)
		}
	}

	return fmt.Sprintf("%d.%d.%d.%d", cidr[0], cidr[1], cidr[2], cidr[3])
}

func GetSubnet(ipnet *net.IPNet) string  {
	_, subnet, _ := net.ParseCIDR(ipnet.String())
	return subnet.IP.String()
}