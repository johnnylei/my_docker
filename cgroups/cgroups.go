package cgroups

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const cgroupMemoryHierarchyMount  = "/sys/fs/cgroup/memory"

func Run()  {
	if os.Args[0] == "/proc/self/exe" {
		fmt.Printf("current pid %d\n", syscall.Getpid())
		cmd := exec.Command("stress --vm-bytes 200m --vm-keep -m 1")
		cmd.SysProcAttr = &syscall.SysProcAttr{}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	cmd := exec.Command("stress --vm-bytes 200m --vm-keep -m 1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNET | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", cmd.Process.Pid)
	if err = os.Mkdir(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit"), 0755); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, "testmemorylimit", "memory.limit_in_bytes"), []byte("100m"), 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err = cmd.Process.Wait(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}