package container

var (
	WORK_SPACE_ROOT string = "/root/my_docker_workspace"
	CONTAINER_FILE_SYSTEM_MOUNT_ROOT string = "/root/my_docker_workspace/mnt"
	STATUS_RUNING string = "running"
	STATUS_STOP string = "stopped"
	STATUS_EXIT string = "exited"
	ConfigName string = "config.json"
	InformationFileName string = "information.json"
	LogFileName string = "log.log"
	ErrorLogFileName string = "error.log"
	DefaultContainerInformationLocation string = "/var/run/mydocker/"
	ENV_CONTAINER_PID string = "ENV_CONTAINER_PID"
	ENV_CONTAINER_EXEC_COMMAND string = "ENV_CONTAINER_EXEC_COMMAND"
	EXEC_PARENT_PROCESS_ID string = "EXEC_PARENT_PROCESS_ID"
	CONTAINER_FILE_SYSTEM_ROOT string = "/root/my_docker_workspace/containers"
	IMAGE_REGISTRY string = "/root/my_docker_workspace/image_registry"
)
