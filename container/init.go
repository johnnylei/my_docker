package container

import (
	"fmt"
	"github.com/johnnylei/my_docker/util"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func Init(c *cli.Context) error  {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	log.Println("current location is :", pwd)
	if err :=  InitContainerFilesystem(pwd, c); err != nil {
		return err
	}

	if err := syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NOSUID), ""); err != nil {
		return fmt.Errorf("mount /proc err: %s", err.Error())
	}

	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID | syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return fmt.Errorf("mount /dev err: %s", err.Error())
	}

	reader := os.NewFile(uintptr(3), "pipe")
	_, message, err := util.ReadPipe(reader)
	if err != nil {
		return fmt.Errorf("read pipe  err: %s", err.Error())
	}

	command := strings.Split(message, " ")
	path, err := exec.LookPath(command[0])
	if err != nil {
		return fmt.Errorf("look path %s err: %s", command, err.Error())
	}

	log.Println("path is: ", path)
	if err:= syscall.Exec(path, command[0:], os.Environ()); err !=nil {
		return fmt.Errorf("call %s  err: %s", path, err.Error())
	}

	return nil
}

func PivotRoot(root string) error  {
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND | syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("moint bind err: %s", err.Error())
	}

	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); !os.IsExist(err) {
		return fmt.Errorf("mkdir pivotDir err:%s", err.Error())
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot root error: %s", err.Error())
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("change root error: %s", err.Error())
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("umount %s error: %s", pivotDir, err.Error())
	}

	if err := os.Remove(pivotDir); err != nil {
		return fmt.Errorf("remove %s error: %s", pivotDir, err.Error())
	}

	return nil
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

	return PivotRoot(containerMountPath)
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
	if len(mounts) < 2 {
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


