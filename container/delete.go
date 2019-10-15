package container

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"os/exec"
)

func Delete(c *cli.Context) error  {
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

	if  err := DestroyContainerFileSystem(name); err != nil {
		return err
	}

	return nil
}

func DestroyContainerFileSystem(name string) error  {
	mountContainerPath := fmt.Sprintf("%s/%s", CONTAINER_FILE_SYSTEM_MOUNT_ROOT, name)
	cmd := exec.Command("umount", mountContainerPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("umount %s failed; error:%s\n", mountContainerPath, err.Error())
	}

	_, err := os.Stat(mountContainerPath)
	if err == nil {
		if err := os.Remove(mountContainerPath); err != nil {
			return fmt.Errorf("remove %s failed; error:%s\n", mountContainerPath, err.Error())
		}
	}

	containerPath := fmt.Sprintf("%s/%s", WORK_SPACE_ROOT, name)
	if err := exec.Command("rm", "-rf", containerPath).Run(); err != nil {
		return fmt.Errorf("remove %s failed; error:%s\n", containerPath, err.Error())
	}

	return nil
}
