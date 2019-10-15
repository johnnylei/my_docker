package container

import (
	"fmt"
	"github.com/johnnylei/my_docker/common"
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

	if err := InitContainerFilesystem(c); err != nil {
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
	} else {
		logFile, errorLogFile, err := InitLogFile(c)
		if err != nil {
			return err
		}

		cmd.Stdout = logFile
		cmd.Stderr = errorLogFile
	}

	containerInitCommand := strings.Join(c.Args(), " ")
	cmd.ExtraFiles = append(cmd.ExtraFiles, read)
	if _, err := write.WriteString(containerInitCommand); err != nil {
		return err
	}

	envSlice := c.StringSlice("env")
	if envSlice != nil {
		cmd.Env = append(os.Environ(), envSlice...)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	uid := util.Uid()
	fmt.Printf("uid:%s\n", uid)
	containerInformation := &common.ContainerInformation{
		Pid: cmd.Process.Pid,
		Id: uid,
		Name: c.String("name"),
		InitCommand: containerInitCommand,
		Status: common.STATUS_RUNING,
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	if containerInformation.CheckExist() {
		return fmt.Errorf("contianer %s is exist", containerInformation.Name)
	}
	if err := containerInformation.Record(); err != nil {
		return err
	}
	imageObject := common.InitImage(c.String("image"))
	imageObject.AppendContainer(containerInformation)
	if err := imageObject.Record(); err != nil {
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
		fmt.Sprintf("%s/%s", common.CONTAINER_FILE_SYSTEM_MOUNT_ROOT, c.String("name")),
	}
	if c.String("v") != "" {
		ret = append(ret, "-v", c.String("v"))
	}


	return ret
}


func CreateImageLayer(imageName string) (string, error)  {
	imageTarPath := fmt.Sprintf("%s/%s.tar", common.IMAGE_REGISTRY, imageName)
	_, err := os.Stat(imageTarPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("there is not image file: %s", imageTarPath)
	}else if err != nil {
		return "", fmt.Errorf("tar file %s error, message:%s\n", imageTarPath, err.Error())
	}

	imagePath := fmt.Sprintf("%s/%s", common.IMAGE_REGISTRY, imageName)
	_, err = os.Stat(imagePath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(imagePath, 0777); err != nil {
			return "", fmt.Errorf("mkdir %s failed", imagePath)
		}
	}else if err != nil {
		return "", fmt.Errorf("image file %s error, message:%s\n", imagePath, err.Error())
	}

	command := exec.Command("tar", "-xvf", imageTarPath, "-C", imagePath)
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("tar failed: %s; tar:%s, target:%s", err.Error(), imageTarPath, imagePath)
	}
	return imagePath, nil
}

func CreateContainerLayer(name string) (string, error)  {
	if name == "" {
		return "", fmt.Errorf("create container layer container name should not be empty")
	}

	containerPath := fmt.Sprintf("%s/%s", common.WORK_SPACE_ROOT, name)
	_, err := os.Stat(containerPath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(containerPath, 0777); err != nil {
			return "", fmt.Errorf("create container layer, create container path failed; %s", err.Error())
		}
	}

	return containerPath, nil
}

func CreateContainerMountLayer(name string) (string, error)  {
	if name == "" {
		return "", fmt.Errorf("create container mount layer container name should not be empty")
	}

	_, err := os.Stat(common.CONTAINER_FILE_SYSTEM_MOUNT_ROOT)
	if os.IsNotExist(err) {
		if err := os.Mkdir(common.CONTAINER_FILE_SYSTEM_MOUNT_ROOT, 0777); err != nil {
			return "", fmt.Errorf("create container mount layer, create container mount path failed; %s; %s",
				err.Error(), common.CONTAINER_FILE_SYSTEM_MOUNT_ROOT)
		}
	}

	mountPath := fmt.Sprintf("%s/%s", common.CONTAINER_FILE_SYSTEM_MOUNT_ROOT, name)
	_, err = os.Stat(mountPath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(mountPath, 0777); err != nil {
			return "", fmt.Errorf("create container mount layer, create container mount path failed; %s; %s",
				err.Error(), mountPath)
		}
	}

	return mountPath, nil
}

func InitContainerFilesystem(c *cli.Context) error  {
	name := c.String("name")
	if name == "" {
		return fmt.Errorf("init container file system; container name should not be empty")
	}

	imageLayerPath, err := CreateImageLayer(c.String("image"))
	if err != nil {
		return err
	}

	containerLayerPath, err := CreateContainerLayer(name)
	if err != nil {
		return err
	}

	containerMountPath, err := CreateContainerMountLayer(name)
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

	if err := initContainerVolume(c); err != nil {
		return err
	}

	return nil
}

func initContainerVolume(c *cli.Context) error  {
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

	destinationMount := fmt.Sprintf("%s/%s%s", common.CONTAINER_FILE_SYSTEM_MOUNT_ROOT, c.String("name"), mounts[1])

	if err := exec.Command("mkdir", "-p", destinationMount).Run(); err != nil {
		return fmt.Errorf("mkdir %s failed", destinationMount)
	}

	log.Printf("mounting volume: source:%s; destination:%s", mounts[0], destinationMount)
	mountOptions := fmt.Sprintf("dirs=%s", mounts[0])
	if err := exec.Command("mount", "-t", "aufs", "-o", mountOptions, "none", destinationMount).Run(); err != nil {
		return fmt.Errorf("init Cointainer volumen failed; mount failed; source:%s; destination:%s",
			mounts[0], destinationMount)
	}

	return nil
}