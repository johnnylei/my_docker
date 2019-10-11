package container

import (
	"fmt"
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
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	log.Println("current location is :", pwd)
	if err :=  PivotRoot(pwd); err != nil {
		return err
	}

	if err := syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NOSUID), ""); err != nil {
		return err
	}

	reader := os.NewFile(uintptr(3), "pipe")
	_, message, err := util.ReadPipe(reader)
	if err != nil {
		return err
	}

	command := strings.Split(message, " ")
	path, err := exec.LookPath(command[0])
	if err != nil {
		return err
	}

	if err:= syscall.Exec(path, command[0:], os.Environ()); err !=nil {
		return err
	}

	return nil
}

func PivotRoot(root string) error  {
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND | syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("moint bind err: %s", err.Error())
	}

	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("mkdir pivotDir err:%s", err.Error())
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot root error: %s", err.Error())
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("change root error: %s", err.Error())
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("umount %s error: %s", pivotDir, err.Error())
	}

	return os.Remove(pivotDir)
}



