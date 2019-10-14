package container
/*
#include <stdio.h>
#include <unistd.h>
#include <errno.h>
#include <string.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <dirent.h>
#include <fcntl.h>
#include <sched.h>

#define BUFFER_SIZE 256
#define LEN(a) sizeof(a)/sizeof(a[0])

__attribute__((constructor)) int enter_namespace(void) {
	printf("enter namespace running");
	char * pid;
	char * command;
	pid = getenv("ENV_CONTAINER_PID");
	command = getenv("ENV_CONTAINER_EXEC_COMMAND");
	if (NULL == pid || NULL == command) {
		printf("pid %s or command %s should not be empty\n", pid, command);
		return -1;
	}

	// 要非常注意把mnt放在最后面，不然文件系统就会出错，找不到文件了，会报open xxx failed
	const char *namespace[] = {"ipc", "pid", "uts", "net", "mnt"};
	char namespace_path[BUFFER_SIZE];
	int i;
	for (i = 0; i < LEN(namespace); i++) {
		memset(namespace_path, '\0', BUFFER_SIZE);
		sprintf(namespace_path, "/proc/%s/ns/%s", pid, namespace[i]);
		int fd = open(namespace_path, O_RDONLY);
		if (fd == -1) {
			printf("open %s failed, error:%s\n", namespace_path, strerror(errno));
			return -1;
		}

		if (setns(fd) == -1) {
			close(fd);
			printf("setns as %s faile, error:%s\n", namespace_path, strerror(errno));
			return -1;
		}

		close(fd);
	}

	if (system(command) == -1) {
		printf("exec %s faile, error:%s\n", command, strerror(errno));
		return -1;
	}

	return 0;
}
 */
 import "C"
import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Exec(context *cli.Context) error {
	if os.Getenv(ENV_CONTAINER_PID) != "" {
		fmt.Printf("%s runnig\n", os.Getenv(ENV_CONTAINER_EXEC_COMMAND))
		return nil
	}

	containerName := context.String("name")
	if containerName == "" {
		return fmt.Errorf("contaner name is required")
	}

	command := strings.Join(context.Args(), " ")
	if command == "" {
		return fmt.Errorf("command should not be empty")
	}

	information, err := LoadContainerInformation(containerName)
	if err != nil {
		return err
	}

	if err := os.Setenv(ENV_CONTAINER_EXEC_COMMAND, command); err != nil {
		return fmt.Errorf("set env %s failed, error:%s\n", ENV_CONTAINER_EXEC_COMMAND, err.Error())
	}

	if err := os.Setenv(ENV_CONTAINER_PID, strconv.Itoa(information.Pid)); err != nil {
		return fmt.Errorf("set env %s failed, error:%s\n", ENV_CONTAINER_PID, err.Error())
	}

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run self failed, error:%s\n", err.Error())
	}

	return nil
}
