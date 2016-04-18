package gocosem

import (
	"net"
	"testing"
)

func createHdlcPipe() (net.Conn, net.Conn) {
	return net.Pipe()
}

func TestX__SendSNRM(t *testing.T) {
	crw, srw := createHdlcPipe()
	defer crw.Close()
	defer srw.Close()

	client := NewHdlcTransport(crw)
	defer client.Close()
	server := NewHdlcTransport(srw)
	defer server.Close()

	err := client.SendSNRM(nil, nil, nil, nil)
	if nil != err {
		t.Fatal("%v", err)
	}
	client.SendDISC()
}
