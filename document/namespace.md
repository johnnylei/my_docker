## docker
- namespace
- cgroups
- aufs

## namespace
[参考链接](https://www.cnblogs.com/sparkdev/p/9365405.html)
- UTS: CLONE_NEWUTS
- IPC: CLONE_NEWIPC
- PID: CLONE_NEWPID
- USER: CLONE_NEWUSER
- MOUNT: CLONE_NEWNS
- NETWORK: CLONE_NEWNET
- CGROUP: CLONE_NEWCGROUP

### abstract
- clone: 创建一个新的进程并把他放到新的namespace中
- setns: 通过 setns() 函数可以将当前进程加入到已有的 namespace 中
- unshare: 使当前进程退出指定类型的namespace，并加入到新创建的namespace（相当于创建并加入新的namespace）

```
ls -l /proc/$$/ns #查看ns信息
pstree -pl #查看树装进程图
```

### UTS
- nodename: 主机名
- domainname: 域名

```
$ readlink /proc/3621/ns/uts #查看进程的uts信息
# 在main.go的交互进程运行
$ hostname #查看hostname
$ hostname -b bird #修改hostname
$ hostname
# 在正常的terminal
$ hostname
# 发现两个hostname是隔离的
```

### IPC
- System V IPC
- POSIX message queues

```
$ ipcs -q #查看系统队列
$ ipcmk -Q #创建一个系统队列
# 在main.go的交互进程和系统terminal发生隔离
```

### PID
- 隔离进程ID

```
# main.go
$ echo $$
# terminal
$ ps aux|grep main.go

# 发现两个结果不一样
```

### MOUNT
隔离各个进程看到的挂载点试图. 在各个不同Namespace中, 看到的文件系统层次是不一样的

/proc：这个目录本身是一个虚拟文件系统。他放置的数据都是在内存当中，例如系统内核、进程、外部设备的状态及网络状态等。因为这个目录下的数据都是在内存当中，所以本身不占任何硬盘空间，比较重要的文件有：/proc/cpuinfo /proc/dma,/proc/interrupts,/proc/oports,/proc/net，等

````$xslt
$ ps -ef
$ ls /proc
$ mount -t proc proc /proc
$ ps -ef
````

### User
- user
- user group

### network
网络隔离
