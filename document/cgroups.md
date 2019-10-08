## abstract
linux cgroups 提供了一组进程及将来的子进程的资源限制, 控制, 统计的能力, 这些资源包括CPU, 内存, 存储, 网络等

- cgroup
- task
- subsystem
- hierarchy

### cgroup
cgroup是对进程分组管理的一种机制, 一个cgroup包含一组进程, 并可以在这个cgroup上增加linux subsystem的各种参数配置, 将一组进程和一组subsystem的系统参数关联起来

### task
tasks 里面是一些进程

### subsystem
- blkio 块设备输入输出的控制
- cpu 设置cgroup中进程的CPU被调度的策略
- cpuacct 统计cgroup里面的进程对cpu的占用
- cpuset 多核机器上, 设置可以使用的cpu和内存
- devices 控制cgroup中进程对设备的访问
- freezer 用于挂起, 恢复cgroup中的进程
- memory 控制cgroup中进程对内存的占用
- net_cls 用于将cgroup中进程产生的网络包分类, 以便Linux的traffic controller可以根分类区分出来自某个cgroup的包并对其限流或监控
- net_prio 设置cgroup中进程产生的网络流量的优先级
- ns 使cgroup中进程在新的Namespace中的fork新进程是, 创建出新的cgroup, 这个cgroup包含新的namespace中的进程

```
$ lssubsys -a
$ mount|grep cgroup
$ cat /proc/$$/cgroup #查看当前进程的cgroup信息
```
### hierarchy
把一组cgroup串成树装, 这样的树就是hierarchy, 通过这样的树装结构, cgroup可以实现继承,

### 关系
- 系统在创建新的hierarchy之后,系统所有的进程都会加入这个hierarchy的cgroup跟节点, 这个cgroup的根节点是hierarchy默认创建的
- 一个subsystem只能附加到一个hierarchy上面
- 一个hierarchy可以附加多个subsystem
- 一个进程可以作为多个cgroup的成员, 但是这些cgroup必须在不同hierarchy上面
- 一个进程fork出子进程时,子进程是和父进程在同一个cgroup中的,也可以根据需要移动到其他的cgroup中

## 操作
#### 创建并挂载一个hierarchy树
```
$ mkdir cgroup-test #创建一个挂载点
$ sudo mount -t cgroup -o none,name=cgroup-test cgroup-test ./cgroup-test #挂载一个hierarchy
$ ls ./cgroup-test
cgroup.clone_children  cgroup.procs  cgroup.sane_behavior  notify_on_release  release_agent  tasks
```

- cgroup.clone_children, cpuset的subsystem会读取这个配置文件, 如果值为1, 子cgroup才会继承cgroup的cpuset的配置
- cgroup.procs 树中当前节点cgroup中进程组的ID, 现在的位置在跟节点, 这个文件会有现在系统中所有进程组的ID
- notify_on_release和release_agent会一起使用, notify_on_release标识这个cgroup最后一个进程退出的是否执行了release_agent;
- release_agent是一个路径, 通常用作进程退出之后自动清理掉不再使用的cgroup
- tasks标识该cgroup下面的进程ID, 如果把一个进程ID写到tasks文件中, 便会将想对应的进程加入到这个cgroup中

#### 扩展出两个子cgroup
```
$ cd cgroup-test
$ sudo mkdir cgroup-1
$ sudo mkdir cgroup-2
$ tree
```

他们会继承父cgroup的属性

#### 在cgroup中添加进程
```
$ cd cgroup-1
$ sudo sh -c "echo $$ >> tasks"
$ cat /proc/$$/cgroup
```
看到当前进程被添加到cgroup-test:/cgroup-1中了

#### 通过subsystem限制cgroup中的进程的资源
```
# 系统为每个subsystem分配了一个默认的hierarchy, 比如memory的hierarchy
$ mount|grep memory
# 首先在不作限制的情况下, 使用stress
$ stress --vm-bytes 200m --vm-keep -m 1
$ top # 本机内存8g, 发现存在占用2.7%

# 现在进行内存限制测试
$ cd /sys/fs/cgroup/memory #通过前面的命令发现memory这个subsystem挂在这里
$ mkdir test-limit-memory && cd test-limit-memory
$ sudo sh -c "echo \"100m\" > memory.limit_in_bytes"
$ sudo sh -c "echo $$ > tasks"

$ stress --vm-bytes 200m --vm-keep -m 1
$ top # 本机内存8G, 现在内存占用1.3%, 正好少了一半

#杀死被挂载的进程, 就可以umount
$  sudo umount /home/johnny/go_workspace/src/github.com/johnnylei/my_docker/cgroup-test
# umount 完成以后,就可以删除cgoup-test文件夹
```

## docker是如何使用cgroups
```
$ docker run -itd -m 128m centos
$ /sys/fs/cgroup/memory/docker/50e3a136839bd2d557f8a6989bc50b8e6bb637c8b430d1eec51b5a7d018ae0b9
$ cat memory.limit_in_bytes
134217728
$ cat memory.usage_in_bytes
2273280
```