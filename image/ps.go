package image

import (
	"fmt"
	"github.com/johnnylei/my_docker/container"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
	"strings"
)

func Ps(context *cli.Context) error  {
	fmt.Printf("name\t\tpath\t\tcontainers")
	if err := filepath.Walk(container.IMAGE_REGISTRY, func(path string, info os.FileInfo, err error) error {
		if !strings.Contains(info.Name(), ".json") {
			return nil
		}

		imageInformation, err := LoadImageFromFile(path)
		if err != nil {
			return nil
		}

		fmt.Printf("%s\t\t%s\t\t%s", imageInformation.Name, imageInformation.Path, strings.Join(func() []string{
			if len(imageInformation.Containers) == 0 {
				return nil
			}

			containers := make([]string, 1)
			for _, item := range imageInformation.Containers {
				containers = append(containers, item.Name)
			}

			return containers
		}(), ";"))
		return nil
	}); err != nil {
		return fmt.Errorf("read %s failed; error %s", container.IMAGE_REGISTRY, err.Error())
	}

	return nil
}
