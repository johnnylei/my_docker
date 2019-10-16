package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"syscall"
)

type IPAM struct {
	SubnetAllocatedPath string
	Subnets *map[string]string
}

func (ipam *IPAM) dump() error  {
	SubnetsJsonBytes, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return fmt.Errorf("json marshal subnets failed")
	}

	_, err = os.Stat(ipam.SubnetAllocatedPath)
	if !os.IsNotExist(err) {
		return fmt.Errorf("get %s stat faield, error:%s", ipam.SubnetAllocatedPath, err.Error())
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(ipam.SubnetAllocatedPath, 0644); err != nil {
			return fmt.Errorf("mk file %s failed, error:%s", ipam.SubnetAllocatedPath, err.Error())
		}
	}

	err = ioutil.WriteFile(ipam.SubnetAllocatedPath, SubnetsJsonBytes, syscall.O_WRONLY | syscall.O_TRUNC)
	if err != nil {
		return fmt.Errorf("dump subnets failed, write %s failed, error:%s", ipam.SubnetAllocatedPath, err.Error())
	}

	return nil
}

func (ipam *IPAM) load() error  {
	SubnetsJsonBytes, err := ioutil.ReadFile(ipam.SubnetAllocatedPath)
	if err != nil {
		return fmt.Errorf("load failed, read %s failed, error:%s", ipam.SubnetAllocatedPath, err.Error())
	}

	err = json.Unmarshal(SubnetsJsonBytes, ipam.Subnets)
	if err != nil {
		return fmt.Errorf("load failed, json Unmarshal %s failed, error:%s", string(SubnetsJsonBytes), err.Error())
	}

	return err
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (*net.IPNet, error)  {
	if err := ipam.load(); err != nil {
		return nil, err
	}

	ip := subnet.IP
	maskBitLen, netBitLen := subnet.Mask.Size()
	subnetString := subnet.String()
	if _, exist := *(ipam.Subnets)[subnetString]; !exist {
		*(ipam.Subnets)[subnetString] = strings.Repeat("0", 1 << (netBitLen - maskBitLen))
	}

	for  index, value := range *(ipam.Subnets)[subnetString] {
		if value == '1' {
			continue
		}

		// 将数据更新为已经分配
		ipalloc := []byte(*(ipam.Subnets)[subnetString])
		ipalloc[index] = '1'
		*(ipam.Subnets)[subnetString] = string(ipalloc)

		for t := uint(4); t > 0; t-- {
			ip[t - 4] += uint8(index >> (t - 4) * 8)
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
		Mask:subnet.Mask,
	}, nil
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error  {
	if err := ipam.load(); err != nil {
		return fmt.Errorf("Release failed, error:%s", err.Error())
	}

	releaseIp := ipaddr.To4()
	releaseIp[3] -= 1

	index := 0
	for t := uint(4); t > 0; t-- {
		index += int(releaseIp[t - 1] - subnet.IP[t - 1]) << ((4 - t) * 8)
	}

	ipalloc := []byte(*(ipam.Subnets)[subnet.String()])
	ipalloc[index] = '0'
	*(ipam.Subnets)[subnet.String()] = string(ipalloc)

	if err := ipam.dump(); err != nil {
		return fmt.Errorf("relase failed, message:%s", err.Error())
	}

	return nil

}
