package core

import (
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"plugin"

	hp "github.com/40t/go-sniffer/plugSrc/http/build"
	mongodb "github.com/40t/go-sniffer/plugSrc/mongodb/build"
	redis "github.com/40t/go-sniffer/plugSrc/redis/build"
	mssql "github.com/feiin/go-sniffer/plugSrc/mssql/build"
	mysql "github.com/feiin/go-sniffer/plugSrc/mysql/build"
	"github.com/google/gopacket"
)

type Plug struct {
	dir           string
	ResolveStream func(net gopacket.Flow, transport gopacket.Flow, r io.Reader)
	BPF           string

	InternalPlugList map[string]PlugInterface
	ExternalPlugList map[string]ExternalPlug
}

// All internal plug-ins must implement this interface
// ResolvePacket - entry
// BPFFilter     - set BPF, like: mysql(tcp and port 3306)
// SetFlag       - plug-in params
// Version       - plug-in version
type PlugInterface interface {
	ResolveStream(net gopacket.Flow, transport gopacket.Flow, r io.Reader)
	BPFFilter() string
	SetFlag([]string)
	Version() string
}

type ExternalPlug struct {
	Name          string
	Version       string
	ResolvePacket func(net gopacket.Flow, transport gopacket.Flow, r io.Reader)
	BPFFilter     func() string
	SetFlag       func([]string)
}

func NewPlug() *Plug {

	var p Plug

	p.dir, _ = filepath.Abs("./plug/")
	p.LoadInternalPlugList()
	p.LoadExternalPlugList()

	return &p
}

func (p *Plug) LoadInternalPlugList() {

	list := make(map[string]PlugInterface)

	//Mysql
	list["mysql"] = mysql.NewInstance()

	//Mongodb
	list["mongodb"] = mongodb.NewInstance()

	//Redis
	list["redis"] = redis.NewInstance()

	//Http
	list["http"] = hp.NewInstance()

	list["mssql"] = mssql.NewInstance()

	p.InternalPlugList = list
}

func (p *Plug) LoadExternalPlugList() {

	dir, err := ioutil.ReadDir(p.dir)
	if err != nil {
		return
	}

	p.ExternalPlugList = make(map[string]ExternalPlug)
	for _, fi := range dir {
		if fi.IsDir() || path.Ext(fi.Name()) != ".so" {
			continue
		}

		plug, err := plugin.Open(p.dir + "/" + fi.Name())
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
		p.ExternalPlugList[fi.Name()] = ExternalPlug{
			ResolvePacket: ResolvePacketFunc.(func(net gopacket.Flow, transport gopacket.Flow, r io.Reader)),
			SetFlag:       setFlagFunc.(func([]string)),
			BPFFilter:     BPFFilterFunc.(func() string),
			Version:       version,
			Name:          fi.Name(),
		}
	}
}

func (p *Plug) ChangePath(dir string) {
	p.dir = dir
}

func (p *Plug) PrintList() {

	//Print Internal Plug
	for inPlugName, _ := range p.InternalPlugList {
		fmt.Println("internal plug : " + inPlugName)
	}

	//split
	fmt.Println("-- --- --")

	//print External Plug
	for exPlugName, _ := range p.ExternalPlugList {
		fmt.Println("external plug : " + exPlugName)
	}
}

func (p *Plug) SetOption(plugName string, plugParams []string) {

	fmt.Println("internalPlug", plugName)

	//Load Internal Plug
	if internalPlug, ok := p.InternalPlugList[plugName]; ok {

		p.ResolveStream = internalPlug.ResolveStream
		internalPlug.SetFlag(plugParams)
		p.BPF = internalPlug.BPFFilter()

		return
	}

	//Load External Plug
	plug, err := plugin.Open("./plug/" + plugName)
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
	p.BPF = BPFFilter.(func() string)()
}
