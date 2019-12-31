[原文链接](https://segmentfault.com/a/1190000008323952)
## subsystem
- cpu: cpu使用限制
- cpuset: cpu亲和性设置
- cpuacct: cpu统计信息

## /sys/fs/cgroup/cpu
- cpu.cfs_period_us  
- cpu.cfs_quota_us
- cgroup.procs
- cpu.shares
- cpu.stat


#### cpu.cfs_period_us && cpu.cfs_quota_us
- cpu.cfs_period_us: 配置周期时长
- cpu.cfs_quota_us: 当前cgroup周期时长内可以使用的时长; -1(默认值)表示不受限制
```cassandraql
$ cat /proc/cpuinfo #查看cpu信息
$ cd /sys/fs/cgroup/cpu
$ mkdir test
$ cat cpu.cfs_period_us
100000
$ echo  10000 > cpu.cfs_quota_us #表示当前cgroup最大只能使用10%cpu
$ echo $$ > cgroup.procs
$ stress --cpu 1
$ top #另外开一个窗口查看cpu使用量10%
```

#### cpu.shares
前提: 
- cgroup test下面如果有test1, test2 两个cgroup;
- test1的cpu.shares为1024(默认值); 
- test2的cpu.shares为512;

cpu忙时现象: 
- test1的cgroup只能占到test的1024/(1024+512) = 66%的cpu
- test2的cgroup只能占到test的512/(1024+512) = 33%的cpu

cpu闲时现象:
- test1或者test2最大可以占到 100% test cgroup的cpu

```
$ cd /sys/fs/cgroup/cpu/test
$ mkdir test1
$ mkdir test2
#窗口1
$ cd test2
$ echo 512 > cpu.shares
$ echo $$ > cgroup.procs
$ stress --cpu 1 

#窗口2
$ cd /sys/fs/cgroup/cpu/test/test1   
$ echo $$ > cgroup.procs
$ stress --cpu 1

#窗口3
$ top #可以看到两个stress 分别占用3.3% 和 6.6%的cpu; 因为test cgroup 限制最大了10%

# 此时如果杀掉窗口1的stress
$ top #可以看到一个stress 占用10%的cpu
```
结论:
- 在闲的时候，shares基本上不起作用，只有在CPU忙的时候起作用，这是一个优点。
- 由于shares是一个绝对值，需要和其它cgroup的值进行比较才能得到自己的相对限额，而在一个部署很多容器的机器上，cgroup的数量是变化的，所以这个限额也是变化的，自己设置了一个高的值，但别人可能设置了一个更高的值，所以这个功能没法精确的控制CPU使用率。

#### cpu.stat
- nr_periods： 表示过去了多少个cpu.cfs_period_us里面配置的时间周期
- nr_throttled： 在上面的这些周期中，有多少次是受到了限制（即cgroup中的进程在指定的时间周期中用光了它的配额）
- throttled_time: cgroup中的进程被限制使用CPU持续了多长时间(纳秒)
