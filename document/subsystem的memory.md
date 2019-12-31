[原文链接](https://segmentfault.com/a/1190000008125359)
## why
代码总会有bug，有时会有内存泄漏，或者有意想不到的内存分配情况，或者这是个恶意程序，运行起来就是为了榨干系统内存，让其它进程无法分配到足够的内存而出现异常，如果系统配置了交换分区，会导致系统大量使用交换分区，从而系统运行很慢。

- 站在一个普通Linux开发者的角度，如果能控制一个或者一组进程所能使用的内存数，那么就算代码有bug，内存泄漏也不会对系统造成影响，因为可以设置内存使用量的上限，当到达这个值之后可以将进程重启。
- 站在一个系统管理者的角度，如果能限制每组进程所能使用的内存量，那么不管程序的质量如何，都能将它们对系统的影响降到最低，从而保证整个系统的稳定性。

## 内存控制能控制些什么？
- 限制cgroup中所有进程所能使用的物理内存总量
- 限制cgroup中所有进程所能使用的物理内存+交换空间总量(CONFIG_MEMCG_SWAP)： 一般在server上，不太会用到swap空间，所以不在这里介绍这部分内容。
- 限制cgroup中所有进程所能使用的内核内存总量及其它一些内核资源(CONFIG_MEMCG_KMEM)： 限制内核内存有什么用呢？其实限制内核内存就是限制当前cgroup所能使用的内核资源，比如进程的内核栈空间，socket所占用的内存空间等，通过限制内核内存，当内存吃紧时，可以阻止当前cgroup继续创建进程以及向内核申请分配更多的内核资源。由于这块功能被使用的较少，本篇中也不对它做介绍。

## 内核相关的配置
- 由于memory subsystem比较耗资源，所以内核专门添加了一个参数cgroup_disable=memory来禁用整个memory subsystem，这个参数可以通过GRUB在启动系统的时候传给内核，加了这个参数后内核将不再进行memory subsystem相关的计算工作，在系统中也不能挂载memory subsystem。
- 上面提到的CONFIG_MEMCG_SWAP和CONFIG_MEMCG_KMEM都是扩展功能，在使用前请确认当前内核是否支持，下面看看ubuntu 16.04的内核：
```cassandraql
#这里CONFIG_MEMCG_SWAP和CONFIG_MEMCG_KMEM等于y表示内核已经编译了该模块，即支持相关功能
dev@dev:~$ cat /boot/config-`uname -r`|grep CONFIG_MEMCG
CONFIG_MEMCG=y
CONFIG_MEMCG_SWAP=y
# CONFIG_MEMCG_SWAP_ENABLED is not set
CONFIG_MEMCG_KMEM=y
```
- CONFIG_MEMCG_SWAP控制内核是否支持Swap Extension，而CONFIG_MEMCG_SWAP_ENABLED（3.6以后的内核新加的参数）控制默认情况下是否使用Swap Extension，由于Swap Extension比较耗资源，所以很多发行版（比如ubuntu）默认情况下会禁用该功能（这也是上面那行被注释掉的原因），当然用户也可以根据实际情况，通过设置内核参数swapaccount=0或者1来手动禁用和启用Swap Extension。

## 怎么控制
在ubuntu 16.04里面，systemd已经帮我们将memory绑定到了/sys/fs/cgroup/memory
```cassandraql
#如果这里发现有多行结果，说明这颗cgroup数被绑定到了多个地方，
#不过不要担心，由于它们都是指向同一颗cgroup树，所以它们里面的内容是一模一样的
dev@dev:~$ mount|grep memory
cgroup on /sys/fs/cgroup/memory type cgroup (rw,nosuid,nodev,noexec,relatime,memory)
```

## 创建子cgroup
在/sys/fs/cgroup/memory下创建一个子目录即创建了一个子cgroup

```
#--------------------------第一个shell窗口----------------------
dev@dev:~$ cd /sys/fs/cgroup/memory
dev@dev:/sys/fs/cgroup/memory$ sudo mkdir test
dev@dev:/sys/fs/cgroup/memory$ ls test
cgroup.clone_children  memory.kmem.failcnt             memory.kmem.tcp.limit_in_bytes      memory.max_usage_in_bytes        memory.soft_limit_in_bytes  notify_on_release
cgroup.event_control   memory.kmem.limit_in_bytes      memory.kmem.tcp.max_usage_in_bytes  memory.move_charge_at_immigrate  memory.stat                 tasks
cgroup.procs           memory.kmem.max_usage_in_bytes  memory.kmem.tcp.usage_in_bytes      memory.numa_stat                 memory.swappiness
memory.failcnt         memory.kmem.slabinfo            memory.kmem.usage_in_bytes          memory.oom_control               memory.usage_in_bytes
memory.force_empty     memory.kmem.tcp.failcnt         memory.limit_in_bytes   
```

```cassandraql
 cgroup.event_control       #用于eventfd的接口
 memory.usage_in_bytes      #显示当前已用的内存
 memory.limit_in_bytes      #设置/显示当前限制的内存额度
 memory.failcnt             #显示内存使用量达到限制值的次数
 memory.max_usage_in_bytes  #历史内存最大使用量
 memory.soft_limit_in_bytes #设置/显示当前限制的内存软额度
 memory.stat                #显示当前cgroup的内存使用情况
 memory.use_hierarchy       #设置/显示是否将子cgroup的内存使用情况统计到当前cgroup里面
 memory.force_empty         #触发系统立即尽可能的回收当前cgroup中可以回收的内存
 memory.pressure_level      #设置内存压力的通知事件，配合cgroup.event_control一起使用
 memory.swappiness          #设置和显示当前的swappiness
 memory.move_charge_at_immigrate #设置当进程移动到其他cgroup中时，它所占用的内存是否也随着移动过去
 memory.oom_control         #设置/显示oom controls相关的配置
 memory.numa_stat           #显示numa相关的内存
```

## 添加进程
和“[创建并管理cgroup](https://segmentfault.com/a/1190000007241437)”中介绍的一样，往cgroup中添加进程只要将进程号写入cgroup.procs就可以了
```cassandraql
#--------------------------第二个shell窗口----------------------
#重新打开一个shell窗口，避免相互影响
dev@dev:~$ cd /sys/fs/cgroup/memory/test/
dev@dev:/sys/fs/cgroup/memory/test$ echo $$
4589
dev@dev:/sys/fs/cgroup/memory/test$ sudo sh -c "echo $$ >> cgroup.procs"
#运行top命令，这样这个cgroup消耗的内存会多点，便于观察
dev@dev:/sys/fs/cgroup/memory/test$ top
#后续操作不再在这个窗口进行，避免在这个bash中运行进程影响cgropu里面的进程数及相关统计
```
## 设置限额
```cassandraql
#--------------------------第一个shell窗口----------------------
#回到第一个shell窗口
dev@dev:/sys/fs/cgroup/memory$ cd test
#这里两个进程id分别时第二个窗口的bash和top进程
dev@dev:/sys/fs/cgroup/memory/test$ cat cgroup.procs
4589
4664
#开始设置之前，看看当前使用的内存数量，这里的单位是字节
dev@dev:/sys/fs/cgroup/memory/test$ cat memory.usage_in_bytes
835584

#设置1M的限额
dev@dev:/sys/fs/cgroup/memory/test$ sudo sh -c "echo 1M > memory.limit_in_bytes"
#设置完之后记得要查看一下这个文件，因为内核要考虑页对齐, 所以生效的数量不一定完全等于设置的数量
dev@dev:/sys/fs/cgroup/memory/test$ cat memory.limit_in_bytes
1048576

#如果不再需要限制这个cgroup，写-1到文件memory.limit_in_bytes即可
dev@dev:/sys/fs/cgroup/memory/test$ sudo sh -c "echo -1 > memory.limit_in_bytes"
#这时可以看到limit被设置成了一个很大的数字
dev@dev:/sys/fs/cgroup/memory/test$ cat memory.limit_in_bytes
9223372036854771712
```
一旦设置了内存限制，将立即生效，并且当物理内存使用量达到limit的时候，memory.failcnt的内容会加1，但这时进程不一定就会被kill掉，内核会尽量将物理内存中的数据移到swap空间上去，如果实在是没办法移动了（设置的limit过小，或者swap空间不足），默认情况下，就会kill掉cgroup里面继续申请内存的进程。

## 触发控制
当物理内存达到上限后，系统的默认行为是kill掉cgroup中继续申请内存的进程，那么怎么控制这样的行为呢？答案是配置memory.oom_control
```cassandraql
$ cat /sys/fs/cgroup/memory/memory.oom_control
oom_kill_disable 0
under_oom 0
```
这个文件里面包含了一个控制是否为当前cgroup启动OOM-killer的标识。如果写0到这个文件，将启动OOM-killer，当内核无法给进程分配足够的内存时，将会直接kill掉该进程；如果写1到这个文件，表示不启动OOM-killer，当内核无法给进程分配足够的内存时，将会暂停该进程直到有空余的内存之后再继续运行；同时，memory.oom_control还包含一个只读的under_oom字段，用来表示当前是否已经进入oom状态，也即是否有进程被暂停了。







