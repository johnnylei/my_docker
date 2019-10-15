```cassandraql
mount -t type -o options device dir
# type: 文件系统类型 aufs, proc, memory ...
# options: 挂载选项
# device(必选): 设备， 有些文件不需要指定具体的设备， 可以随便写一个字符串
# dir(必选)

$ mount -t proc none /proc # 挂载内核文件系统
$ mount -t aufs -o=./test1=ro:./test2=rw ./root #aufs式挂载
$ mount -t cgroup -o none,name=cgroup-test cgroup-test ./cgroup-test #挂载一个hierarchy
$ mount -t tmpfs tmpfs /tmp/tmpfs #在/tmp/tmpfs文件夹挂载成为一个内存文件系统,这样读取/tmp/tmpfs的速度就会快

# iso文件
$ mkdir -p iso/subdir01
$ mkisofs -o ./test.iso ./iso
$ mount ./test.iso /mnt
```
