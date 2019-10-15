package container

import (
	"fmt"
	"github.com/johnnylei/my_docker/common"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path"
)

func Commit(context *cli.Context) error  {
	name := context.String("name")
	if name == "" {
		return fmt.Errorf("container name should not be empty")
	}

	fileSystemPath := path.Join(common.CONTAINER_FILE_SYSTEM_MOUNT_ROOT, name)
	if _, err := os.Stat(fileSystemPath); err != nil {
		return fmt.Errorf("%s error, error message:%s", fileSystemPath, err.Error())
	}

	imageTarName := fmt.Sprintf("%s.tar", name)
	cmd := exec.Command("tar", "-cf", imageTarName, "-C", fileSystemPath, ".")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tar failed, tar is:%s; file system path:%s", imageTarName, fileSystemPath)
	}

	return nil
}
