```cassandraql
$ df -l
Filesystem     1K-blocks    Used Available Use% Mounted on
udev             1005712       4   1005708   1% /dev
tmpfs             203620     616    203004   1% /run
/dev/dm-0       18180876 2776448  14457844  17% /
none                   4       0         4   0% /sys/fs/cgroup
none                5120       0      5120   0% /run/lock
none             1018092       0   1018092   0% /run/shm
none              102400       0    102400   0% /run/user
/dev/sda1         240972   76259    152272  34% /boot
none            18180876 2776448  14457844  17% /tmp/mnt/mydocker
$ free
             total       used       free     shared    buffers     cached
Mem:       2036184     385816    1650368        620      35692     236644
-/+ buffers/cache:     113480    1922704
Swap:      2093052          0    2093052

$ mount -o remount,size=777M tmpfs /dev/shm #mount时可以根据实际需要调整大小
```
是一个临时文件系统，驻留在内存中，所以/dev/shm/这个目录不在硬盘上，而是在内存里。因为是在内存里，所以读写非常快，可以提供较高的访问速度。linux下，tmpfs默认最大为内存的一半大小，使用df -h命令刚才已经看到了，但是这个df查看到的挂载内存大小的数值，如果没有使用，是没有去真正占用的，只有真正在tmpfs存储数据了，才会去占用。比如，tmpfs大小是499M,用了10M大小，内存里就会使用真正使用10M，剩余的489M是可以继续被服务器其他程序来使用的。但是因为数据是在内存里，所以断电后文件会丢失，内存数据不会和硬盘中数据一样可以永久保存。了解了tmpfs这个特性可以用来提高服务器性能，把一些对读写性能要求较高，但是数据又可以丢失的这样的数据保存在/dev/shm中，来提高访问速度。

```cassandraql
$ vim  /etc/fstab 
tmpfs                   /dev/shm                tmpfs   defaults,size=777M     0 0
```

- tmpfs就是一种存在于内存的文件系统, 
- 可以挂载到任意的文件夹下面, 提高读取速率
- 同时不会持久化保持
```cassandraql
$ mount -t tmpfs tmpfs /tmp/tmpfs #在/tmp/tmpfs文件夹挂载成为一个内存文件系统,这样读取/tmp/tmpfs的速度就会快
```