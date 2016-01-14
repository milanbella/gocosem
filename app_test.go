package gocosem

import (
	"bytes"
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
	blocks            map[uint8][][]byte // invoke id bolocks to be sent in case of block transfer
}

func (conn *tMockCosemServerConnection) sendEncodedReply(t *testing.T, invokeIdAndPriority tDlmsInvokeIdAndPriority, reply []byte) {
	var FNAME string = "tMockCosemServerConnection.sendEncodedReply()"
	t.Logf("%s: reply: %02X", FNAME, reply) //@@@@@@@@@@@@@

	invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)
	l := 10 // block length
	if len(reply) > l {
		// use block transfer
		t.Logf("%s: using block transfer", FNAME)

		blocks := make([][]byte, len(reply)/10+1)
		b := reply[0:]
		var i int
		for i = 0; len(b) > l; i += 1 {
			blocks[i] = b[0:l]
			b = b[l:]
		}
		blocks[i] = b
		blocks = blocks[0 : i+1] // truncate sicnce we may have allocated more
		conn.blocks[invokeId] = blocks

		t.Logf("%s: blocks count: %d", FNAME, len(blocks))
		for i = 0; i < len(blocks); i += 1 {
			t.Logf("%s: block[%d]: %02X", FNAME, i, blocks[i]) //@@@@@@@@@@@@@@@@
		}

		err, reply := encode_GetResponsewithDataBlock(invokeIdAndPriority, false, 1, 0, blocks[0])
		if nil != err {
			t.Fatalf("%v\n", err)
			return
		}
		ch := make(DlmsChannel)
		ipTransportSend(ch, conn.rwc, conn.logicalDevice, conn.applicationClient, reply)
		msg := <-ch
		if nil != msg.err {
			t.Fatalf("%v\n", err)
			return
		}

	} else {
		t.Logf("%s: using normal transfer", FNAME)
		ch := make(DlmsChannel)
		ipTransportSend(ch, conn.rwc, conn.logicalDevice, conn.applicationClient, reply)
		msg := <-ch
		if nil != msg.err {
			t.Fatalf("%v\n", msg.err)
			return
		}
	}
}

func (conn *tMockCosemServerConnection) replyToRequest(t *testing.T, pdu []byte) {
	var FNAME string = "tMockCosemServerConnection.replyToRequest()"

	if bytes.Equal(pdu[0:2], []byte{0xC0, 0x01}) {
		t.Logf("%s: GetRequestNormal", FNAME)
		err, invokeIdAndPriority, classId, instanceId, attributeId, accessSelector, accessParameters := decode_GetRequestNormal(pdu)
		if nil != err {
			t.Fatalf("%v\n", err)
			return
		}
		dataAccessResult, data := conn.srv.getData(t, classId, instanceId, attributeId, accessSelector, accessParameters)
		t.Logf("%s: dataAccessResult: %d", FNAME, dataAccessResult)
		err, rawData := encode_GetResponseNormal(invokeIdAndPriority, dataAccessResult, data)
		if nil != err {
			t.Fatalf("%v\n", err)
			return
		}
		conn.sendEncodedReply(t, invokeIdAndPriority, rawData)
	} else if bytes.Equal(pdu[0:2], []byte{0xC0, 0x03}) {
		t.Logf("%s: GetRequestWithList", FNAME)
		err, invokeIdAndPriority, classIds, instanceIds, attributeIds, accessSelectors, accessParameters := decode_GetRequestWithList(pdu)
		if nil != err {
			t.Fatalf("%v\n", err)
			return
		}
		count := len(classIds)
		var rawData []byte
		datas := make([]*tDlmsData, count)
		dataAccessResults := make([]tDlmsDataAccessResult, count)
		for i := 0; i < count; i += 1 {
			dataAccessResult, data := conn.srv.getData(t, classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
			t.Logf("%s: dataAccessResult[%d]: %d", FNAME, i, dataAccessResult)
			dataAccessResults[i] = dataAccessResult
			datas[i] = data
		}
		err, rawData = encode_GetResponseWithList(invokeIdAndPriority, dataAccessResults, datas)
		if nil != err {
			t.Fatalf("%v\n", err)
			return
		}
		conn.sendEncodedReply(t, invokeIdAndPriority, rawData)
	} else if bytes.Equal(pdu[0:2], []byte{0xC0, 0x02}) {
		t.Logf("%s: GetRequestForNextDataBlock", FNAME)
		err, invokeIdAndPriority, blockNumber := decode_GetRequestForNextDataBlock(pdu)
		if nil != err {
			t.Fatalf("%v\n", err)
			return
		}
		invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)

		var dataAccessResult tDlmsDataAccessResult
		var rawData []byte
		var lastBlock bool

		if nil == conn.blocks[invokeId] {
			t.Logf("no blocks for invokeId: setting dataAccessResult to 1")
			dataAccessResult = 1
			rawData = nil
		} else if int(blockNumber) >= len(conn.blocks[invokeId]) {
			t.Logf("no such block for invokeId: setting dataAccessResult to 1")
			dataAccessResult = 1
			rawData = nil
		} else {
			dataAccessResult = 0
			rawData = conn.blocks[invokeId][blockNumber]
		}
		t.Logf("%s: dataAccessResult: %d", FNAME, dataAccessResult)

		if (len(conn.blocks[invokeId]) - 1) == int(blockNumber) {
			lastBlock = true
		} else {
			lastBlock = false
		}

		if lastBlock {
			conn.blocks[invokeId] = nil
		}

		if !lastBlock {
			blockNumber += 1
		}
		err, reply := encode_GetResponsewithDataBlock(invokeIdAndPriority, lastBlock, blockNumber, dataAccessResult, rawData)
		if nil != err {
			t.Fatalf("%v\n", err)
			return
		}
		ch := make(DlmsChannel)
		ipTransportSend(ch, conn.rwc, conn.logicalDevice, conn.applicationClient, reply)
		msg := <-ch
		if nil != msg.err {
			t.Fatalf("%v\n", err)
			return
		}
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
			t.Fatalf("%v\n", msg.err)
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
		t.Fatalf("%v\n", msg.err)
		return
	}
	m := msg.data.(map[string]interface{})

	logicalDevice := m["dstWport"].(uint16)
	applicationClient := m["srcWport"].(uint16)

	// reply with aare
	ipTransportSend(ch, rwc, logicalDevice, applicationClient, aare)
	msg = <-ch
	if nil != msg.err {
		t.Fatalf("%v\n", msg.err)
		return
	}

	conn := new(tMockCosemServerConnection)
	conn.srv = srv
	conn.rwc = rwc
	conn.logicalDevice = logicalDevice
	conn.applicationClient = applicationClient
	conn.blocks = make(map[uint8][][]byte)
	srv.connections.PushBack(conn)

	go conn.receiveAndReply(t)
}

func (srv *tMockCosemServer) objectKey(instanceId *tDlmsOid) string {
	return fmt.Sprintf("%d_%d_%d_%d_%d_%d_%d", instanceId[0], instanceId[1], instanceId[2], instanceId[3], instanceId[4], instanceId[5])
}

func (srv *tMockCosemServer) getData(t *testing.T, classId tDlmsClassId, instanceId *tDlmsOid, attributeId tDlmsAttributeId, accessSelector *tDlmsAccessSelector, accessParameters *tDlmsData) (dataAccessResult tDlmsDataAccessResult, data *tDlmsData) {
	if nil == instanceId {
		panic("assertion failed")
	}
	key := srv.objectKey(instanceId)
	obj, ok := srv.objects[key]
	if !ok {
		t.Logf("no such instance id: setting dataAccessResult to 1")
		return 1, nil
	} else {
		if obj.classId == classId {
			data, ok = obj.attributes[attributeId]
			if !ok {
				t.Logf("no such instance attribute: setting dataAccessResult to 1")
				return 1, nil
			}
			return 0, data
		} else {
			t.Logf("instance class mismatch: setting dataAccessResult to 1")
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
		t.Fatalf("net.Listen() failed: %v\n", err)
		return
	}
	srv.ln = ln

	t.Logf("%s: mock server bound to %s", FNAME, tcpAddr)

	for !srv.closed {
		conn, err := srv.ln.Accept()
		if err != nil {
			t.Fatalf("net.Accept() failed: %v\n", err)
			return
		}
		go srv.acceptApp(t, conn, aare)
	}

}

func (srv *tMockCosemServer) setAttribute(instanceId *tDlmsOid, classId tDlmsClassId, attributeId tDlmsAttributeId, data *tDlmsData) {

	key := srv.objectKey(instanceId)
	obj := srv.objects[key]
	if nil == obj {
		obj = new(tMockCosemObject)
		srv.objects[key] = obj
	}
	obj.classId = classId
	attributes := obj.attributes
	if nil == attributes {
		attributes = make(map[tDlmsAttributeId]*tDlmsData)
		obj.attributes = attributes
	}
	attributes[attributeId] = data
}

func startMockCosemServer(t *testing.T, addr string, port int, aare []byte) (srv *tMockCosemServer) {

	tcpAddr := fmt.Sprintf("%s:%d", addr, port)

	srv = new(tMockCosemServer)
	srv.connections = list.New()
	srv.objects = make(map[string]*tMockCosemObject)
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
		t.Fatalf("%s\n", msg.err)
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
		t.Fatalf("%s\n", msg.err)
	}
	t.Logf("transport connected")
	dconn := msg.data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.err {
		t.Fatalf("%s\n", msg.err)
	}
	t.Logf("application connected")
	aconn := msg.data.(*AppConn)
	aconn.Close()
	srv.Close(t)
}

func TestX_GetRequestNormal(t *testing.T) {

	srv := startMockCosemServer(t, c_TEST_ADDR, c_TEST_PORT, c_TEST_AARE)

	data := (new(tDlmsData))
	data.setBytes([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	srv.setAttribute(&tDlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.err {
		t.Fatalf("%s\n", msg.err)
	}
	t.Logf("transport connected")
	dconn := msg.data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.err {
		t.Fatalf("%s\n", msg.err)
	}
	t.Logf("application connected")
	aconn := msg.data.(*AppConn)

	val := new(DlmsValueRequest)
	val.classId = 1
	val.instanceId = &tDlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.attributeId = 0x02
	vals := make([]*DlmsValueRequest, 1)
	vals[0] = val
	aconn.getRquest(ch, 10000, true, vals)
	msg = <-ch
	if nil != msg.err {
		t.Fatalf("%s\n", msg.err)
	}

	aconn.Close()
	srv.Close(t)
}

func TestX_GetRequestWithList(t *testing.T) {

	srv := startMockCosemServer(t, c_TEST_ADDR, c_TEST_PORT, c_TEST_AARE)

	data := (new(tDlmsData))
	data.setBytes([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	srv.setAttribute(&tDlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	data = (new(tDlmsData))
	data.setBytes([]byte{0x06, 0x07, 0x08, 0x08, 0x0A})
	srv.setAttribute(&tDlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.err {
		t.Fatalf("%s\n", msg.err)
	}
	t.Logf("transport connected")
	dconn := msg.data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.err {
		t.Fatalf("%s\n", msg.err)
	}
	t.Logf("application connected")
	aconn := msg.data.(*AppConn)

	vals := make([]*DlmsValueRequest, 2)

	val := new(DlmsValueRequest)
	val.classId = 1
	val.instanceId = &tDlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.attributeId = 0x02
	vals[0] = val

	val = new(DlmsValueRequest)
	val.classId = 1
	val.instanceId = &tDlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.attributeId = 0x02
	vals[1] = val

	aconn.getRquest(ch, 10000, true, vals)
	msg = <-ch
	if nil != msg.err {
		t.Fatalf("%s\n", msg.err)
	}

	aconn.Close()
	srv.Close(t)
}
