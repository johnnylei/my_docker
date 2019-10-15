package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func InitImage(Name string) *Image  {
	image, err := LoadImage(Name)
	if err == nil {
		return image
	}

	return &Image{
		Name:Name,
		Path:fmt.Sprintf("%s/%s.json", IMAGE_REGISTRY, Name),
	}
}

func LoadImage(Name string) (*Image, error)  {
	imageInformationPath := fmt.Sprintf("%s/%s.json", IMAGE_REGISTRY, Name)
	return LoadImageFromFile(imageInformationPath)
}

func LoadImageFromFile(imageInformationPath string) (*Image,  error)  {
	imageInformationBytes, err := ioutil.ReadFile(imageInformationPath)
	if err != nil {
		return nil, fmt.Errorf("read %s failed, error: %s\n", imageInformationPath, err.Error())
	}

	image := &Image{}
	if err := json.Unmarshal(imageInformationBytes, image); err != nil {
		return nil, fmt.Errorf("load %s failed, error:%s\n", imageInformationPath, err.Error())
	}
	return image, err
}

type Image struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Containers []*ContainerInformation `json:"containers"`
}

func (image *Image) Record() error  {
	InformationFile := fmt.Sprintf("%s/%s.json", IMAGE_REGISTRY, image.Name)
	if _, err := os.Stat(InformationFile); os.IsNotExist(err) {
		if _, err := os.Create(InformationFile); err != nil {
			return fmt.Errorf("create %s failed, err:%s", InformationFile, err.Error())
		}
	}

	InformationFileFd, err := os.OpenFile(InformationFile, os.O_WRONLY | os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open %s failed, error:%s", InformationFile, err.Error())
	}

	informationBytes, err := json.Marshal(image)
	if err != nil {
		return fmt.Errorf("json encode failed; err:%s", err.Error())
	}

	informationJson := string(informationBytes)
	if _, err := InformationFileFd.WriteString(informationJson); err != nil {
		return fmt.Errorf("wirte information failed; file:%s, information:%s, error:%s", InformationFile, informationJson, err.Error())
	}

	return nil
}

func (image *Image) AppendContainer(container *ContainerInformation)  {
	for index, item := range image.Containers {
		if item.Name == container.Name {
			image.Containers[index] = container
			return
		}
	}
	image.Containers = append(image.Containers, container)
}