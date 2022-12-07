package parse

import (
	"github.com/google/gopacket"
	"go-sniffer/pkg/model"
)

func GenSession(net, transport gopacket.Flow) *model.MysqlSession {
	clientIp := net.Src().String()
	sourceIp := net.Dst().String()
	clientPort := transport.Src().String()
	serverPort := transport.Dst().String()

	return &model.MysqlSession{
		ClientIP:   clientIp,
		ClientPort: clientPort,
		ServerIP:   sourceIp,
		ServerPort: serverPort,
	}
}
