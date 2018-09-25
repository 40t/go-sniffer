
# go-sniffer

> 捕获mysql,redis,http,mongodb等协议...完善中


## Support List:
- [Mysql](#Mysql)
- [Redis](#Redis)
- [Http](#Http)
- ARP 测试中...
- Mongodb 开发中...
- Fast-cgi
- GRPC
- ...

## Demo:
![image](https://github.com/40t/go-sniffer/raw/master/images/demo.gif)
## Setup:
- 支持 : `MacOS` `Linux` `Unix`
- 不支持 : `windows`
- 依赖:`libcap` `google/gopacket`库

### 依赖库安装：Centos
``` bash
$ yum install libcap-devel
```
### 依赖库安装: Ubuntu
``` bash
$ apt install libcap-dev
```
### RUN
``` bash
$ go get github.com/google/gopacket
$ go get github.com/40t/go-sniffer
$ cd $GOPATH/src/github.com/40t/go-sniffer
$ go run main.go
```
## Usage:
``` bash
=======================================================================
[使用说明]

    go-sniffer [设备名] [插件名] [插件参数(可选)]

    [例子]
          go-sniffer en0 redis          抓取redis数据包
          go-sniffer en0 mysql -p 3306  抓取mysql数据包,端口3306

    go-sniffer --[命令]
               --help 帮助信息
               --env  环境变量
               --list 插件列表
               --ver  版本信息
               --dev  设备列表
    [例子]
          go-sniffer --list 查看可抓取的协议

=======================================================================
[设备名] : lo0 :   127.0.0.1
[设备名] : en0 : x:x:x:x:x5:x  192.168.1.3
[设备名] : utun2 :   1.1.11.1
=======================================================================
```