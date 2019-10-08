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

	instances := []Subsystem{}
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

	return &CgroupsManager{
		SubsystemInstances:instances,
	}, nil
}

type CgroupsManager struct {
	SubsystemInstances []Subsystem
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

type Subsystem interface {
	Name() string
	Apply(pid int) error
	Remove() error
}

func InitMemorySubsystem(relatePath string, res *ResourceConfig) (*MemorySubsystem, error) {
	if res == nil {
		return nil, fmt.Errorf("resource config should not be empty")
	}

	subsystem := &MemorySubsystem{
		Path: path.Join(CgroupBasePah, "memory", relatePath),
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
		Path: path.Join(CgroupBasePah, "cpuset", relatePath),
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

type MemorySubsystem struct {
	Path string
}

func (Subsystem *MemorySubsystem) Name() string  {
	return "memory"
}

func (subsystem *MemorySubsystem) Apply(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("pid should not be 0")
	}

	if err:= ioutil.WriteFile(path.Join(subsystem.Path, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (subsystem *MemorySubsystem) Remove() error  {
	return os.Remove(subsystem.Path)
}

type CpuSetSubsystem struct {
	Path string
}

func (s *CpuSetSubsystem) Name() string  {
	return "cpuset"
}

func (subsystem *CpuSetSubsystem) Remove() error  {
	return os.Remove(subsystem.Path)
}

func (subsystem *CpuSetSubsystem) Apply(pid int) error  {
	if pid <= 0 {
		return fmt.Errorf("pid should not be 0")
	}

	if err:= ioutil.WriteFile(path.Join(subsystem.Path, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		log.Fatal(err)
	}

	return nil
}

