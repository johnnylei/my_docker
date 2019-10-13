package container

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
)

func Ps(context *cli.Context) error  {
	containers, err := ioutil.ReadDir(DefaultContainerInformationLocation)
	if err != nil {
		return fmt.Errorf("read dir %s failed, error:%s", DefaultContainerInformationLocation, err.Error())
	}

	if len(containers) == 0 {
		return nil
	}

	//containerInformationList := []ContainerInformation{}
	for _, container := range containers {
		log.Println(container.Name())
	}

	return nil
}
