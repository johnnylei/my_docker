package container

import (
	"fmt"
	"github.com/johnnylei/my_docker/common"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path"
)

func Logs(context *cli.Context) error  {
	containerName := context.String("name")
	if containerName == "" {
		return fmt.Errorf("contianer name should not be empty")
	}

	LogFileName, errorLogFileName := GetLogPath(containerName)
	if _, err := os.Stat(LogFileName); err != nil {
		return fmt.Errorf("logfile %s is invalid, error:%s", LogFileName, err.Error())
	}

	cmd := exec.Command("tail", "-n", "10", LogFileName)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tail execute failed; err :%s", err.Error())
	}

	fmt.Printf("for more information please visit file:%s, and error messag please visit file: %s\n", LogFileName, errorLogFileName)
	return nil
}

func GetLogPath(ContainerName string) (string, string) {
	return path.Join(common.DefaultContainerInformationLocation, ContainerName, common.LogFileName),
	path.Join(common.DefaultContainerInformationLocation, ContainerName, common.ErrorLogFileName)
}

func InitLogFile(context *cli.Context) (*os.File, *os.File, error) {
	if _, err := os.Stat(common.DefaultContainerInformationLocation); os.IsNotExist(err) {
		if err := os.Mkdir(common.DefaultContainerInformationLocation, 077); err != nil {
			return nil, nil, fmt.Errorf("mkdir %s failed, err:%s", common.DefaultContainerInformationLocation, err.Error())
		}
	}

	containerName := context.String("name")
	BasePath := path.Join(common.DefaultContainerInformationLocation, containerName)
	if _, err := os.Stat(BasePath); os.IsNotExist(err) {
		if err := os.Mkdir(BasePath, 0777); err != nil {
			return nil, nil, fmt.Errorf("mkdir %s failed, err:%s", BasePath, err.Error())
		}
	}

	logFile := path.Join(common.DefaultContainerInformationLocation, containerName, common.LogFileName)
	logFileResource, err := os.OpenFile(logFile, os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0644)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("open log file %s failed; error:%s", logFile, err.Error())
	}

	if err != nil && os.IsNotExist(err) {
		logFileResource, err = os.Create(logFile)
		if err != nil {
			return nil, nil, fmt.Errorf("create log file %s failed; error:%s", logFile, err.Error())
		}
	}

	errorLogFile := path.Join(common.DefaultContainerInformationLocation, containerName, common.ErrorLogFileName)
	errorLogFileResource, err := os.OpenFile(errorLogFile, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("open error log file %s failed; error:%s", errorLogFile, err.Error())
	}

	if err != nil && os.IsNotExist(err) {
		errorLogFileResource, err = os.Create(errorLogFile)
		if err != nil {
			return nil, nil, fmt.Errorf("create log file %s failed; error:%s", logFile, err.Error())
		}
	}

	return logFileResource, errorLogFileResource, nil
}
