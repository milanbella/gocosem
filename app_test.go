package gocosem

import (
	"bytes"
	"container/list"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

type tMockCosemObject struct {
	classId    DlmsClassId
	attributes map[DlmsAttributeId]*DlmsData
}

type tMockCosemServer struct {
	closed         bool
	ln             net.Listener
	connections    *list.List // list of *tMockCosemServerConnection
	objects        map[string]*tMockCosemObject
	blockLength    int
	replyDelayMsec int
	blockDelayMsec int
}

type tMockCosemServerConnection struct {
	srv               *tMockCosemServer
	closed            bool
	rwc               io.ReadWriteCloser
	logicalDevice     uint16
	applicationClient uint16
	blocks            map[uint8][][]byte // invoke id bolocks to be sent in case of block transfer
}

func (conn *tMockCosemServerConnection) sendEncodedReply(t *testing.T, invokeIdAndPriority tDlmsInvokeIdAndPriority, reply []byte) (err error) {
	var FNAME string = "tMockCosemServerConnection.sendEncodedReply()"

	invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)
	l := conn.srv.blockLength // block length
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
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		ch := make(DlmsChannel)
		ipTransportSend(ch, conn.rwc, conn.logicalDevice, conn.applicationClient, reply)
		msg := <-ch
		if nil != msg.Err {
			errorLog.Printf("%s: %v\n", FNAME, msg.Err)
			return err
		}

	} else {
		t.Logf("%s: using normal transfer", FNAME)
		ch := make(DlmsChannel)
		ipTransportSend(ch, conn.rwc, conn.logicalDevice, conn.applicationClient, reply)
		msg := <-ch
		if nil != msg.Err {
			errorLog.Printf("%s: %v\n", FNAME, msg.Err)
			return err
		}
	}
	return nil
}

func (conn *tMockCosemServerConnection) replyToRequest(t *testing.T, pdu []byte) (err error) {
	var FNAME string = "tMockCosemServerConnection.replyToRequest()"

	if bytes.Equal(pdu[0:2], []byte{0xC0, 0x01}) {
		t.Logf("%s: GetRequestNormal", FNAME)
		err, invokeIdAndPriority, classId, instanceId, attributeId, accessSelector, accessParameters := decode_GetRequestNormal(pdu)
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		dataAccessResult, data := conn.srv.getData(t, classId, instanceId, attributeId, accessSelector, accessParameters)
		t.Logf("%s: dataAccessResult: %d", FNAME, dataAccessResult)
		err, rawData := encode_GetResponseNormal(invokeIdAndPriority, dataAccessResult, data)
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		err = conn.sendEncodedReply(t, invokeIdAndPriority, rawData)
		if nil != err {
			return err
		}
	} else if bytes.Equal(pdu[0:2], []byte{0xC0, 0x03}) {
		t.Logf("%s: GetRequestWithList", FNAME)
		err, invokeIdAndPriority, classIds, instanceIds, attributeIds, accessSelectors, accessParameters := decode_GetRequestWithList(pdu)
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		count := len(classIds)
		var rawData []byte
		datas := make([]*DlmsData, count)
		dataAccessResults := make([]DlmsDataAccessResult, count)
		for i := 0; i < count; i += 1 {
			dataAccessResult, data := conn.srv.getData(t, classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
			t.Logf("%s: dataAccessResult[%d]: %d", FNAME, i, dataAccessResult)
			dataAccessResults[i] = dataAccessResult
			datas[i] = data
		}
		err, rawData = encode_GetResponseWithList(invokeIdAndPriority, dataAccessResults, datas)
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		conn.sendEncodedReply(t, invokeIdAndPriority, rawData)
	} else if bytes.Equal(pdu[0:2], []byte{0xC0, 0x02}) {
		f := func() error {
			t.Logf("%s: GetRequestForNextDataBlock", FNAME)
			err, invokeIdAndPriority, blockNumber := decode_GetRequestForNextDataBlock(pdu)
			if nil != err {
				errorLog.Printf("%s: %v\n", FNAME, err)
				return err
			}
			invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)

			var dataAccessResult DlmsDataAccessResult
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
				errorLog.Printf("%s: %v\n", FNAME, err)
				return err
			}
			ch := make(DlmsChannel)
			ipTransportSend(ch, conn.rwc, conn.logicalDevice, conn.applicationClient, reply)
			msg := <-ch
			if nil != msg.Err {
				errorLog.Printf("%s: %v\n", FNAME, msg.Err)
				return err
			}
			return nil
		}
		if conn.srv.blockDelayMsec > 0 {
			<-time.After(time.Millisecond * time.Duration(conn.srv.blockDelayMsec))
			err = f()
			return err
		} else {
			err = f()
			return err
		}
	} else {
		panic("assertion failed")
	}
	return nil
}

func (conn *tMockCosemServerConnection) receiveAndReply(t *testing.T) (err error) {
	var (
		FNAME string = "tMockCosemServerConnection.receiveAndReply()"
	)

	for (!conn.closed) && (!conn.srv.closed) {

		ch := make(DlmsChannel)
		ipTransportReceive(ch, conn.rwc, &conn.applicationClient, &conn.logicalDevice)
		msg := <-ch
		if nil != msg.Err {
			errorLog.Printf("%s: %v\n", FNAME, msg.Err)
			conn.rwc.Close()
			break
		}
		m := msg.Data.(map[string]interface{})
		if nil == m["pdu"] {
			panic("assertion failed")
		}

		go func() {
			if conn.srv.replyDelayMsec <= 0 {
				err := conn.replyToRequest(t, m["pdu"].([]byte))
				if nil != err {
					errorLog.Printf("%s: %v\n", FNAME, err)
					conn.rwc.Close()
				}
			} else {
				<-time.After(time.Millisecond * time.Duration(conn.srv.replyDelayMsec))
				err := conn.replyToRequest(t, m["pdu"].([]byte))
				if nil != err {
					errorLog.Printf("%s: %v\n", FNAME, err)
					conn.rwc.Close()
				}
			}
		}()
	}
	t.Logf("%s: mock server: closing client connection", FNAME)
	conn.rwc.Close()
	return nil
}

func (srv *tMockCosemServer) objectKey(instanceId *DlmsOid) string {
	return fmt.Sprintf("%d_%d_%d_%d_%d_%d_%d", instanceId[0], instanceId[1], instanceId[2], instanceId[3], instanceId[4], instanceId[5])
}

func (srv *tMockCosemServer) getData(t *testing.T, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector *DlmsAccessSelector, accessParameters *DlmsData) (dataAccessResult DlmsDataAccessResult, data *DlmsData) {
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

func (srv *tMockCosemServer) setAttribute(instanceId *DlmsOid, classId DlmsClassId, attributeId DlmsAttributeId, data *DlmsData) {

	key := srv.objectKey(instanceId)
	obj := srv.objects[key]
	if nil == obj {
		obj = new(tMockCosemObject)
		srv.objects[key] = obj
	}
	obj.classId = classId
	attributes := obj.attributes
	if nil == attributes {
		attributes = make(map[DlmsAttributeId]*DlmsData)
		obj.attributes = attributes
	}
	attributes[attributeId] = data
}

func (srv *tMockCosemServer) acceptApp(t *testing.T, rwc io.ReadWriteCloser, aare []byte) (err error) {
	var (
		FNAME string = "tMockCosemServer.acceptApp()"
	)

	t.Logf("%s: mock server waiting for client to connect", FNAME)

	// receive aarq
	ch := make(DlmsChannel)
	ipTransportReceive(ch, rwc, nil, nil)
	msg := <-ch
	if nil != msg.Err {
		errorLog.Printf("%s: %v\n", FNAME, msg.Err)
		rwc.Close()
		return err
	}
	m := msg.Data.(map[string]interface{})

	logicalDevice := m["dstWport"].(uint16)
	applicationClient := m["srcWport"].(uint16)

	// reply with aare
	ipTransportSend(ch, rwc, logicalDevice, applicationClient, aare)
	msg = <-ch
	if nil != msg.Err {
		errorLog.Printf("%s: %v\n", FNAME, msg.Err)
		rwc.Close()
		return err
	}

	conn := new(tMockCosemServerConnection)
	conn.srv = srv
	conn.rwc = rwc
	conn.logicalDevice = logicalDevice
	conn.applicationClient = applicationClient
	conn.blocks = make(map[uint8][][]byte)
	srv.connections.PushBack(conn)

	go conn.receiveAndReply(t)
	return nil
}

func (srv *tMockCosemServer) accept(t *testing.T, ch DlmsChannel, tcpAddr string, aare []byte) {
	var (
		FNAME string = "tMockCosemServer.accept()"
	)

	ln, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		errorLog.Printf("%s: %v\n", FNAME, err)
		msg := new(DlmsChannelMessage)
		msg.Err = err
		ch <- msg
		return
	}
	srv.ln = ln

	t.Logf("%s: mock server bound to %s", FNAME, tcpAddr)
	msg := new(DlmsChannelMessage)
	msg.Err = nil
	ch <- msg

	for {
		conn, err := srv.ln.Accept()
		if err != nil {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return
		}
		go srv.acceptApp(t, conn, aare)
	}
}

var mockCosemServer *tMockCosemServer

func startMockCosemServer(t *testing.T, ch DlmsChannel, addr string, port int, aare []byte) {

	tcpAddr := fmt.Sprintf("%s:%d", addr, port)

	mockCosemServer = new(tMockCosemServer)
	mockCosemServer.connections = list.New()
	go mockCosemServer.accept(t, ch, tcpAddr, aare)
}

func (srv *tMockCosemServer) Close() {
	for e := srv.connections.Front(); e != nil; e = e.Next() {
		sconn := e.Value.(*tMockCosemServerConnection)
		if !sconn.closed {
			sconn.closed = true
			sconn.rwc.Close()
		}
	}
	srv.connections = list.New()
}

func (srv *tMockCosemServer) Init() {
	srv.Close()

	srv.connections = list.New()
	srv.objects = make(map[string]*tMockCosemObject)
	srv.blockLength = 1000
	srv.replyDelayMsec = 0
}

const c_TEST_ADDR = "localhost"
const c_TEST_PORT = 4059

var c_TEST_AARE = []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x18, 0x1F, 0x08, 0x00, 0x00, 0x07}

func ensureMockCosemServer(t *testing.T) {

	if nil == mockCosemServer {
		ch := make(DlmsChannel)
		startMockCosemServer(t, ch, c_TEST_ADDR, c_TEST_PORT, c_TEST_AARE)
		msg := <-ch
		if nil != msg.Err {
			t.Fatalf("%s\n", msg.Err)
			mockCosemServer = nil
		}
	}
}

func TestX__TcpConnect(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	dconn.Close()

	mockCosemServer.Close()
}

func TestX_AppConnect(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	aconn.Close()

	mockCosemServer.Close()
}

func TestX_GetRequestNormal(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)

	val := new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsValueRequest, 1)
	vals[0] = val
	aconn.getRquest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	if !bytes.Equal(data.GetOctetString(), rep.DataAt(0).GetOctetString()) {
		t.Fatalf("value differs")
	}

	aconn.Close()

	mockCosemServer.Close()
}

func TestX_GetRequestNormal_blockTransfer(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	mockCosemServer.blockLength = 10

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)

	val := new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsValueRequest, 1)
	vals[0] = val
	aconn.getRquest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	if !bytes.Equal(data.GetOctetString(), rep.DataAt(0).GetOctetString()) {
		t.Fatalf("value differs")
	}

	aconn.Close()

	mockCosemServer.Close()
}

func TestX_GetRequestWithList(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()

	data1 := (new(DlmsData))
	data1.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data1)

	data2 := (new(DlmsData))
	data2.SetOctetString([]byte{0x06, 0x07, 0x08, 0x08, 0x0A})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}, 1, 0x02, data2)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)

	vals := make([]*DlmsValueRequest, 2)

	val := new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val

	val = new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[1] = val

	aconn.getRquest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	if !bytes.Equal(data1.GetOctetString(), rep.DataAt(0).GetOctetString()) {
		t.Fatalf("value differs")
	}
	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}
	if !bytes.Equal(data2.GetOctetString(), rep.DataAt(1).GetOctetString()) {
		t.Fatalf("value differs")
	}

	aconn.Close()

	mockCosemServer.Close()
}

func TestX_GetRequestWithList_blockTransfer(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	mockCosemServer.blockLength = 10

	data1 := (new(DlmsData))
	data1.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data1)

	data2 := (new(DlmsData))
	data2.SetOctetString([]byte{0x06, 0x07, 0x08, 0x08, 0x0A})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}, 1, 0x02, data2)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)

	vals := make([]*DlmsValueRequest, 2)

	val := new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val

	val = new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[1] = val

	aconn.getRquest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	if !bytes.Equal(data1.GetOctetString(), rep.DataAt(0).GetOctetString()) {
		t.Fatalf("value differs")
	}
	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}
	if !bytes.Equal(data2.GetOctetString(), rep.DataAt(1).GetOctetString()) {
		t.Fatalf("value differs")
	}

	aconn.Close()

	mockCosemServer.Close()
}

func TestX_GetRequestWithList_blockTransfer_timeout(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	mockCosemServer.blockLength = 10
	mockCosemServer.replyDelayMsec = 1000

	data1 := (new(DlmsData))
	data1.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data1)

	data2 := (new(DlmsData))
	data2.SetOctetString([]byte{0x06, 0x07, 0x08, 0x08, 0x0A})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}, 1, 0x02, data2)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)

	vals := make([]*DlmsValueRequest, 2)

	val := new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val

	val = new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[1] = val

	// expect request timeout

	aconn.getRquest(ch, 500, 10000, true, vals)
	msg = <-ch
	if ErrorRequestTimeout != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}

	// timeouted request must not disable following requests

	mockCosemServer.replyDelayMsec = 0
	aconn.getRquest(ch, 500, 100, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	if !bytes.Equal(data1.GetOctetString(), rep.DataAt(0).GetOctetString()) {
		t.Fatalf("value differs")
	}
	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}
	if !bytes.Equal(data2.GetOctetString(), rep.DataAt(1).GetOctetString()) {
		t.Fatalf("value differs")
	}

	aconn.Close()

	mockCosemServer.Close()
}

func TestX_GetRequestWithList_blockTransfer_blockTimeout(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	mockCosemServer.blockLength = 10
	mockCosemServer.blockDelayMsec = 1000

	data1 := (new(DlmsData))
	data1.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data1)

	data2 := (new(DlmsData))
	data2.SetOctetString([]byte{0x06, 0x07, 0x08, 0x08, 0x0A})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}, 1, 0x02, data2)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)

	vals := make([]*DlmsValueRequest, 2)

	val := new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val

	val = new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[1] = val

	// expect block request timeout

	aconn.getRquest(ch, 10000, 900, true, vals)
	msg = <-ch
	if ErrorBlockTimeout != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}

	// timeouted request must not disable following requests

	mockCosemServer.replyDelayMsec = 0
	aconn.getRquest(ch, 10000, 2000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	if !bytes.Equal(data1.GetOctetString(), rep.DataAt(0).GetOctetString()) {
		t.Fatalf("value differs")
	}
	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}
	if !bytes.Equal(data2.GetOctetString(), rep.DataAt(1).GetOctetString()) {
		t.Fatalf("value differs")
	}

	aconn.Close()

	mockCosemServer.Close()
}

func TestX_1000parallelRequests(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)

	val := new(DlmsValueRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsValueRequest, 1)
	vals[0] = val

	sink := make(DlmsChannel)
	count := int(1000)

	for i := 0; i < count; i += 1 {
		go func() {
			aconn.getRquest(ch, 10000, 1000, true, vals)
			msg = <-ch
			sink <- msg
		}()
	}

sinkLoop:
	for {
		msg := <-sink
		count -= 1
		if nil != msg.Err {
			t.Fatalf("%s\n", msg.Err)
		}
		rep := msg.Data.(DlmsResponse)
		t.Logf("response delivered: in %v", rep.DeliveredIn())
		if 0 != rep.DataAccessResultAt(0) {
			t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
		}
		if !bytes.Equal(data.GetOctetString(), rep.DataAt(0).GetOctetString()) {
			t.Fatalf("value differs")
		}
		if 0 == count {
			break sinkLoop
		}
	}

	aconn.Close()
	mockCosemServer.Close()
}
