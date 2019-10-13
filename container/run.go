package container

import (
	"fmt"
	"github.com/johnnylei/my_docker/subsystem"
	"github.com/johnnylei/my_docker/util"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var (
	WorkSpaceRoot  = "/tmp"
	STATUS_RUNING string = "running"
	STATUS_STOP string = "stop"
	STATUS_EXIT string = "exited"
	ConfigName string = "config.json"
	InformationFileName string = "information.json"
	DefaultContainerInformationLocation string = "/var/run/mydocker/"
)

func Run(c *cli.Context) error  {
	if len(c.Args()) < 1 {
		return fmt.Errorf("missing container command")
	}

	tty :=  c.Bool("ti")
	detach := c.Bool("d")
	if tty && detach {
		return fmt.Errorf("-ti and -d should not be both exist")
	}

	read, write, err := util.NewPipe()
	if err != nil {
		return err
	}

	if err := InitContainerFilesystem(WorkSpaceRoot, c); err != nil {
		return err
	}

	cmd := exec.Command("/proc/self/exe", BuildInitArgs(c)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	containerInitCommand := strings.Join(c.Args(), " ")
	cmd.ExtraFiles = append(cmd.ExtraFiles, read)
	if _, err := write.WriteString(containerInitCommand); err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	containerInformation := &ContainerInformation{
		Pid: cmd.Process.Pid,
		Id: util.Uid(),
		Name: c.String("name"),
		InitCommand: containerInitCommand,
		Status: STATUS_RUNING,
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := containerInformation.Record(); err != nil {
		return err
	}

	resourceConfig := &subsystem.ResourceConfig {
		MemoryLimit: c.Int("m"),
		CpuSet: c.String("cpuset"),
		CpuShare: c.String("cpushare"),
	}

	manager, err := subsystem.InitCgroupsManager("mydocker-cgroup", resourceConfig)
	if err != nil {
		return err
	}

	if err := manager.Run(cmd.Process.Pid); err != nil {
		return err
	}

	defer manager.Destroy()

	if tty {
		if err := cmd.Wait(); err != nil {
			return err
		}

		if err := containerInformation.Destroy(); err != nil {
			return  err
		}
	}

	return nil
}

func BuildInitArgs(c *cli.Context) []string  {
	ret := []string{
		"init",
		"-name",
		c.String("name"),
		"-root",
		fmt.Sprintf("%s/mnt/%s", WorkSpaceRoot, c.String("name")),
	}
	if c.String("v") != "" {
		ret = append(ret, "-v", c.String("v"))
	}


	return ret
}


func CreateImageLayer(path string, imageName string) (string, error)  {
	if path == "" {
		return "", fmt.Errorf("create image layer path should not be empty")
	}

	imageTarPath := fmt.Sprintf("%s/busybox.tar", path)
	if _, err := os.Stat(imageTarPath); os.IsNotExist(err) {
		return "", fmt.Errorf("there is not image file: %s", imageTarPath)
	}

	imagePath := fmt.Sprintf("%s/%s", path, imageName)
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		if err := os.Mkdir(imagePath, 0777); err != nil {
			return "", fmt.Errorf("mkdir %s failed", imagePath)
		}
	}
	command := exec.Command("tar", "-xvf", imageTarPath, "-C", imagePath)
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("tar failed: %s; tar:%s, target:%s", err.Error(), imageTarPath, imagePath)
	}

	return imagePath, nil
}

func CreateContainerLayer(path string, name string) (string, error)  {
	if path == "" {
		return "", fmt.Errorf("create container layer path should not be empty")
	}

	if name == "" {
		return "", fmt.Errorf("create container layer container name should not be empty")
	}

	containerPath := fmt.Sprintf("%s/%s", path, name)
	if err := os.Mkdir(containerPath, 0777); !os.IsExist(err) {
		return "", fmt.Errorf("create container layer, create container path failed; %s", err.Error())
	}

	return containerPath, nil
}

func CreateContainerMountLayer(path string, name string) (string, error)  {
	if path == "" {
		return "", fmt.Errorf("create container mount layer path should not be empty")
	}

	if name == "" {
		return "", fmt.Errorf("create container mount layer container name should not be empty")
	}

	mountPath := fmt.Sprintf("%s/mnt", path)
	if err := os.Mkdir(mountPath, 0777); !os.IsExist(err) {
		return "", fmt.Errorf("create container mount layer, create container mount path failed; %s; %s", err.Error(), mountPath)
	}

	mountPath = fmt.Sprintf("%s/%s", mountPath, name)
	if err := os.Mkdir(mountPath, 0777); !os.IsExist(err) {
		return "", fmt.Errorf("create container mount layer, create container mount path failed; %s; %s", err.Error(), mountPath)
	}

	return mountPath, nil
}

func InitContainerFilesystem(path string, c *cli.Context) error  {
	if path == "" {
		return fmt.Errorf("init container file system; path should not be empty")
	}

	name := c.String("name")
	if name == "" {
		return fmt.Errorf("init container file system; container name should not be empty")
	}

	imageLayerPath, err := CreateImageLayer(path, "busybox")
	if err != nil {
		return err
	}

	containerLayerPath, err := CreateContainerLayer(path, name)
	if err != nil {
		return err
	}

	containerMountPath, err := CreateContainerMountLayer(path, name)
	if err != nil {
		return err
	}

	// mount -t aufs -o dirs=containerLayer:imageLayer none ./container
	mountOptions := fmt.Sprintf("dirs=%s:%s", containerLayerPath, imageLayerPath)
	log.Printf("mount root: %s, %s", mountOptions, containerMountPath)
	cmd := exec.Command("mount", "-t", "aufs", "-o", mountOptions, "none", containerMountPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("InitContainerFilesystem mount error, %s", err.Error())
	}

	if err := initContainerVolume(path, c); err != nil {
		return err
	}

	return nil
}

func initContainerVolume(path string, c *cli.Context) error  {
	if path == "" {
		return fmt.Errorf("init container volume; path should not be empty")
	}

	volume := c.String("v")
	if volume == "" {
		return nil
	}

	mounts := strings.Split(volume, ":")
	if len(mounts) != 2 {
		return fmt.Errorf("invalid volume, usage -v source:destination")
	}

	if _, err := os.Stat(mounts[0]); os.IsNotExist(err) {
		return fmt.Errorf("source mount not exist; %s, %v", mounts[0], []byte(mounts[0]))
	}

	destinationMount := fmt.Sprintf("%s/mnt/%s%s", path, c.String("name"), mounts[1])

	if err := exec.Command("mkdir", "-p", destinationMount).Run(); err != nil {
		return fmt.Errorf("mkdir %s failed", destinationMount)
	}

	log.Printf("mounting volume: source:%s; destination:%s", mounts[0], destinationMount)
	mountOptions := fmt.Sprintf("dirs=%s", mounts[0])
	if err := exec.Command("mount", "-t", "aufs", "-o", mountOptions, "none", destinationMount).Run(); err != nil {
		return fmt.Errorf("init Cointainer volumen failed; mount failed; source:%s; destination:%s", mounts[0], destinationMount)
	}

	return nil
}