package gocosem

import (
	"bytes"
	"io"
	"net"
	"os"
	"strings"
	"testing"
)

var hdlcTestServerSockName string = "./hdlcTestServer.sock"
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
		os.Remove(hdlcTestServerSockName)
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

func TestX__pppfcs16(t *testing.T) {
	var buf bytes.Buffer
	//var b []byte = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
	var b []byte = []byte{0xA0, 0x11, 0x00, 0x02, 0x00, 0x07, 0x03, 0x10}
	var fcs16 uint16

	_, err := buf.Write(b)
	if nil != err {
		t.Fatalf("%v", err)
	}
	fcs16 = pppfcs16(PPPINITFCS16, b)

	p := make([]byte, 1)

	fcs16 ^= 0xFFFF
	p[0] = byte(fcs16 & 0x00FF)
	t.Logf("fcs16: %02X", p[0])
	_, err = buf.Write(p)
	if nil != err {
		t.Fatalf("%v", err)
	}

	p[0] = byte(fcs16 & 0xFF00 >> 8)
	t.Logf("fcs16: %02X", p[0])
	_, err = buf.Write(p)
	if nil != err {
		t.Fatalf("%v", err)
	}

	fcs16 = pppfcs16(PPPINITFCS16, buf.Bytes())
	if PPPGOODFCS16 != fcs16 {
		t.Fatalf("wrong checksum")
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
		t.Fatalf("%v", err)
	}
	client.SendDISC()
}

func TestX__WriteRead(t *testing.T) {
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
		t.Fatalf("%v", err)
	}
	defer client.SendDISC()

	bc := []byte{1, 2, 3, 4, 5}
	n, err := client.Write(bc)
	if nil != err {
		t.Fatalf("%v", err)
	}
	if n != len(bc) {
		t.Fatalf("bad length", err)
	}

	bs := make([]byte, 5)
	ch := make(chan bool)
	go func(ch chan bool) {
		n, err = server.Read(bs)
		ch <- true
	}(ch)
	<-ch
	if nil != err {
		t.Fatalf("%v", err)
	}
	if n != len(bs) {
		t.Fatalf("bad length", err)
	}

	if 0 != bytes.Compare(bc, bs) {
		t.Fatalf("bytes does not match")
	}

}
