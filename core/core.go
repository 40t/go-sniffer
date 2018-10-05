package core

type Core struct{
	Version string
}

var cxt Core

func New() Core {

	cxt.Version = "0.1"

	return cxt
}

func (c *Core) Run()  {

	//new plugin
	plug := NewPlug()

	//parse commend
	cmd := NewCmd(plug)
	cmd.Run()

	//dispatch
	NewDispatch(plug, cmd).Capture()
}