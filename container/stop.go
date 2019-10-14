package container

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

func Stop(context *cli.Context) error  {
	containerName := context.String("name")
	if containerName == "" {
		return fmt.Errorf("container name should not be empty")
	}

	information, err := LoadContainerInformation(containerName)
	if err != nil {
		return err
	}

	process := &os.Process{
		Pid:information.Pid,
	}
	
	if err := process.Kill(); err != nil {
		return fmt.Errorf("kill process %d failed, err:%s", process.Pid, err.Error())
	}

	information.Stop()
	if err := information.Record(); err != nil {
		return err
	}

	return nil
}
