**本篇将介绍一个简单的subsystem，名字叫pids，功能是限制cgroup及其所有子孙cgroup里面能创建的总的task数量。**

subsystem挂载点: /sys/fs/cgroup/pids
```cassandraql
$ mount|grep pids
$ cd /sys/fs/cgroup/pids
$ mkdir test #创建cgroup
$ cd test
$ ls
cgroup.clone_children  cgroup.procs  notify_on_release  pids.current  pids.max  tasks
```
- pids.current:表示当前cgroup及其所有子孙cgroup中现有的总的进程数量
- pids.max: 当前cgroup及其所有子孙cgroup中所允许创建的总的最大进程数量，在根cgroup下没有这个文件，原因显而易见，因为我们没有必要限制整个系统所能创建的进程数量。

```cassandraql
$ cd /sys/fs/cgroup/pids/test
$ echo 1 > pids.max
$ echo $$ > cgroup.procs
$ ls
-bash: fork: retry: No child processes
-bash: fork: retry: No child processes
-bash: fork: retry: No child processes
-bash: fork: retry: No child processes
# 因为当前terminal进程限制为1,所以不能执行任何命令了，因为当前terminal就会占用一个进程
```