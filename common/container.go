package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

func LoadContainerInformation(Name string) (*ContainerInformation, error)  {
	informationPath := path.Join(DefaultContainerInformationLocation, Name, InformationFileName)
	return LoadContainerInformationFormFIle(informationPath)
}

func LoadContainerInformationFormFIle(informationPath string) (*ContainerInformation, error)  {
	informationBytes, err := ioutil.ReadFile(informationPath)
	if err != nil {
		return nil, fmt.Errorf("read %s failed, error: %s\n", informationPath, err.Error())
	}

	information := &ContainerInformation{}
	if err := json.Unmarshal(informationBytes, information); err != nil {
		return nil, fmt.Errorf("load %s failed, error:%s\n", informationPath, err.Error())
	}
	return information,err
}

type ContainerInformation struct {
	Pid int `json:"pid"`
	Id string `json:"id"`
	Name string `json:"name"`
	InitCommand string `json:"init_command"`
	Status string `json:"status"`
	CreatedTime string `json:"created_time"`
	Path string
}

func (information *ContainerInformation) GetPath() string  {
	if information.Path != "" {
		return information.Path
	}

	information.Path = path.Join(DefaultContainerInformationLocation, information.Name, InformationFileName)
	return information.Path
}

func (information *ContainerInformation) CheckExist() bool  {
	_, err := os.Stat(information.GetPath())
	if err != nil {
		return false
	}

	return true
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

	InformationFileFd, err := os.OpenFile(InformationFile, os.O_WRONLY | os.O_TRUNC, 0644)
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
	if err := exec.Command("rm", "-rf", informationFile).Run(); err != nil {
		return fmt.Errorf("remove information file %s failed; error:%s", informationFile, err.Error())
	}

	return nil
}

func (information *ContainerInformation) Stop()  {
	information.Status = STATUS_STOP
	information.Pid = 0
	information.CreatedTime = ""
	information.InitCommand = ""
}