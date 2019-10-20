package network

import (
	"encoding/json"
	"fmt"
	"github.com/johnnylei/my_docker/common"
	"github.com/urfave/cli"
	"github.com/vishvananda/netlink"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

var (
	drivers = map[string]Driver {
		"bridge": &Bridge{
			Loaded: false,
		},
	}
)

type Network struct {
	Name string `json:"name"`
	IpRange *net.IPNet `json:"ip_range"`
	Driver string `json:"driver"`
	DriverType string `json:"driver_type"`
	Loaded bool
}


func DeleteNetwork(context *cli.Context) error  {
	nw := &Network{
		Name: context.String("name"),
		Loaded: false,
	}
	return nw.Delete()
}

func CreateNetwork(context *cli.Context) error  {
	SubnetString := context.String("subnet")
	_, ipnet, err := net.ParseCIDR(SubnetString)
	if err != nil {
		return fmt.Errorf("parse network %s failed, error:%s", SubnetString, err.Error())
	}

	allocated, err := ipam.CheckSubnetAllocated(ipnet)
	if err != nil {
		return err
	}

	if allocated {
		return fmt.Errorf("subnet %s has been allocated", SubnetString)
	}

	gateway, err := ipam.Allocate(ipnet)
	if err != nil {
		return err
	}
	ipnet.IP = gateway

	networkName := context.String("name")
	driver := drivers[context.String("driver-type")]
	err = driver.Create(ipnet.String(), networkName)
	if err != nil {
		return err
	}

	return driver.GetNetwork().dump()
}

func (nw *Network) dump() error  {
	if _, err := os.Stat(common.NETWORK_INFORMATION_DIRECTORY); err != nil {
		if err := os.MkdirAll(common.NETWORK_INFORMATION_DIRECTORY, 0644); err != nil {
			return fmt.Errorf("mkdir %s failed, error:%s", common.NETWORK_INFORMATION_DIRECTORY, err.Error())
		}
	}

	dumpPath := path.Join(common.NETWORK_INFORMATION_DIRECTORY, nw.Name + ".json")
	if _, err := os.Stat(dumpPath); err != nil {
		if _, err := os.Create(dumpPath); err != nil {
			return fmt.Errorf("create %s failed, error %s", dumpPath, err.Error())
		}
	}

	content, err := json.Marshal(nw)
	if err != nil {
		return fmt.Errorf("json Marshal failed, error:%s", err.Error())
	}

	if err := ioutil.WriteFile(dumpPath, content, syscall.O_WRONLY | syscall.O_TRUNC | syscall.O_CREAT); err != nil {
		return fmt.Errorf("write file %s failed, error:%s", dumpPath, err.Error())
	}

	return nil
}

func (nw *Network) load() error  {
	dumpPath := path.Join(common.NETWORK_INFORMATION_DIRECTORY, nw.Name + ".json")
	contentBytes, err := ioutil.ReadFile(dumpPath)
	if err != nil {
		return fmt.Errorf("read %s failed, error:%s", dumpPath, err.Error())
	}

	if err := json.Unmarshal(contentBytes, nw); err != nil {
		return fmt.Errorf("json Unmarshal failed, error:%s", err.Error())
	}

	return nil
}

func (nw *Network) Delete() error  {
	if !nw.Loaded {
		if err := nw.load(); err != nil {
			return fmt.Errorf("DeleteNetwork failed, %s", err.Error())
		}
	}

	_, subnet, _ := net.ParseCIDR(nw.IpRange.String())
	if err := ipam.DropSubnet(subnet); err != nil {
		return fmt.Errorf("DeleteNetwork failed, %s", err.Error())
	}

	driver := drivers[nw.DriverType]
	if err := driver.Delete(nw.Driver); err != nil {
		return fmt.Errorf("DeleteNetwork failed, %s", err.Error())
	}

	dumpPath := path.Join(common.NETWORK_INFORMATION_DIRECTORY, nw.Name + ".json")
	if err := os.Remove(dumpPath); err != nil {
		return fmt.Errorf("delete %s failed, error:%s", dumpPath, err.Error())
	}

	return nil
}

func Connect(context *cli.Context, containerInfo *common.ContainerInformation) error  {
	nw := &Network{
		Name: context.String("net"),
		Loaded: false,
	}

	if err := nw.load(); err != nil {
		return fmt.Errorf("connect failed, %s", err.Error())
	}

	ip, err := ipam.Allocate(nw.IpRange)
	if err != nil {
		return fmt.Errorf("connect failed, %s", err.Error())
	}

	endpoint := &Endpoint{
		ID: fmt.Sprintf("%s-%s", containerInfo.Id, nw.Name),
		IPAddress: ip.IP,
		NW: nw,
		PortMapping: containerInfo.PortMapping,
	}
	driver := drivers[nw.DriverType]
	if err := driver.Connect(nw, endpoint); err != nil {
		ipam.DropSubnet(nw.IpRange)
		return fmt.Errorf("driver %s connect nw to endpoint failed, error:%s", driver.GetName(), err.Error())
	}

	if err := ConfigVethNetWork(endpoint, containerInfo); err != nil {
		netlink.LinkDel(&endpoint.Device)
		ipam.DropSubnet(nw.IpRange)
		return fmt.Errorf("configVethNetWork failed, error:%s", err.Error())
	}

	ConfigPortMapping(endpoint)

	return nil
}

func ConfigPortMapping(endpoint *Endpoint) {
	for _, portMappingItem := range endpoint.PortMapping {
		portMapping := strings.Split(portMappingItem, ":")
		if len(portMapping) != 2 {
			log.Printf("port mapping usage: sourcePort:DestinationPort")
			continue
		}

		// iptables -t nat -A POSTROUTING -i eth0 -j DNAT -p tcp --dport 80 --to-destionation xxx.xx.xxx.xx:xxx-xxx.xxx.xxx.xxx:xxx
		args := fmt.Sprintf("-t nat -A PREROUTING -i %s -j DNAT -p tcp --dport %s --to-destination %s:%s",
			endpoint.NW.Name, portMapping[0], endpoint.IPAddress, portMapping[1])
		if output, err := exec.Command("iptables", strings.Split(args, " ")...).Output(); err != nil {
			log.Printf("DNAT failed, iptables %s, output:%s, error:%s", args, string(output), err.Error())
			continue
		}
	}
}

