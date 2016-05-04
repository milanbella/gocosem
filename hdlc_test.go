package gocosem

import (
	"bytes"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

var hdlcTestServerSockName string = "./hdlcTestServer.sock"
var hdlcTestServer net.Listener

func generateBytes(len int) []byte {
	b := make([]byte, len)
	for i := 0; i < len; i++ {
		b[i] = byte(i)
	}
	return b
}

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

func TestX__hdlc_hdlcPipe(t *testing.T) {
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

func TestX__hdlc_pppfcs16(t *testing.T) {
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

func TestX__hdlc_SendSNRM(t *testing.T) {
	hdlcTestInit(t)

	crw, srw := createHdlcPipe(t)
	defer crw.Close()
	defer srw.Close()

	clientId := uint8(1)
	logicalDeviceId := uint16(2)
	physicalDeviceId := new(uint16)
	*physicalDeviceId = 3

	client := NewHdlcTransport(crw, true, clientId, logicalDeviceId, physicalDeviceId)
	defer func() {
		client.Close()
	}()
	server := NewHdlcTransport(srw, false, clientId, logicalDeviceId, physicalDeviceId)
	defer func() {
		server.Close()
	}()

	err := client.SendSNRM(nil, nil)
	if nil != err {
		t.Fatalf("%v", err)
	}
	client.SendDISC()
}

func TestX__hdlc_WriteRead(t *testing.T) {
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

	err := client.SendSNRM(nil, nil)
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
		n, err := server.Read(bs)
		if nil != err {
			t.Fatalf("%v", err)
		}
		if n != len(bs) {
			t.Fatalf("bad length", err)
		}
		if 0 != bytes.Compare(bc, bs) {
			t.Fatalf("bytes does not match")
		}
		ch <- true
	}(ch)
	<-ch
}

func TestX__hdlc_WriteRead_i50(t *testing.T) {
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

	maxInfoFieldLengthTransmit := uint8(50)
	maxInfoFieldLengthReceive := uint8(50)

	err := client.SendSNRM(&maxInfoFieldLengthTransmit, &maxInfoFieldLengthReceive)
	if nil != err {
		t.Fatalf("%v", err)
	}
	defer client.SendDISC()

	bc := generateBytes(2500)
	n, err := client.Write(bc)
	if nil != err {
		t.Fatalf("%v", err)
	}
	if n != len(bc) {
		t.Fatalf("bad length", err)
	}

	bs := make([]byte, len(bc))
	ch := make(chan bool)
	go func(ch chan bool) {
		n, err := server.Read(bs)
		if nil != err {
			ch <- true
			t.Fatalf("%v", err)
			return
		}
		if n != len(bs) {
			ch <- true
			t.Fatalf("bad length", err)
			return
		}
		if 0 != bytes.Compare(bc, bs) {
			ch <- true
			t.Fatalf("bytes does not match")
			return
		}
		ch <- true
	}(ch)
	<-ch

}

func TestX__hdlc_WriteRead_parallel_transmit(t *testing.T) {
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

	err := client.SendSNRM(nil, nil)
	if nil != err {
		t.Fatalf("%v", err)
	}
	defer client.SendDISC()

	bt := []byte{1, 2, 3, 4, 5}

	// both server and client transmit at the same time

	ch := make(chan string)
	chf := make(chan string)

	go func() {
		<-ch
		n, err := client.Write(bt)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		chf <- "1"
	}()

	go func() {
		<-ch
		n, err := server.Write(bt)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		chf <- "2"
	}()

	go func() {
		<-ch
		br := make([]byte, len(bt))
		n, err := client.Read(br)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		if 0 != bytes.Compare(bt, br) {
			close(chf)
			t.Fatalf("bytes does not match")
			return
		}
		chf <- "3"
	}()

	go func() {
		<-ch
		br := make([]byte, len(bt))
		n, err := server.Read(br)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length")
			return
		}
		if 0 != bytes.Compare(bt, br) {
			close(chf)
			t.Fatalf("bytes does not match")
			return
		}
		chf <- "4"
	}()

	close(ch)

	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)

}

func TestX__hdlc_WriteRead_i50_w1_parallel_transmit(t *testing.T) {
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

	maxInfoFieldLengthTransmit := uint8(50)
	maxInfoFieldLengthReceive := uint8(50)

	err := client.SendSNRM(&maxInfoFieldLengthTransmit, &maxInfoFieldLengthReceive)
	if nil != err {
		t.Fatalf("%v", err)
	}
	defer client.SendDISC()

	bt := generateBytes(2500)

	// both server and client transmit at the same time

	ch := make(chan string)
	chf := make(chan string)

	go func() {
		<-ch
		n, err := client.Write(bt)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		chf <- "1"
	}()

	go func() {
		<-ch
		n, err := server.Write(bt)
		if nil != err {
			close(chf)
			return
			t.Fatalf("%v", err)
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		chf <- "2"
	}()

	go func() {
		<-ch
		br := make([]byte, len(bt))
		n, err := client.Read(br)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		if 0 != bytes.Compare(bt, br) {
			close(chf)
			t.Fatalf("bytes does not match")
			return
		}
		chf <- "3"
	}()

	go func() {
		<-ch
		br := make([]byte, len(bt))
		n, err := server.Read(br)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		if 0 != bytes.Compare(bt, br) {
			close(chf)
			t.Fatalf("bytes does not match")
			return
		}
		chf <- "4"
	}()

	close(ch)

	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)

}

func TestX__hdlc_WriteRead_i22_w1_parallel_transmit(t *testing.T) {
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

	maxInfoFieldLengthTransmit := uint8(22)
	maxInfoFieldLengthReceive := uint8(22)

	err := client.SendSNRM(&maxInfoFieldLengthTransmit, &maxInfoFieldLengthReceive)
	if nil != err {
		t.Fatalf("%v", err)
	}
	defer client.SendDISC()

	bt := generateBytes(2500)

	// both server and client transmit at the same time

	ch := make(chan string)
	chf := make(chan string)

	go func() {
		<-ch
		n, err := client.Write(bt)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		chf <- "1"
	}()

	go func() {
		<-ch
		n, err := server.Write(bt)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		chf <- "2"
	}()

	go func() {
		<-ch
		br := make([]byte, len(bt))
		n, err := client.Read(br)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		if 0 != bytes.Compare(bt, br) {
			close(chf)
			t.Fatalf("bytes does not match")
			return
		}
		chf <- "3"
	}()

	go func() {
		<-ch
		br := make([]byte, len(bt))
		n, err := server.Read(br)
		if nil != err {
			t.Fatalf("%v", err)
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length", err)
			return
		}
		if 0 != bytes.Compare(bt, br) {
			close(chf)
			t.Fatalf("bytes does not match")
			return
		}
		chf <- "4"
	}()

	close(ch)

	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)
}

func TestX__hdlc_WriteRead_i22_w3_parallel_transmit_drop_every_5th_frame(t *testing.T) {
	hdlcTestInit(t)

	crw, srw := createHdlcPipe(t)
	defer crw.Close()
	defer srw.Close()

	clientId := uint8(1)
	logicalDeviceId := uint16(2)
	physicalDeviceId := new(uint16)
	*physicalDeviceId = 3

	client := NewHdlcTransport(crw, true, clientId, logicalDeviceId, physicalDeviceId)
	client.readFrameImpl = 1 // this read frame implementation drops every second frame
	client.responseTimeout = time.Duration(1) * time.Millisecond
	defer client.Close()
	server := NewHdlcTransport(srw, false, clientId, logicalDeviceId, physicalDeviceId)
	server.readFrameImpl = 1 // this read frame implementation drops every second frame
	defer server.Close()

	maxInfoFieldLengthTransmit := uint8(22)
	maxInfoFieldLengthReceive := uint8(22)

	err := client.SendSNRM(&maxInfoFieldLengthTransmit, &maxInfoFieldLengthReceive)
	if nil != err {
		t.Fatalf("%v", err)
	}
	defer client.SendDISC()

	bt := generateBytes(2500)

	// both server and client transmit at the same time

	ch := make(chan string)
	chf := make(chan string)

	go func() {
		<-ch
		n, err := client.Write(bt)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length")
			return
		}
		chf <- "1"
	}()

	go func() {
		<-ch
		n, err := server.Write(bt)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length")
			return
		}
		chf <- "2"
	}()

	go func() {
		<-ch
		br := make([]byte, len(bt))
		n, err := client.Read(br)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length")
			return
		}
		if 0 != bytes.Compare(bt, br) {
			close(chf)
			t.Fatalf("bytes does not match")
			return
		}
		chf <- "3"
	}()

	go func() {
		<-ch
		br := make([]byte, len(bt))
		n, err := server.Read(br)
		if nil != err {
			close(chf)
			t.Fatalf("%v", err)
			return
		}
		if n != len(bt) {
			close(chf)
			t.Fatalf("bad length")
			return
		}
		if 0 != bytes.Compare(bt, br) {
			close(chf)
			t.Fatalf("bytes does not match")
			return
		}
		chf <- "4"
	}()

	close(ch)

	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)
	t.Logf("%s\n", <-chf)

}
