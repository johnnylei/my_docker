package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type ContainerInformation struct {
	Pid int `json:"pid"`
	Id string `json:"id"`
	Name string `json:"name"`
	InitCommand string `json:"init_command"`
	Status string `json:"status"`
	CreatedTime string `json:"created_time"`
}

func (information *ContainerInformation) Record() error  {
	if _, err := os.Stat(DefaultContainerInformationLocation); os.IsNotExist(err) {
		if err := os.Mkdir(DefaultContainerInformationLocation, 077); err != nil {
			return fmt.Errorf("mkdir %s failed, err:%s", DefaultContainerInformationLocation, err.Error())
		}
	}

	BasePath := path.Join(DefaultContainerInformationLocation, information.Name)
	if _, err := os.Stat(BasePath); os.IsNotExist(err) {
		if err := os.Mkdir(BasePath, 0777); err != nil {
			return fmt.Errorf("mkdir %s failed, err:%s", BasePath, err.Error())
		}
	}

	InformationFile := path.Join(BasePath, InformationFileName)
	if _, err := os.Stat(InformationFile); os.IsNotExist(err) {
		if _, err := os.Create(InformationFile); err != nil {
			return fmt.Errorf("create %s failed, err:%s", InformationFile, err.Error())
		}
	}

	InformationFileFd, err := os.OpenFile(InformationFile, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open %s failed, error:%s", InformationFile, err.Error())
	}

	informationBytes, err := json.Marshal(information)
	if err != nil {
		return fmt.Errorf("json encode failed; err:%s", err.Error())
	}

	informationJson := string(informationBytes)
	if _, err := InformationFileFd.WriteString(informationJson); err != nil {
		return fmt.Errorf("wirte information failed; file:%s, information:%s, error:%s", InformationFile, informationJson, err.Error())
	}

	return nil
}

func (information *ContainerInformation) Destroy() error  {
	informationFile := path.Join(DefaultContainerInformationLocation, information.Name)
	if err := os.Remove(informationFile); err != nil {
		return fmt.Errorf("remove information file %s failed; error:%s", informationFile, err.Error())
	}

	return nil
}

