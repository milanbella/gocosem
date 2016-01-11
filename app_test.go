package gocosem

import (
	"container/list"
	"fmt"
	"io"
	"net"
	"testing"
)

type MockCosemServer struct {
	closed      bool
	ln          net.Listener
	connections *list.List // list of *MockCosemServerConnection
}

type MockCosemServerConnection struct {
	srv               *MockCosemServer
	closed            bool
	rwc               io.ReadWriteCloser
	logicalDevice     uint16
	applicationClient uint16
}

func (conn *MockCosemServerConnection) replyToRequest(t *testing.T, pdu []byte) {
	//TODO:
}

func (conn *MockCosemServerConnection) receiveAndReply(t *testing.T) {
	var (
		FNAME string = "MockCosemServerConnection.receiveAndReply()"
	)

	for (!conn.closed) && (!conn.srv.closed) {

		ch := make(DlmsChannel)
		ipTransportReceive(ch, conn.rwc, &conn.applicationClient, &conn.logicalDevice)
		msg := <-ch
		if nil != msg.err {
			t.Fatalf(fmt.Sprintf("%v\n", msg.err))
		}
		m := msg.data.(map[string]interface{})
		if nil == m["pdu"] {
			panic("assertion failed")
		}

		go conn.replyToRequest(t, m["pdu"].([]byte))
	}
	t.Logf("%s: mock server: closing client connection", FNAME)
	conn.rwc.Close()
}

func (srv *MockCosemServer) acceptApp(t *testing.T, rwc io.ReadWriteCloser, aare []byte) {
	var (
		FNAME string = "MockCosemServer.acceptApp()"
	)

	t.Logf("%s: mock server waiting for client to connect", FNAME)

	// receive aarq
	ch := make(DlmsChannel)
	ipTransportReceive(ch, rwc, nil, nil)
	msg := <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%v\n", msg.err))
	}
	m := msg.data.(map[string]interface{})

	logicalDevice := m["dstWport"].(uint16)
	applicationClient := m["srcWport"].(uint16)

	// reply with aare
	ipTransportSend(ch, rwc, logicalDevice, applicationClient, aare)
	msg = <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%v\n", msg.err))
	}

	conn := new(MockCosemServerConnection)
	conn.srv = srv
	conn.rwc = rwc
	conn.logicalDevice = logicalDevice
	conn.applicationClient = applicationClient
	srv.connections.PushBack(conn)

	go conn.receiveAndReply(t)
}

func (srv *MockCosemServer) Close(t *testing.T) {
	var (
		FNAME string = "MockCosemServer.Close()"
	)
	t.Logf("%s: mock server closing ...", FNAME)
	for e := srv.connections.Front(); e != nil; e = e.Next() {
		sconn := e.Value.(*MockCosemServerConnection)
		if !sconn.closed {
			sconn.closed = true
			sconn.rwc.Close()
		}
	}
	srv.ln.Close()
	t.Logf("%s: mock server closed", FNAME)
}

func (srv *MockCosemServer) accept(t *testing.T, tcpAddr string, aare []byte) {
	var (
		FNAME string = "MockCosemServer.accept()"
	)

	ln, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		t.Fatalf(fmt.Sprintf("net.Listen() failed: %v\n", err))
	}
	srv.ln = ln

	t.Logf("%s: mock server bound to %s", FNAME, tcpAddr)

	for !srv.closed {
		conn, err := srv.ln.Accept()
		if err != nil {
			t.Fatalf(fmt.Sprintf("net.Accept() failed: %v\n", err))
		}
		go srv.acceptApp(t, conn, aare)
	}

}

func StartMockCosemServer(t *testing.T, addr string, port int, aare []byte) (srv *MockCosemServer) {

	tcpAddr := fmt.Sprintf("%s:%d", addr, port)

	srv = new(MockCosemServer)
	srv.connections = list.New()
	go srv.accept(t, tcpAddr, aare)
	return srv
}

const c_TEST_ADDR = "localhost"
const c_TEST_PORT = 4059

var c_TEST_AARE = []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x18, 0x1F, 0x08, 0x00, 0x00, 0x07}

func TestX__TcpConnect(t *testing.T) {

	srv := StartMockCosemServer(t, c_TEST_ADDR, c_TEST_PORT, c_TEST_AARE)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.err))
	}
	t.Logf("transport connected")
	dconn := msg.data.(*DlmsConn)
	dconn.Close()
	srv.Close(t)
}

func TestX_AppConnect(t *testing.T) {

	srv := StartMockCosemServer(t, c_TEST_ADDR, c_TEST_PORT, c_TEST_AARE)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.err))
	}
	t.Logf("transport connected")
	dconn := msg.data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.err))
	}
	t.Logf("application connected")
	aconn := msg.data.(*AppConn)
	aconn.Close()
	srv.Close(t)
}
