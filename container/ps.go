package container

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"path"
)

func Ps(context *cli.Context) error  {
	containers, err := ioutil.ReadDir(DefaultContainerInformationLocation)
	if err != nil {
		return fmt.Errorf("read dir %s failed, error:%s", DefaultContainerInformationLocation, err.Error())
	}

	if len(containers) == 0 {
		return nil
	}

	fmt.Printf("id\tname\tcreate_time\tcommand\tstatus\t")
	for _, container := range containers {
		containerInformationPath := path.Join(DefaultContainerInformationLocation, container.Name(), InformationFileName)
		informationByte, err := ioutil.ReadFile(containerInformationPath)
		if err != nil {
			continue
		}

		information := &ContainerInformation{}
		if err := json.Unmarshal(informationByte, information); err != nil {
			continue
		}

		fmt.Printf("%s\t%s\t%s\t%s\t%s\t", information.Id, information.Name, information.CreatedTime, information.InitCommand, information.Status)
	}

	return nil
}
