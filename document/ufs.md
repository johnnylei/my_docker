
[Linux文件系统之aufs](https://segmentfault.com/a/1190000008489207)
```cassandraql
$ mkdir image1 image2 container
$ echo image1 > image1/001 && echo image1 > image1/002
$ echo image2 > image2/002 && echo image2 > image1/003
# 默认模式就是copy on write, image1 可以读写的, image2是copy on write, write到image1中
$ mount -t aufs -o br=./image1:./image2 none ./container
$ ls container
001 002 003 # image2/002 被image1的002覆盖了, 所以 `cat continer/002` 显示image1
$ echo “container” >> container/001
# image01/001也会被修改
$ echo “container” >> container/003
# image02/003不会被修改, copy image02/003 -> image01/003 , 同时修改image01/003
# 以只读模式mount
$ mount -t aufs -o br=./image01=ro:./image02=ro none ./container
# 读写模式mount
$ mount -t aufs -o br=./image01=rw:./image02=rw none ./container
```

## docker如何使用aufs
每个image都是由一系列的read-only layer组成, image layer的内容都存在docker hosts filesystem的`/var/lib/docker/aufs/diff`目录下
- /var/lib/docker/aufs/diff # 自身镜像存放的东西
- /var/lib/docker/aufs/mnt #diff文件mount 后的结果
- /var/lib/docker/aufs/layers #镜像aufs祖先

```cassandraql
# dockerfile
FROM ubuntu:15.04
RUN echo "hello world" > /tmp/newfile

$ docker build -t ubuntu-changed .
$ docker hisotry ubuntu-changed # 查看dokcer image 构建过程
$ docker inpsect ubuntu-changed #查看镜像详细信息
```
#### container aufs
- /var/lib/docker/aufs/mnt: container layer的mount目录
- /var/lib/docker/containers/<container-id>: container的metadata和配置文件
- /var/lib/docker/diff: container read write真实目录
- /var/lib/docker/aufs/layers #container aufs祖先
- /sys/fs/aufs/ # container start以后会在这个生成一个文件夹
```
$ cat /sys/fs/aufs/si_b2a30563119048e6/* #查看相关的信息

/var/lib/docker/aufs/diff/066473c8437089ea4cada0e08bb35d8b88fd74c1cf28ae2efca59a7cfe8904da=rw
/var/lib/docker/aufs/diff/066473c8437089ea4cada0e08bb35d8b88fd74c1cf28ae2efca59a7cfe8904da-init=ro+wh
/var/lib/docker/aufs/diff/2b625d1f976bdb5f20b89390a6ed1c488706e0ab19a3523412aa2b3b4f630bbe=ro+wh
/var/lib/docker/aufs/diff/e1abb3cf52e73eabbb37d4af701bfc346dbd2db3579fc69434422934fc8390ca=ro+wh
/var/lib/docker/aufs/diff/6a5f75e6168b2c59ef9b93af1e946e995c934ced02a988f035707b506ff6e29c=ro+wh
/var/lib/docker/aufs/diff/3364f29aa94fc98a0298326667d03d9a4b757160ce6e4e0281e9a9c82beb220c=ro+wh
/var/lib/docker/aufs/diff/45461f598161b2f757bf1054ad72a6e342063cf44b010570642344827e051089=ro+wh
/var/lib/docker/aufs/diff/80c5fa11d914fafaee61de4d49571b523edb322e7764d645d6ad12b8d32a126e=ro+wh
/var/lib/docker/aufs/diff/dcc7e63313ef121d2f682ea890fe39a5f04a984ce2060aaf4a3eb5724f00ea1b=ro+wh
/var/lib/docker/aufs/diff/898c93b4e89d44fae521d00044262b482257befe1d941a73bf849fcff5f5d5ad=ro+wh
/var/lib/docker/aufs/diff/b26eb59c39b3f1ebc3aa0fc6adb308d14a4fa9ab9ef3aa090fd27325a07d7871=ro+wh
/var/lib/docker/aufs/diff/f09a02a4da8c532d83c15aeedbf6f7150df02a9fd0cfb9e40e921152a330a50a=ro+wh
/var/lib/docker/aufs/diff/f64dca8b61a787c57f1adad7e704c6850c6048cf0fc594f37e2f1de0ec1a798d=ro+wh
/var/lib/docker/aufs/diff/da9166939e9d1b68d06b06f38555d1447418398037803aa8520f93ed84fc8e1d=ro+wh
/var/lib/docker/aufs/diff/2ebf81c52c0cf0ecf3b313b374b025a866e92afd0249ab707d006e5cb89aafb8=ro+wh
/var/lib/docker/aufs/diff/e1ed703b12472a5673811df8f095a9fe50aee55d5688185d8ccb8318ddb09485=ro+wh
/var/lib/docker/aufs/diff/d45f75223e65bfebba54514c7085b7cd898fd84e5ba4fbcc94999a49038fbbc1=ro+wh
/var/lib/docker/aufs/diff/7b742585728f06cdff69bc56710d4d163ab154a357f5189a3d50ca87a45cee06=ro+wh
64
65
74
75
76
77
78
79
80
81
66
67
68
69
70
71
72
73
/dev/shm/aufs.xino

```