[runc阅读参考链接](https://segmentfault.com/a/1190000017543294)

## abstract
Open Container Initiative(OCI)，OCI具体包含两个标准：运行时标准(runtime-spec)和容器镜像标准(image-spec)。简单来说，容器镜像标准定义了容器镜像的打包形式(pack format)，而运行时标准定义了如何去运行一个容器。

runC是一个遵循OCI标准的用来运行容器的命令行工具(CLI Tool)，它也是一个Runtime的实现。
```cassandraql
$ docker info
```