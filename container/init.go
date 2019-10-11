package container

import (
	"github.com/johnnylei/my_docker/util"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func Init(c *cli.Context) error  {
	e := &util.MyContainerError{}
	e.Message = "container Init;"
	pwd, err := os.Getwd()
	if err != nil {
		e.Message += err.Error()
		return e
	}

	log.Println("current location is :", pwd)
	if err :=  PivotRoot(pwd); err != nil {
		e.Message += err.Error()
		return e
	}

	if err := syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NOSUID), ""); err != nil {
		e.Message += err.Error()
		return e
	}

	reader := os.NewFile(uintptr(3), "pipe")
	_, message, err := util.ReadPipe(reader)
	if err != nil {
		e.Message += err.Error()
		return e
	}

	command := strings.Split(message, " ")
	path, err := exec.LookPath(command[0])
	if err != nil {
		e.Message += err.Error()
		return e
	}

	if err:= syscall.Exec(path, command[0:], os.Environ()); err !=nil {
		e.Message += err.Error()
		return e
	}

	return nil
}

func PivotRoot(root string) error  {
	e := &util.MyContainerError{}
	e.Message = "PivotRoot;"
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND | syscall.MS_REC, ""); err != nil {
		e.Message += err.Error()
		return e
	}

	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		e.Message += err.Error()
		return e
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		e.Message += err.Error()
		return e
	}

	if err := os.Chdir("/"); err != nil {
		e.Message += err.Error()
		return e
	}

	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		e.Message += err.Error()
		return e
	}

	return nil
}



