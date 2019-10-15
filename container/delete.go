package container

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"os/exec"
)

func Delete(c *cli.Context) error  {
	path := "/tmp"
	name := c.String("name")
	if name == "" {
		return fmt.Errorf("container name should not be null")
	}

	information, err := LoadContainerInformation(name)
	if err != nil {
		return err
	}
	if information.Status == STATUS_RUNING {
		return fmt.Errorf("%s is running could not delete\n", information.Name)
	}

	if err := information.Destroy(); err != nil {
		return err
	}

	if  err := DestroyContainerFileSystem(path, name); err != nil {
		return err
	}

	return nil
}

func DestroyContainerFileSystem(path string, name string) error  {
	mountContainerPath := fmt.Sprintf("%s/mnt/%s", path, name)
	cmd := exec.Command("umount", mountContainerPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("umount %s failed; error:%s\n", mountContainerPath, err.Error())
	}

	if err := os.Remove(mountContainerPath); !os.IsNotExist(err) {
		return fmt.Errorf("remove %s failed; error:%s\n", mountContainerPath, err.Error())
	}

	containerPath := fmt.Sprintf("%s/%s", path, name)
	if err := os.Remove(containerPath); err != nil {
		return fmt.Errorf("remove %s failed; error:%s\n", containerPath, err.Error())
	}

	return nil
}
