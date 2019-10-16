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
#define STRLEN(str) strlen(str) + 1

char ** split(char * source, const char * delimiter) {
	char ** ret = malloc(sizeof(char *));
	char * item;
	int item_len = 0;
	char * token;
	token = strtok(source, delimiter);
	if (NULL == token) {
		item_len = STRLEN(source);
		item = (char *)malloc(item_len * sizeof(char));
		strncpy(item, source, item_len);
		ret[0] = item;
		return ret;
	}

	int i = 0;
	while(1) {
		item_len = STRLEN(token);
		item = malloc(sizeof(char) * item_len);
		strncpy(item, token, item_len);
		ret[i] = item;
		token = strtok(NULL, delimiter);
		if (NULL == token) {
			break;
		}

		i++;
		ret = realloc(ret, (i + 1) * sizeof(char *));
	}

	return ret;
}

void destroyTwoDimensionalArray(char ** arr) {
	int i = 0;
	while(1) {
		if (arr[i] == NULL) {
			break;
		}

		free(arr[i]);
		i++;
	}

	free(arr);
}

__attribute__((constructor)) int enter_namespace(void) {
	char * exec_parent_process_id = getenv("EXEC_PARENT_PROCESS_ID");
	if (NULL == exec_parent_process_id) {
		return -1;
	}

	if (atoi(exec_parent_process_id) != getppid()) {
		return -1;
	}

	char * read_buffer = malloc(BUFFER_SIZE);
	int ret = read(3, read_buffer, BUFFER_SIZE);
	close(3);
	if (ret == -1) {
		return -1;
	}

	char ** exec_message = split(read_buffer, ";;");
	free(read_buffer);
	char *pid = exec_message[0];
	char *command = exec_message[1];

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
		destroyTwoDimensionalArray(exec_message);
		printf("exec %s faile, error:%s\n", command, strerror(errno));
		return -1;
	}

	destroyTwoDimensionalArray(exec_message);
	return 0;
}
 */
 import "C" // import "c" 首先必须要独立写，其次与c代码之间不能更有空格
 import (
	 "fmt"
	 "github.com/johnnylei/my_docker/common"
	 "github.com/johnnylei/my_docker/util"
	 "github.com/urfave/cli"
	 "io/ioutil"
	 "os"
	 "os/exec"
	 "strconv"
	 "strings"
 )

func Exec(context *cli.Context) error {
	if context.Bool("child") {
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

	information, err := common.LoadContainerInformation(containerName)
	if err != nil {
		return err
	}

	read, write, err := util.NewPipe()
	defer write.Close()
	if err != nil {
		return fmt.Errorf("create pipe failed, error:%s\n", err.Error())
	}
	if _, err := write.WriteString(fmt.Sprintf("%d;;%s", information.Pid, command)); err != nil {
		return fmt.Errorf("write pid:%d, command:%s; to pipe failed\n", information.Pid, command)
	}

	cmd := exec.Command("/proc/self/exe", "exec", "-child")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.ExtraFiles = append(cmd.ExtraFiles, read)
	if err := os.Setenv(common.EXEC_PARENT_PROCESS_ID, strconv.Itoa(os.Getpid())); err != nil {
		return fmt.Errorf("set env EXEC_PARENT_PROCESS_ID %d failed, error:%s\n", cmd.Process.Pid, err.Error());
	}

	containerEnv, err := getEnvFromPid(information.Pid)
	if err == nil && containerEnv != nil {
		cmd.Env = append(os.Environ(), containerEnv...)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run self failed, error:%s\n", err.Error())
	}

	return nil
}

func getEnvFromPid(pid int) ([]string, error)  {
	envFilePath := fmt.Sprintf("/proc/%d/environ", pid)
	buffer, err := ioutil.ReadFile(envFilePath)
	if err != nil {
		return nil, fmt.Errorf("read %s failed, error:%s\n", envFilePath, err.Error())
	}

	return strings.Split(string(buffer), "\u0000"), nil
}