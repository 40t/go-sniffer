package build

import (
	"github.com/google/gopacket"
	"io"
	"log"
	"strconv"
	"fmt"
	"os"
	"bufio"
	"net/http"
)

const (
	Port       = 80
	Version    = "0.1"
)

const (
	CmdPort    = "-p"
)

type H struct {
	port       int
	version    string
}

var hp *H

func NewInstance() *H {
	if hp == nil {
		hp = &H{
			port   :Port,
			version:Version,
		}
	}
	return hp
}

func (m *H) ResolveStream(net, transport gopacket.Flow, buf io.Reader) {

	bio := bufio.NewReader(buf)
	for {
		req, err := http.ReadRequest(bio)

		if err == io.EOF {
			return
		} else if err != nil {
			continue
		} else {

			var msg = "["
			msg += req.Method
			msg += "] ["
			msg += req.Host + req.URL.String()
			msg += "] ["
			req.ParseForm()
			msg += req.Form.Encode()
			msg += "]"

			log.Println(msg)

			req.Body.Close()
		}
	}
}

func (m *H) BPFFilter() string {
	return "tcp and port "+strconv.Itoa(m.port);
}

func (m *H) Version() string {
	return Version
}

func (m *H) SetFlag(flg []string)  {

	c := len(flg)

	if c == 0 {
		return
	}
	if c >> 1 == 0 {
		fmt.Println("ERR : Http Number of parameters")
		os.Exit(1)
	}
	for i:=0;i<c;i=i+2 {
		key := flg[i]
		val := flg[i+1]

		switch key {
		case CmdPort:
			port, err := strconv.Atoi(val);
			m.port = port
			if err != nil {
				panic("ERR : port")
			}
			if port < 0 || port > 65535 {
				panic("ERR : port(0-65535)")
			}
			break
		default:
			panic("ERR : mysql's params")
		}
	}
}