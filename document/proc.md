/proc[参考链接](https://www.jianshu.com/p/3fba2e5b1e17)

Linux 内核提供了一种通过 /proc 文件系统，在运行时访问内核内部数据结构、改变内核设置的机制。proc文件系统是一个伪文件系统，它只存在内存当中，而不占用外存空间。它以文件系统的方式为访问系统内核数据的操作提供接口。

用户和应用程序可以通过proc得到系统的信息，并可以改变内核的某些参数。由于系统的信息，如进程，是动态改变的，所以用户或应用程序读取proc文件时，proc文件系统是动态从系统内核读出所需信息并提交的。下面列出的这些文件或子文件夹，并不是都是在你的系统中存在，这取决于你的内核配置和装载的模块。另外，在/proc下还有三个很重要的目录：net，scsi和sys。 Sys目录是可写的，可以通过它来访问或修改内核的参数，而net和scsi则依赖于内核配置。例如，如果系统不支持scsi，则scsi 目录不存在。

除了以上介绍的这些，还有的是一些以数字命名的目录，它们是进程目录。系统中当前运行的每一个进程都有对应的一个目录在/proc下，以进程的 PID号为目录名，它们是读取进程信息的接口。而self目录则是读取进程本身的信息接口。

很多系统工具都是简单的去读取这个文件系统的某个文件内容, 比如`lsmod`就是读取`/proc/modules`

读取/proc/self/maps可以得到当前进程的内存映射关系，通过读该文件的内容可以得到内存代码段基址。

/proc/self/mem是进程的内存内容，通过修改该文件相当于直接修改当前进程的内存

```cassandraql
/proc/N # PID 为N的进程信息
/proc/N/cmdline 进程启动命令
/proc/N/cwd 链接到进程当前工作目录
/proc/N/environ 进程环境变量列表
/proc/N/exe 链接到进程的执行命令文件
/proc/N/fd 包含进程相关的所有文件描述符
/proc/N/maps 与进程相关的内存映射信息
/proc/N/mem 指代进程持有的内存, 不可读
/proc/N/root 链接到进程的根目录
/proc/N/stat 进程的状态
/proc/N/statm 进程使用的内核状态
/proc/N/status 进程的状态信息, 比stat/statm更加具有可读性
/proc/self 链接到当前正在运行的进程

/proc/filesystem #内核支持的文件系统
```


## 扩展阅读2
[cgroups - Linux control groups](http://www.man7.org/linux/man-pages/man7/cgroups.7.html)

[关于虚拟机的博客](https://segmentfault.com/u/wuyangchun)

[创建并管理cgroup](https://segmentfault.com/a/1190000007241437)

[proc](http://man7.org/linux/man-pages/man5/proc.5.html)