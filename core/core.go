package core

type Core struct{
	//版本信息
	Version string
}

var cxt Core

func New() Core {

	cxt.Version = "0.1"

	return cxt
}

func (c *Core) Run()  {

	//插件
	plug := NewPlug()

	//解析参数
	cmd := NewCmd(plug)
	cmd.Run()

	//开启抓包
	NewDispatch(plug, cmd).Capture()
}