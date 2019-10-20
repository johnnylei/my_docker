package container

import (
	"fmt"
	"github.com/johnnylei/my_docker/common"
	"github.com/urfave/cli"
)

func Delete(c *cli.Context) error  {
	name := c.String("name")
	if name == "" {
		return fmt.Errorf("container name should not be null")
	}

	information, err := common.LoadContainerInformation(name)
	if err != nil {
		return err
	}
	if information.Status == common.STATUS_RUNING {
		return fmt.Errorf("%s is running could not delete\n", information.Name)
	}

	if err := information.Destroy(); err != nil {
		return err
	}

	return nil
}
