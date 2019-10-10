package subsystem

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

const CgroupBasePah = "/sys/fs/cgroup"

type ResourceConfig struct {
	CpuSet string
	MemoryLimit int
	CpuShare string
}

func InitCgroupsManager(relatePath string, resourceConfig *ResourceConfig) (*CgroupsManager, error)  {
	if relatePath == "" || resourceConfig == nil {
		return nil, fmt.Errorf("relate path or resource config should not be empty")
	}

	instances := []SubsystemInterface{}
	if resourceConfig.MemoryLimit > 0 {
		memorySubsystem, err := InitMemorySubsystem(relatePath, resourceConfig)
		if err != nil {
			return nil, err
		}

		instances = append(instances, memorySubsystem)
	}

	if resourceConfig.CpuSet != "" {
		cpuSetSubsystem, err := InitCpuSetSubsystem(relatePath, resourceConfig)
		if err != nil {
			return nil, err
		}

		instances = append(instances, cpuSetSubsystem)
	}

	if resourceConfig.CpuShare != "" {
		subsystem, err := InitCpuShareSubsystem(relatePath, resourceConfig)
		if err != nil {
			return nil, err
		}

		instances = append(instances, subsystem)
	}

	return &CgroupsManager{
		SubsystemInstances:instances,
	}, nil
}

type CgroupsManager struct {
	SubsystemInstances []SubsystemInterface
}

func (cgroupsManager *CgroupsManager) Run(pid int) error  {
	for _, subsystem := range cgroupsManager.SubsystemInstances {
		if err := subsystem.Apply(pid); err != nil {
			return err
		}
	}

	return nil
}

func (cgroupsManager *CgroupsManager) Destroy() error  {
	for _, subsystem := range cgroupsManager.SubsystemInstances {
		if err := subsystem.Remove(); err != nil {
			return err
		}
	}

	return nil
}

func InitMemorySubsystem(relatePath string, res *ResourceConfig) (*MemorySubsystem, error) {
	if res == nil {
		return nil, fmt.Errorf("resource config should not be empty")
	}

	subsystem := &MemorySubsystem{
		&Subsystem{
			Path: path.Join(CgroupBasePah, "memory", relatePath),
		},
	}

	if _, err := os.Stat(subsystem.Path); os.IsNotExist(err) {
		if err := os.Mkdir(subsystem.Path, 0755); err != nil {
			log.Fatal(err)
		}
	}

	err := ioutil.WriteFile(path.Join(subsystem.Path, "memory.limit_in_bytes"), []byte(strconv.Itoa(res.MemoryLimit)), 0644)
	if err != nil {
		log.Fatal(err)
	}

	return subsystem, nil
}

func InitCpuSetSubsystem(relatePath string, res *ResourceConfig) (*CpuSetSubsystem, error) {
	if res == nil {
		return nil, fmt.Errorf("resource config should not be empty")
	}

	subsystem := &CpuSetSubsystem{
		&Subsystem{
			Path: path.Join(CgroupBasePah, "cpuset", relatePath),
		},
	}

	if _, err := os.Stat(subsystem.Path); os.IsNotExist(err) {
		if err := os.Mkdir(subsystem.Path, 0755); err != nil {
			log.Fatal(err)
		}
	}

	err := ioutil.WriteFile(path.Join(subsystem.Path, "cpuset.cpus"), []byte(res.CpuSet), 0644)
	if err != nil {
		log.Fatal(err)
	}

	return subsystem, nil
}

func InitCpuShareSubsystem(relatePath string, res *ResourceConfig) (*CpuSetSubsystem, error) {
	if res == nil {
		return nil, fmt.Errorf("resource config should not be empty")
	}

	subsystem := &CpuSetSubsystem{
		&Subsystem{
			Path: path.Join(CgroupBasePah, "cpu", relatePath),
		},
	}

	if _, err := os.Stat(subsystem.Path); os.IsNotExist(err) {
		if err := os.Mkdir(subsystem.Path, 0755); err != nil {
			log.Fatal(err)
		}
	}

	err := ioutil.WriteFile(path.Join(subsystem.Path, "cpu.shares"), []byte(res.CpuShare), 0644)
	if err != nil {
		log.Fatal(err)
	}

	return subsystem, nil
}

type SubsystemInterface interface {
	Remove() error
	Apply(pid int) error
}

type Subsystem struct {
	Path string
	Name string
}

func (subsystem *Subsystem) Remove() error  {
	return os.Remove(subsystem.Path)
}

func (subsystem *Subsystem) Apply(pid int) error  {
	if pid <= 0 {
		return fmt.Errorf("pid should not be 0")
	}

	if err:= ioutil.WriteFile(path.Join(subsystem.Path, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		log.Fatal(err)
	}

	return nil
}

type MemorySubsystem struct {
	*Subsystem
}

type CpuSetSubsystem struct {
	*Subsystem
}


type CpuShareSubsystem struct {
	*Subsystem
}
