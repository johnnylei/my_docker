package container

var (
	WorkSpaceRoot  = "/tmp"
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
)
