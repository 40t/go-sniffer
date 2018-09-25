package build

import (
	"github.com/google/gopacket"
	"io"
	"strings"
	"fmt"
	"strconv"
	"bufio"
)

type Redis struct {
	port int
	version string
	cmd chan string
	done chan bool
}

const (
	Port       int = 6379
	Version string = "0.1"
	CmdPort string = "-p"
)

var redis = &Redis {
	port:Port,
	version:Version,
}

func NewInstance() *Redis{
	return redis
}

/**
	解析流
 */
func (red Redis) ResolveStream(net, transport gopacket.Flow, r io.Reader) {

	//只解析clint发出去的包
	buf := bufio.NewReader(r)
	var cmd string
	var cmdCount = 0
	for {

		line, _, _ := buf.ReadLine()
		//判断链接是否断开
		if len(line) == 0 {
			buff := make([]byte, 1)
			_, err := r.Read(buff)
			if err == io.EOF {
				red.done <- true
				return
			}
		}

		//过滤无用数据
		if !strings.HasPrefix(string(line), "*") {
			continue
		}

		//过滤服务器返回数据
		if strings.EqualFold(transport.Src().String(), strconv.Itoa(red.port)) == true {
			continue
		}

		//解析
		l := string(line[1])
		cmdCount, _ = strconv.Atoi(l)
		cmd = ""
		for j := 0; j < cmdCount * 2; j++ {
			c, _, _ := buf.ReadLine()
			if j & 1 == 0 {
				continue
			}
			cmd += " " + string(c)
		}
		fmt.Println(cmd)
	}
}

/**
	SetOption
 */
func (red *Redis) SetFlag(flg []string)  {
	c := len(flg)
	if c == 0 {
		return
	}
	if c >> 1 != 1 {
		panic("Mysql参数数量不正确!")
	}
	for i:=0;i<c;i=i+2 {
		key := flg[i]
		val := flg[i+1]

		switch key {
		case CmdPort:
			port, err := strconv.Atoi(val);
			redis.port = port
			if err != nil {
				panic("端口数不正确")
			}
			if port < 0 || port > 65535 {
				panic("参数不正确: 端口范围(0-65535)")
			}
			break
		default:
			panic("参数不正确")
		}
	}
}

/**
	BPFFilter
 */
func (red *Redis) BPFFilter() string {
	return "tcp and port "+strconv.Itoa(redis.port);
}

/**
	Version
 */
func (red *Redis) Version() string {
	return red.version;
}

