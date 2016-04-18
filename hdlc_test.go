package gocosem

import (
	"io"
	"net"
	"strings"
	"testing"
)

var hdlcTestServerSockName string = "/tmp/hdlcTestServer.sock"
var hdlcTestServer net.Listener

func createHdlcPipe(t *testing.T) (conn1 net.Conn, conn2 net.Conn) {
	conn1, err := net.Dial("unixpacket", hdlcTestServerSockName)
	if nil != err {
		t.Fatalf("%v", err)
	}
	conn2, err = hdlcTestServer.Accept()
	if nil != err {
		t.Fatalf("%v", err)
	}
	return conn1, conn2

}

func hdlcTestInit(t *testing.T) {
	if nil == hdlcTestServer {
		ln, err := net.Listen("unixpacket", hdlcTestServerSockName)
		if err != nil {
			t.Fatalf("%v", err)
		}
		hdlcTestServer = ln
	}
}

func TestX__hdlcPipe(t *testing.T) {
	hdlcTestInit(t)

	crw, srw := createHdlcPipe(t)
	defer crw.Close()
	defer srw.Close()

	msg := "12345"
	_, err := crw.Write([]byte("12345"))
	if nil != err {
		t.Fatalf("%v", err)
	}

	p := make([]byte, len(msg))
	_, err = io.ReadFull(srw, p)
	if nil != err {
		t.Fatal("%v", err)
	}

	if 0 != strings.Compare(msg, string(p)) {
		t.Fatalf("no match")
	}

}

func TestX__SendSNRM(t *testing.T) {
	hdlcTestInit(t)

	crw, srw := createHdlcPipe(t)
	defer crw.Close()
	defer srw.Close()

	clientId := uint8(1)
	logicalDeviceId := uint16(2)
	physicalDeviceId := new(uint16)
	*physicalDeviceId = 3

	client := NewHdlcTransport(crw, true, clientId, logicalDeviceId, physicalDeviceId)
	defer client.Close()
	server := NewHdlcTransport(srw, false, clientId, logicalDeviceId, physicalDeviceId)
	defer server.Close()

	err := client.SendSNRM(nil, nil, nil, nil)
	if nil != err {
		t.Fatal("%v", err)
	}
	client.SendDISC()
}
