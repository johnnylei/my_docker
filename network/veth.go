package network

import (
	"fmt"
	"github.com/johnnylei/my_docker/common"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"log"
	"net"
	"os"
	"runtime"
)

func ConfigVethNetWork(endpoint *Endpoint, containerInfo *common.ContainerInformation) error  {
	peerlink, err := netlink.LinkByName(endpoint.Device.PeerName)
	if err != nil {
		return fmt.Errorf("configVethNetWork failed, link %s not found, error:%s", endpoint.Device.PeerName, err.Error())
	}

	// defer先调用外层函数,然后往后执行,最后调用里层函数
	// defer需要先把外层函数解析出来变成一个函数名
	defer func(peerlink *netlink.Link, containerInfo *common.ContainerInformation) func() {
		netnsfile, err := os.OpenFile(fmt.Sprintf("/proc/%d/ns/net", containerInfo.Pid), os.O_RDONLY, 0)
		if err != nil {
			log.Fatal(fmt.Sprintf("open /proc/%d/ns/net faild", containerInfo.Pid))
		}

		netnsFd := netnsfile.Fd()
		runtime.LockOSThread()

		if err := netlink.LinkSetNsFd(*peerlink, int(netnsFd)); err != nil {
			log.Fatal(fmt.Sprintf("error set link netns, %v", err))
		}

		originNs, err := netns.Get()
		if err != nil {
			log.Fatal(fmt.Sprintf("get current net namespace failed, %v", err))
		}

		// 把当前进程加入到net ns中
		if err := netns.Set(netns.NsHandle(netnsFd)); err != nil {
			log.Fatal(fmt.Sprintf("error set ns; %s", err.Error()))
		}

		return func() {
			netns.Set(originNs)
			originNs.Close()
			runtime.UnlockOSThread()
			netnsfile.Close()
		}
	}(&peerlink, containerInfo)()

	interfaceIp := *endpoint.NW.IpRange
	interfaceIp.IP = endpoint.IPAddress
	fmt.Printf("%s\n", interfaceIp.String())
	if err := ConfigInterfaceNetworkFromSubnetString(peerlink, interfaceIp.String()); err != nil {
		return fmt.Errorf("configVethNetWork failed, link %s config network failed, error:%s",
			endpoint.Device.PeerName, err.Error())
	}

	if err := SetInterfaceUp(peerlink); err != nil {
		return fmt.Errorf("configVethNetWork failed, link %s up failed, error:%s",
			endpoint.Device.PeerName, err.Error())
	}

	if err := SetInterfaceUpByName("lo"); err != nil {
		return fmt.Errorf("configVethNetWork failed, link lo up failed, error:%s", err.Error())
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerlink.Attrs().Index,
		Gw: endpoint.NW.IpRange.IP,
		Dst: cidr,
	}

	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("add default route failed, error:%s", err.Error())
	}

	return nil
}