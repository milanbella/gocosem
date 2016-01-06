package gocosem

import (
	"container/list"
	"fmt"
	"io"
	"net"
	"testing"
)

type tMockCosemObject struct {
	classId    tDlmsClassId
	attributes map[tDlmsAttributeId]*tDlmsData
}

type tMockCosemServer struct {
	closed      bool
	ln          net.Listener
	connections *list.List // list of *tMockCosemServerConnection
	objects     map[string]*tMockCosemObject
}

type tMockCosemServerConnection struct {
	srv               *tMockCosemServer
	closed            bool
	rwc               io.ReadWriteCloser
	logicalDevice     uint16
	applicationClient uint16
}

func (conn *tMockCosemServerConnection) replyToRequest(t *testing.T, pdu []byte) {
	var (
		FNAME string = "tMockCosemServerConnection.replyToRequest()"
	)

	if bytes.Equal(b[0:2], []byte{0xC0, 0x01}) {
		err, invokeIdAndPriority, classId, instanceId, attributeId, accessSelector, accessParameters := decode_GetRequestNormal(pdu)
		if nil != err {
			t.Fatalf(fmt.Sprintf("%v\n", msg.err))
			return
		}
		dataAccessResult, data = srv.getData(classId, instanceId, attributeId, accessSelector, accessParameters)
		var rawData []byte
		var buf bytes.Buffer
		err := binary.Write(&buf, binary.BigEndian, dataAccessResult)
		if nil != err {
			t.Fatalf(fmt.Sprintf("%v\n", msg.err))
			return
		}
		if 0 == dataAccessResult {
			_, err = buf.Write(data)
			if nil != err {
				t.Fatalf(fmt.Sprintf("%v\n", msg.err))
				return
			}
		}
		rawData = buf.Bytes()
	} else if bytes.Equal([]byte{0xC0, 0x03}) {
		err, invokeIdAndPriority, classIds, instanceIds, attributeIds, accessSelectors, accessParameters := decode_GetRequestWithList(pdu)
		if nil != err {
			t.Fatalf(fmt.Sprintf("%v\n", msg.err))
			return
		}
		count := len(classIds)
		var rawData []byte
		var buf bytes.Buffer
		err := binary.Write(&buf, binary.BigEndian, uint8(count))
		if nil != err {
			t.Fatalf(fmt.Sprintf("%v\n", msg.err))
			return
		}
		for i = 0; i < count; i += 1 {
			dataAccessResult, data := srv.getData(classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
			err := binary.Write(&buf, binary.BigEndian, dataAccessResult)
			if nil != err {
				t.Fatalf(fmt.Sprintf("%v\n", msg.err))
				return
			}
			if 0 == dataAccessResult {
				_, err = buf.Write(data)
				if nil != err {
					t.Fatalf(fmt.Sprintf("%v\n", msg.err))
					return
				}
			}
		}
		rawData = buf.Bytes()
	} else {
		panic("assertion failed")
	}
}

func (conn *tMockCosemServerConnection) receiveAndReply(t *testing.T) {
	var (
		FNAME string = "tMockCosemServerConnection.receiveAndReply()"
	)

	for (!conn.closed) && (!conn.srv.closed) {

		ch := make(DlmsChannel)
		ipTransportReceive(ch, conn.rwc, &conn.applicationClient, &conn.logicalDevice)
		msg := <-ch
		if nil != msg.err {
			t.Fatalf(fmt.Sprintf("%v\n", msg.err))
			return
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

func (srv *tMockCosemServer) acceptApp(t *testing.T, rwc io.ReadWriteCloser, aare []byte) {
	var (
		FNAME string = "tMockCosemServer.acceptApp()"
	)

	t.Logf("%s: mock server waiting for client to connect", FNAME)

	// receive aarq
	ch := make(DlmsChannel)
	ipTransportReceive(ch, rwc, nil, nil)
	msg := <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%v\n", msg.err))
		return
	}
	m := msg.data.(map[string]interface{})

	logicalDevice := m["dstWport"].(uint16)
	applicationClient := m["srcWport"].(uint16)

	// reply with aare
	ipTransportSend(ch, rwc, logicalDevice, applicationClient, aare)
	msg = <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%v\n", msg.err))
		return
	}

	conn := new(tMockCosemServerConnection)
	conn.srv = srv
	conn.rwc = rwc
	conn.logicalDevice = logicalDevice
	conn.applicationClient = applicationClient
	srv.connections.PushBack(conn)

	go conn.receiveAndReply(t)
}

func (srv *tMockCosemServer) objectKey(instanceId *tDlmsOid) string {
	return fmt.Sprintf("%d_%d_%d_%d_%d_%d_%d", instanceId[0], instanceId[1], instanceId[2], instanceId[3], instanceId[4], instanceId[5])
}

func (srv *tMockCosemServer) getData(classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	if nil == tDlmsOid {
		panic("assertion failed")
	}
	key := objectKey(instanceId)
	obj, ok = srv.objects[objectKey]
	if !ok {
		return 1, nil
	} else {
		if obj.classId == classId {
			data, ok = obj.attributes[attributeId]
			if !ok {
				return 1, nil
			}
			return 0, data
		} else {
			return 1, nil
		}
	}
}

func (srv *tMockCosemServer) Close(t *testing.T) {
	var (
		FNAME string = "tMockCosemServer.Close()"
	)
	t.Logf("%s: mock server closing ...", FNAME)
	for e := srv.connections.Front(); e != nil; e = e.Next() {
		sconn := e.Value.(*tMockCosemServerConnection)
		if !sconn.closed {
			sconn.closed = true
			sconn.rwc.Close()
		}
	}
	srv.ln.Close()
	t.Logf("%s: mock server closed", FNAME)
}

func (srv *tMockCosemServer) accept(t *testing.T, tcpAddr string, aare []byte) {
	var (
		FNAME string = "tMockCosemServer.accept()"
	)

	ln, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		t.Fatalf(fmt.Sprintf("net.Listen() failed: %v\n", err))
		return
	}
	srv.ln = ln

	t.Logf("%s: mock server bound to %s", FNAME, tcpAddr)

	for !srv.closed {
		conn, err := srv.ln.Accept()
		if err != nil {
			t.Fatalf(fmt.Sprintf("net.Accept() failed: %v\n", err))
			return
		}
		go srv.acceptApp(t, conn, aare)
	}

}

func startMockCosemServer(t *testing.T, addr string, port int, aare []byte) (srv *tMockCosemServer) {

	tcpAddr := fmt.Sprintf("%s:%d", addr, port)

	srv = new(tMockCosemServer)
	srv.connections = list.New()
	go srv.accept(t, tcpAddr, aare)
	return srv
}

const c_TEST_ADDR = "localhost"
const c_TEST_PORT = 4059

var c_TEST_AARE = []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x18, 0x1F, 0x08, 0x00, 0x00, 0x07}

func TestX__TcpConnect(t *testing.T) {

	srv := startMockCosemServer(t, c_TEST_ADDR, c_TEST_PORT, c_TEST_AARE)

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

	srv := startMockCosemServer(t, c_TEST_ADDR, c_TEST_PORT, c_TEST_AARE)

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
