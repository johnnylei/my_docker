package container

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"path"
)

func InitLogFile(context *cli.Context) (*os.File, *os.File, error) {
	if _, err := os.Stat(DefaultContainerInformationLocation); os.IsNotExist(err) {
		if err := os.Mkdir(DefaultContainerInformationLocation, 077); err != nil {
			return nil, nil, fmt.Errorf("mkdir %s failed, err:%s", DefaultContainerInformationLocation, err.Error())
		}
	}

	containerName := context.String("name")
	BasePath := path.Join(DefaultContainerInformationLocation, containerName)
	if _, err := os.Stat(BasePath); os.IsNotExist(err) {
		if err := os.Mkdir(BasePath, 0777); err != nil {
			return nil, nil, fmt.Errorf("mkdir %s failed, err:%s", BasePath, err.Error())
		}
	}

	logFile := path.Join(DefaultContainerInformationLocation, containerName, LogFileName)
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

	errorLogFile := path.Join(DefaultContainerInformationLocation, containerName, ErrorLogFileName)
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
