package network

import (
	"encoding/json"
	"fmt"
	"github.com/johnnylei/my_docker/common"
	"github.com/urfave/cli"
	"io/ioutil"
	"net"
	"os"
	"path"
	"syscall"
)

var (
	drivers = map[string]*Bridge {}
)

func Init()  {
	bridge := &Bridge{}
	drivers = map[string]*Bridge {
		bridge.Name(): bridge,
	}
}

type Network struct {
	Name string `json:"name"`
	IpRange *net.IPNet `json:"ip_range"`
	Driver string `json:"driver"`
}

func CreateNetwork(context *cli.Context) error  {
	SubnetString := context.String("subnet")
	_, ipnet, err := net.ParseCIDR(SubnetString)
	if err != nil {
		return fmt.Errorf("parse network %s failed, error:%s", SubnetString, err.Error())
	}

	ipam := &IPAM{
		SubnetAllocatedPath:common.IPAM_ALLOCAT_SUBNET_DUMP_PATH,
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

	networkName := context.String("name")
	driverName := context.String("driver")
	nw, err := drivers[driverName].Create(gateway.String(), networkName)
	if err != nil {
		return err
	}

	return nw.dump()
}

func (nw *Network) dump() error  {
	if _, err := os.Stat(common.NETWORK_INFORMATION_DIRECTORY); err != nil {
		if err := os.MkdirAll(common.NETWORK_INFORMATION_DIRECTORY, 0644); err != nil {
			return fmt.Errorf("mkdir %s failed, error:%s", common.NETWORK_INFORMATION_DIRECTORY, err.Error())
		}
	}

	dumpPath := path.Join(common.NETWORK_INFORMATION_DIRECTORY, nw.Name, ".json")
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
	dumpPath := path.Join(common.NETWORK_INFORMATION_DIRECTORY, nw.Name, ".json")
	contentBytes, err := ioutil.ReadFile(dumpPath)
	if err != nil {
		return fmt.Errorf("read %s failed, error:%s", dumpPath, err.Error())
	}

	if err := json.Unmarshal(contentBytes, nw); err != nil {
		return fmt.Errorf("json Unmarshal failed, error:%s", err.Error())
	}

	return nil
}