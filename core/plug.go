package core

import (
	"io/ioutil"
	"plugin"
	"github.com/google/gopacket"
	"io"
	mysql "github.com/40t/go-sniffer/plugSrc/mysql/build"
	redis "github.com/40t/go-sniffer/plugSrc/redis/build"
	hp "github.com/40t/go-sniffer/plugSrc/http/build"
	"path/filepath"
	"fmt"
	"path"
)

type Plug struct {

	//当前插件路径
	dir string
	//解析包
	ResolveStream func(net gopacket.Flow, transport gopacket.Flow, r io.Reader)
	//BPF
	BPF string

	//内部插件列表
	InternalPlugList map[string]PlugInterface
	//外部插件列表
	ExternalPlugList map[string]ExternalPlug
}

// 内部插件必须实现此接口
// ResolvePacket - 包入口
// BPFFilter     - 设置BPF规则,例如mysql: (tcp and port 3306)
// SetFlag       - 设置参数
// Version       - 返回插件版本,例如0.1.0
type PlugInterface interface {
	//解析流
	ResolveStream(net gopacket.Flow, transport gopacket.Flow, r io.Reader)
	//BPF
	BPFFilter() string
	//设置插件需要的参数
	SetFlag([]string)
	//获取版本
	Version() string
}

//外部插件
type ExternalPlug struct {
	Name          string
	Version       string
	ResolvePacket func(net gopacket.Flow, transport gopacket.Flow, r io.Reader)
	BPFFilter     func() string
	SetFlag       func([]string)
}

//实例化
func NewPlug() *Plug {

	var p Plug

	//设置默认插件目录
	p.dir, _ = filepath.Abs( "./plug/")

	//加载内部插件
	p.LoadInternalPlugList()

	//加载外部插件
	p.LoadExternalPlugList()

	return &p
}

//加载内部插件
func (p *Plug) LoadInternalPlugList() {

	list := make(map[string]PlugInterface)

	//Mysql
	list["mysql"]   = mysql.NewInstance()

	//TODO Mongodb

	//TODO ARP

	//Redis
	list["redis"]   = redis.NewInstance()
	//Http
	list["http"]    = hp.NewInstance()

	p.InternalPlugList = list
}

//加载外部so后缀插件
func (p *Plug) LoadExternalPlugList() {

	dir, err := ioutil.ReadDir(p.dir)
	if err != nil {
		panic(p.dir + "不存在，或者无权访问")
	}

	p.ExternalPlugList = make(map[string]ExternalPlug)
	for _, fi := range dir {
		if fi.IsDir() || path.Ext(fi.Name()) != ".so" {
			continue
		}

		plug, err := plugin.Open(p.dir+"/"+fi.Name())
		if err != nil {
			panic(err)
		}

		versionFunc, err := plug.Lookup("Version")
		if err != nil {
			panic(err)
		}

		setFlagFunc, err := plug.Lookup("SetFlag")
		if err != nil {
			panic(err)
		}

		BPFFilterFunc, err := plug.Lookup("BPFFilter")
		if err != nil {
			panic(err)
		}

		ResolvePacketFunc, err := plug.Lookup("ResolvePacket")
		if err != nil {
			panic(err)
		}

		version := versionFunc.(func() string)()
		p.ExternalPlugList[fi.Name()] = ExternalPlug {
			ResolvePacket:ResolvePacketFunc.(func(net gopacket.Flow, transport gopacket.Flow, r io.Reader)),
			SetFlag:setFlagFunc.(func([]string)),
			BPFFilter:BPFFilterFunc.(func() string),
			Version:version,
			Name:fi.Name(),
		}
	}
}

//改变插件地址
func (p *Plug) ChangePath(dir string) {
	p.dir = dir
}

//打印插件列表
func (p *Plug) PrintList() {

	//Print Internal Plug
	for inPlugName, _ := range p.InternalPlugList {
		fmt.Println("内部插件:"+inPlugName)
	}

	//split
	fmt.Println("-- --- --")

	//print External Plug
	for exPlugName, _ := range p.ExternalPlugList {
		fmt.Println("外部插件:"+exPlugName)
	}
}

//选择当前使用的插件 && 加载插件
func (p *Plug) SetOption(plugName string, plugParams []string) {

	//Load Internal Plug
	if internalPlug, ok := p.InternalPlugList[plugName]; ok {

		p.ResolveStream = internalPlug.ResolveStream
		internalPlug.SetFlag(plugParams)
		p.BPF =  internalPlug.BPFFilter()

		return
	}

	//Load External Plug
	plug, err := plugin.Open("./plug/"+ plugName)
	if err != nil {
		panic(err)
	}
	resolvePacket, err := plug.Lookup("ResolvePacket")
	if err != nil {
		panic(err)
	}
	setFlag, err := plug.Lookup("SetFlag")
	if err != nil {
		panic(err)
	}
	BPFFilter, err := plug.Lookup("BPFFilter")
	if err != nil {
		panic(err)
	}
	p.ResolveStream = resolvePacket.(func(net gopacket.Flow, transport gopacket.Flow, r io.Reader))
	setFlag.(func([]string))(plugParams)
	p.BPF = BPFFilter.(func()string)()
}