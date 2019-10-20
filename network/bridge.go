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
	Loaded bool `json:"loaded"`
	NW *Network	`json:"nw"`
}

func (bridge *Bridge) GetName() string  {
	return bridge.Name
}

func (bridge *Bridge) GetNetwork() *Network  {
	return bridge.NW
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

	//bridge.NW = &Network{
	//	Loaded: false,
	//	Name: bridge.Name,
	//}
	//if err := bridge.NW.load();err != nil {
	//	return fmt.Errorf("bridge load:%s", err.Error())
	//}

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
	args := fmt.Sprintf("-t nat -D POSTROUTING -s %s -o %s -j MASQUERADE", bridge.NW.IpRange.String(), bridge.Name)
	if err := exec.Command("iptables", strings.Split(args, " ")...).Run(); err != nil {
		return fmt.Errorf("delete iptables failed, err:%s", err.Error())
	}

	// 删除路由
	// route del -net 172.17.0.0 netmask 255.255.0.0
	args = fmt.Sprintf("del -net %s netmask %s", GetSubnet(bridge.NW.IpRange), MaskToCIDRFormat(bridge.NW.IpRange.Mask))
	if err := exec.Command("route", strings.Split(args, " ")...).Run(); err != nil {
		return fmt.Errorf("delete route failed, route %s; err:%s", args, err.Error())
	}
	return nil
}

func (bridge *Bridge) Connect(NW *Network, endpoint *Endpoint) error  {
	bridgeLink, err := netlink.LinkByName(NW.Name)
	if err != nil {
		return fmt.Errorf("connect nw enpoint failed, error:%s", err.Error())
	}

	la := netlink.NewLinkAttrs()
	la.MasterIndex = bridgeLink.Attrs().Index
	la.Name = endpoint.ID[:5]
	endpoint.Device = netlink.Veth{
		LinkAttrs:la,
		PeerName: "cif-"+endpoint.ID[:5],
	}
	if err := netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("create veth %s failed, error:%s", endpoint.Device.Name, err.Error())
	}

	if err := netlink.LinkSetUp(&endpoint.Device); err != nil {
		netlink.LinkDel(&endpoint.Device)
		return fmt.Errorf("link %s set up failed, error:%s", endpoint.Device.Name, err.Error())
	}

	return nil
}

func (bridge *Bridge) Create(subnet string, name string) error  {
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("parse %s failed, error:%s", subnet, err.Error())
	}
	ipnet.IP = ip
	bridge.NW = &Network{
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

	if err := ConfigInterfaceNetworkByName(bridge.Name, bridge.NW); err != nil {
		return err
	}

	if err := SetInterfaceUpByName(bridge.Name); err != nil {
		return err
	}

	if err := ConfigMASQUERADE(bridge.NW); err != nil {
		return err
	}

	return nil
}

func (bridge *Bridge) createBridgetInterface() error  {
	device, _ := net.InterfaceByName(bridge.NW.Name)
	if device != nil {
		return fmt.Errorf("interface %s exist", bridge.NW.Name)
	}

	la := netlink.NewLinkAttrs()
	la.Name = bridge.NW.Name
	if err := netlink.LinkAdd(&netlink.Bridge{LinkAttrs: la}); err != nil {
		return fmt.Errorf("add bridge %s failed, error:%s", la.Name, err.Error())
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