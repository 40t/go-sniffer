package core

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"log"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"time"
)

type Dispatch struct {
	device string
	payload []byte
	Plug *Plug
}

func NewDispatch(plug *Plug, cmd *Cmd) *Dispatch {
	return &Dispatch {
		Plug: plug,
		device:cmd.Device,
	}
}

func (d *Dispatch) Capture() {

	// Init device
	handle, err := pcap.OpenLive(d.device, 65535, false, pcap.BlockForever)
	if err != nil {
		return
	}

	// Set filter
	fmt.Println(d.Plug.BPF)
	err = handle.SetBPFFilter(d.Plug.BPF)
	if err != nil {
		log.Fatal(err)
	}

	// Capture
	src     := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := src.Packets()

	// Set up assembly
	streamFactory := &ProtocolStreamFactory{
		dispatch:d,
	}
	streamPool := NewStreamPool(streamFactory)
	assembler  := NewAssembler(streamPool)
	ticker     := time.Tick(time.Minute)

	// Loop until ctrl+z
	for {
		select {
		case packet := <-packets:
			if packet.NetworkLayer() == nil ||
				packet.TransportLayer() == nil ||
				packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				fmt.Println("包不能解析")
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(
				packet.NetworkLayer().NetworkFlow(),
				tcp, packet.Metadata().Timestamp,
			)
		case <-ticker:
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}

type ProtocolStreamFactory struct {
	dispatch *Dispatch
}

type ProtocolStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}

func (m *ProtocolStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {

	//init stream struct
	stm := &ProtocolStream {
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
	}

	//new stream
	fmt.Println("# 新连接:", net, transport)

	//decode packet
	go m.dispatch.Plug.ResolveStream(net, transport, &(stm.r))

	return &(stm.r)
}