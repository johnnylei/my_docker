package container

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

func Delete(c *cli.Context) error  {
	path := "/tmp"
	name := c.String("name")
	if name == "" {
		return fmt.Errorf("container name should not be null")
	}

	if  err := DestroyContainerFileSystem(path, name); err != nil {
		return err
	}

	return nil
}

func DestroyContainerFileSystem(path string, name string) error  {
	mountContainerPath := fmt.Sprintf("%s/mnt/%s", path, name)
	//cmd := exec.Command("umount", mountContainerPath)
	//if err := cmd.Run(); err != nil {
	//	return fmt.Errorf("umount %s failed", mountContainerPath)
	//}

	if err := os.Remove(mountContainerPath); !os.IsNotExist(err) {
		return fmt.Errorf("remove %s failed", mountContainerPath)
	}

	containerPath := fmt.Sprintf("%s/%s", path, name)
	if err := os.Remove(containerPath); err != nil {
		return fmt.Errorf("remove %s failed", containerPath)
	}

	return nil
}
