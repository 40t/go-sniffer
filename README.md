
# go-sniffer

> Capture mysql,redis,http,mongodb etc protocol...

[![GitHub license](https://img.shields.io/github/license/40t/go-sniffer.svg?style=popout-square)](https://github.com/40t/go-sniffer/blob/master/LICENSE)



## Support List:
- [mysql](#mysql)
- [Redis](#redis)
- [Http](#http)
- [Mongodb](#mongodb)
- Kafka (developing)
- ...

## Demo:
``` bash
$ go-sniffer en0 mysql
```
![image](https://github.com/40t/go-sniffer/raw/master/images/demo.gif)
## Setup:
- support : `MacOS` `Linux` `Unix`
- not support : `windows`

### Centos
``` bash
$ yum install libcap-devel
```
### Ubuntu
``` bash
$ apt-get install libcap-dev
```
### MacOs
``` bash
All is ok
```
### RUN
``` bash
$ go get -v github.com/40t/go-sniffer
$ cd $(go env GOPATH)/src/github.com/40t/go-sniffer
$ go run main.go
```
## Usage:
``` bash
==================================================================================
[Usage]

    go-sniffer [device] [plug] [plug's params(optional)]

    [Example]
          go-sniffer en0 redis          Capture redis packet
          go-sniffer en0 mysql -p 3306  Capture mysql packet

    go-sniffer --[commend]
               --help "this page"
               --env  "environment variable"
               --list "Plug-in list"
               --ver  "version"
               --dev  "device"
    [Example]
          go-sniffer --list "show all plug-in"

==================================================================================
[device] : lo0 :   127.0.0.1
[device] : en0 : xx:xx:xx:xx:xx:xx  192.168.199.221
==================================================================================
```

### Example:
``` bash
$ go-sniffer lo0 mysql 
$ go-sniffer en0 redis 
$ go-sniffer eth0 http -p 8080
$ go-sniffer eth1 mongodb
```
## License:
[MIT](http://opensource.org/licenses/MIT)
