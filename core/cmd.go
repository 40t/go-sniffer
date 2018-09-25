package core

import (
	"os"
	"strings"
	"fmt"
	"net"
	"strconv"
)

const InternalCmdPrefix = "--"
const (
	InternalCmdHelp = "help"  //帮助文档
	InternalCmdEnv  = "env"   //环境变量
	InternalCmdList = "list"  //插件列表
	InternalCmdVer  = "ver"   //版本信息
	InternalDevice  = "dev"   //设备链表
)

type Cmd struct {
	Device string
	plugHandle *Plug
}

func NewCmd(p *Plug) *Cmd {

	return &Cmd{
		plugHandle:p,
	}
}

//start
func (cm *Cmd) Run() {

	//使用帮助
	if len(os.Args) <= 1 {
		cm.printHelpMessage();
		os.Exit(1)
	}

	//解析命令
	firstArg := string(os.Args[1])
	if strings.HasPrefix(firstArg, InternalCmdPrefix) {
		cm.parseInternalCmd()
	} else {
		cm.parsePlugCmd()
	}
}

//解析内部参数
func (cm *Cmd) parseInternalCmd() {

	arg := string(os.Args[1])
	cmd := strings.Trim(arg, InternalCmdPrefix)

	switch cmd {
		case InternalCmdHelp:
			cm.printHelpMessage()
			break;
		case InternalCmdEnv:
			fmt.Println("插件路径:"+cm.plugHandle.dir)
			break
		case InternalCmdList:
			cm.plugHandle.PrintList()
			break
		case InternalCmdVer:
			fmt.Println(cxt.Version)
			break
		case InternalDevice:
			cm.printDevice()
			break;
	}
	os.Exit(1)
}

//使用说明
func (cm *Cmd) printHelpMessage()  {

	fmt.Println("==================================================================================")
	fmt.Println("[使用说明]")
	fmt.Println("")
	fmt.Println("    go-sniffer [设备名] [插件名] [插件参数(可选)]")
	fmt.Println()
	fmt.Println("    [例子]")
	fmt.Println("          go-sniffer en0 redis          抓取redis数据包")
	fmt.Println("          go-sniffer en0 mysql -p 3306  抓取mysql数据包,端口3306")
	fmt.Println()
	fmt.Println("    go-sniffer --[命令]")
	fmt.Println("               --help 帮助信息")
	fmt.Println("               --env  环境变量")
	fmt.Println("               --list 插件列表")
	fmt.Println("               --ver  版本信息")
	fmt.Println("               --dev  设备列表")
	fmt.Println("    [例子]")
	fmt.Println("          go-sniffer --list 查看可抓取的协议")
	fmt.Println()
	fmt.Println("==================================================================================")
	cm.printDevice()
	fmt.Println("==================================================================================")
}

//打印插件
func (cm *Cmd) printPlugList() {
	l := len(cm.plugHandle.InternalPlugList)
	l += len(cm.plugHandle.ExternalPlugList)
	fmt.Println("#    插件数量："+strconv.Itoa(l))
}

//打印设备
func (cm *Cmd) printDevice() {
	ifaces, err:= net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _,a:=range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					fmt.Println("[设备名] : "+iface.Name+" : "+iface.HardwareAddr.String()+"  "+ip4.String())
				}
			}
		}
	}
}

//解析插件需要的参数
func (cm *Cmd) parsePlugCmd()  {

	if len(os.Args) < 3 {
		fmt.Println("缺少[插件名]")
		fmt.Println("go-sniffer [设备名] [插件名] [插件参数(可选)]")
		os.Exit(1)
	}

	cm.Device  = os.Args[1]
	plugName  := os.Args[2]
	plugParams:= os.Args[3:]
	cm.plugHandle.SetOption(plugName, plugParams)
}




