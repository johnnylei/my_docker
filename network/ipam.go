package network

import (
	"encoding/json"
	"fmt"
	"github.com/johnnylei/my_docker/common"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"syscall"
)

var (
	ipam = &IPAM{
		SubnetAllocatedPath:common.IPAM_ALLOCAT_SUBNET_DUMP_PATH,
		Loaded:false,
		Subnets:&map[string]string{},
	}
)

type IPAM struct {
	SubnetAllocatedPath string
	Subnets *map[string]string
	Loaded bool
}

func (ipam *IPAM) dump() error  {
	SubnetsJsonBytes, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return fmt.Errorf("json marshal subnets failed")
	}

	subnetAllocatedPathDir, _ := path.Split(ipam.SubnetAllocatedPath)
	_, err = os.Stat(subnetAllocatedPathDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("get %s stat faield, error:%s", ipam.SubnetAllocatedPath, err.Error())
		} else {
			if err := os.MkdirAll(subnetAllocatedPathDir, 0644); err != nil {
				return fmt.Errorf("mk file %s failed, error:%s", subnetAllocatedPathDir, err.Error())
			}
		}
	}

	if _, err = os.Stat(ipam.SubnetAllocatedPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("open %s failed, error:%s", ipam.SubnetAllocatedPath, err.Error())
	} else if err != nil {
		_, err := os.Create(ipam.SubnetAllocatedPath)
		if err != nil {
			return fmt.Errorf("create %s failed, error:%s", ipam.SubnetAllocatedPath, err.Error())
		}
	}

	err = ioutil.WriteFile(ipam.SubnetAllocatedPath, SubnetsJsonBytes, syscall.O_WRONLY | syscall.O_TRUNC | syscall.O_CREAT)
	if err != nil {
		return fmt.Errorf("dump subnets failed, write %s failed, error:%s", ipam.SubnetAllocatedPath, err.Error())
	}

	return nil
}

func (ipam *IPAM) Load() error  {
	SubnetsJsonBytes, err := ioutil.ReadFile(ipam.SubnetAllocatedPath)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(SubnetsJsonBytes, ipam.Subnets)
	if err != nil {
		return fmt.Errorf("load failed, json Unmarshal %s failed, error:%s", string(SubnetsJsonBytes), err.Error())
	}

	ipam.Loaded = true
	return nil
}

func (ipam *IPAM) CheckSubnetAllocated(subnet *net.IPNet) (bool, error)  {
	if !ipam.Loaded {
		if err := ipam.Load(); err != nil {
			return false, fmt.Errorf("CheckSubnetAllocated failed, error:%s", err.Error())
		}
	}

	if _, exist := (*ipam.Subnets)[subnet.String()]; exist {
		return true, nil
	}

	return false, nil
}

func (ipam *IPAM) DropSubnet(subnet *net.IPNet) error {
	if !ipam.Loaded {
		if err := ipam.Load(); err != nil {
			return fmt.Errorf("DropSubnet failed, error:%s", err.Error())
		}
	}

	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		return nil
	}

	delete((*ipam.Subnets), subnet.String())
	if err := ipam.dump(); err != nil {
		return fmt.Errorf("DropSubnet failed, message:%s", err.Error())
	}

	return nil
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (*net.IPNet, error)  {
	if !ipam.Loaded {
		if err := ipam.Load(); err != nil {
			return nil, err
		}
	}

	ip := []byte(string(subnet.IP))
	maskBitLen, netBitLen := subnet.Mask.Size()
	subnetString := subnet.String()
	if _, exist := (*ipam.Subnets)[subnetString]; !exist {
		(*ipam.Subnets)[subnetString] = strings.Repeat("0", 1 << uint8(netBitLen - maskBitLen))
	}

	for  index, value := range (*ipam.Subnets)[subnetString] {
		if value == '1' {
			continue
		}

		// 将数据更新为已经分配
		ipalloc := []byte((*ipam.Subnets)[subnetString])
		ipalloc[index] = '1'
		(*ipam.Subnets)[subnetString] = string(ipalloc)

		for t := uint(4); t > 0; t-- {
			ip[t - 1] += uint8(index >> ((4 - t) * 8))
		}

		// 从1开始分配的
		ip[3] += 1
		break
	}

	if err := ipam.dump(); err != nil {
		return nil, fmt.Errorf("allocate ip failed, message:%s", err.Error())
	}

	return &net.IPNet{
		IP:ip,
		Mask:[]byte(subnet.Mask),
	}, nil
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr net.IP) error  {
	if !ipam.Loaded {
		if err := ipam.Load(); err != nil {
			return fmt.Errorf("Release failed, error:%s", err.Error())
		}
	}

	releaseIp := ipaddr.To4()
	releaseIp[3] -= 1

	index := 0
	for t := uint(4); t > 0; t-- {
		index += int(releaseIp[t - 1] - subnet.IP[t - 1]) << ((4 - t) * 8)
	}


	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[index] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	if err := ipam.dump(); err != nil {
		return fmt.Errorf("relase failed, message:%s", err.Error())
	}

	return nil
}

func TestAllocate()  {
	ipam := IPAM{
		SubnetAllocatedPath: "/tmp/ipam.json",
		Subnets:&map[string]string{},
	}

	_, ipnet, _ := net.ParseCIDR("172.17.0.0/24")
	allocatedIpNet, err := ipam.Allocate(ipnet)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("ip address:%v", allocatedIpNet.IP)
}

func TestRelase()  {
	ipam := IPAM{
		SubnetAllocatedPath: "/tmp/ipam.json",
		Subnets:&map[string]string{},
	}

	ip, ipnet, _ := net.ParseCIDR("172.17.0.3/24")
	if err := ipam.Release(ipnet, ip); err != nil {
		fmt.Println(err)
		return
	}
}
