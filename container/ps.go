package container

import (
	"fmt"
	"github.com/johnnylei/my_docker/common"
	"github.com/urfave/cli"
	"io/ioutil"
	"path"
)

func Ps(context *cli.Context) error  {
	containers, err := ioutil.ReadDir(common.DefaultContainerInformationLocation)
	if err != nil {
		return fmt.Errorf("read dir %s failed, error:%s", common.DefaultContainerInformationLocation, err.Error())
	}

	if len(containers) == 0 {
		return nil
	}

	fmt.Printf("id\t\tname\t\tcreate_time\t\tcommand\t\tstatus\t\t\n")
	for _, container := range containers {
		containerInformationPath := path.Join(common.DefaultContainerInformationLocation, container.Name(), common.InformationFileName)
		information, err := common.LoadContainerInformationFormFIle(containerInformationPath)
		if err != nil {
			continue
		}

		fmt.Printf("%s\t\t%s\t\t%s\t\t%s\t\t%s\t\t\n", information.Id, information.Name, information.CreatedTime, information.InitCommand, information.Status)
	}

	return nil
}
