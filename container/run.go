package container

import (
	"fmt"
	"github.com/johnnylei/my_docker/subsystem"
	"github.com/johnnylei/my_docker/util"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Run(c *cli.Context) error  {
	if len(c.Args()) < 1 {
		return fmt.Errorf("missing container command")
	}

	read, write, err := util.NewPipe()
	if err != nil {
		return err
	}

	cmd := exec.Command("/proc/self/exe", "init", "-name", c.String("name"))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}

	if tty := c.Bool("ti"); tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.ExtraFiles = append(cmd.ExtraFiles, read)
	if _, err := write.WriteString(strings.Join(c.Args(), " ")); err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	resourceConfig := &subsystem.ResourceConfig {
		MemoryLimit: c.Int("m"),
		CpuSet: c.String("cpuset"),
		CpuShare: c.String("cpushare"),
	}

	manager, err := subsystem.InitCgroupsManager("mydocker-cgroup", resourceConfig)
	if err != nil {
		return err
	}

	if err := manager.Run(cmd.Process.Pid); err != nil {
		return err
	}

	defer manager.Destroy()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}