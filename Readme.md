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