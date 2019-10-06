## 环境
- ubuntu 14.04
- kernel 3.13.x
- go 1.71.1

### kernel降级
ubuntu 14.04卸载kernel4.04安装3.13.0.83-generic,由于默认安装的kernel是4.4, 而实验环境要求3.13.x

[参考文档](https://blog.csdn.net/u013431916/article/details/82530523)
```cassandraql
$ vim /etc/apt/sources.list
# 在文件最后增加并保存如下
deb http://security.ubuntu.com/ubuntu trusty-security main

$ apt-get update
$ apt-get install linux-image-extra-3.13.0-83-generic
# 查看系统安装kernel
$ sudo dpkg --get-selections  |  grep  'linux' 
$ vim /etc/default/grub
# 找到如下
GRUB_DEFAULT=0
# 修改为
GRUB_DEFAULT="Advanced options for Ubuntu>Ubuntu, with Linux 3.13.0-83-generic"
$ update-grub
$ reboot
$ uname -r
```

### kernel降级以后一个docker问题
降级后发现docker ps 错误了
```cassandraql
$ docker ps
Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
$ service docker status # docker是运行的
docker start/running, process 26524

# 查看错误日志
$ tail -f /var/log/upstart/docker.log
/var/run/docker.sock is up
time="2019-10-05T23:54:42.714599727+08:00" level=info msg="libcontainerd: new containerd process, pid: 20285" 
time="2019-10-05T23:54:43.721743867+08:00" level=error msg="'overlay' not found as a supported filesystem on this host. Please ensure kernel is new enough and has overlay support loaded." 
time="2019-10-05T23:54:43.722175526+08:00" level=error msg="[graphdriver] prior storage driver overlay2 failed: driver not supported" 
Error starting daemon: error initializing graphdriver: driver not supported

```
没搞定
