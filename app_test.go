package gocosem

import (
	"bytes"
	"container/list"
	"encoding/binary"
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
	closed              bool
	ln                  net.Listener
	connections         *list.List // list of *tMockCosemServerConnection
	objects             map[string]*tMockCosemObject
	blockLength         int
	replyDelayMsec      int
	blockDelayMsec      int
	blockDelayLastBlock bool
}

type tMockCosemServerConnection struct {
	srv               *tMockCosemServer
	closed            bool
	rwc               io.ReadWriteCloser
	logicalDevice     uint16
	applicationClient uint16
	blocks            map[uint8][][]byte // invoke id bolocks to be sent in case of block transfer
}

func (conn *tMockCosemServerConnection) sendEncodedReply(t *testing.T, b0 byte, b1 byte, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResult DlmsDataAccessResult, reply []byte) (err error) {
	var FNAME string = "tMockCosemServerConnection.sendEncodedReply()"

	var buf bytes.Buffer

	invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)
	l := conn.srv.blockLength // block length
	//if len(reply) > l {
	if 0x02 == b1 {
		// use block transfer
		t.Logf("%s: using block transfer", FNAME)

		blocks := make([][]byte, len(reply)/l+1)
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
		/*
			for i = 0; i < len(blocks); i += 1 {
				t.Logf("%s: block[%d]: %02X", FNAME, i, blocks[i])
			}
		*/

		//buf.Write([]byte{0xC4, 0x02, byte(invokeIdAndPriority)})
		_, err := buf.Write([]byte{b0, b1, byte(invokeIdAndPriority)})
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		err = encode_GetResponsewithDataBlock(&buf, false, 1, dataAccessResult, blocks[0])
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		ch := make(DlmsChannel)
		ipTransportSend(ch, conn.rwc, conn.logicalDevice, conn.applicationClient, buf.Bytes())
		msg := <-ch
		if nil != msg.Err {
			errorLog.Printf("%s: %v\n", FNAME, msg.Err)
			return err
		}

	} else {
		t.Logf("%s: using normal transfer", FNAME)
		ch := make(DlmsChannel)
		_, err := buf.Write([]byte{b0, b1, byte(invokeIdAndPriority)})
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		if 0x03 != b1 {
			_, err := buf.Write([]byte{byte(dataAccessResult)})
			if nil != err {
				errorLog.Printf("%s: %v\n", FNAME, err)
				return err
			}
		}
		_, err = buf.Write(reply)
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		ipTransportSend(ch, conn.rwc, conn.logicalDevice, conn.applicationClient, buf.Bytes())
		msg := <-ch
		if nil != msg.Err {
			errorLog.Printf("%s: %v\n", FNAME, msg.Err)
			return err
		}
	}
	return nil
}

func (conn *tMockCosemServerConnection) replyToRequest(t *testing.T, r io.Reader) (err error) {
	var FNAME string = "tMockCosemServerConnection.replyToRequest()"

	p := make([]byte, 3)
	err = binary.Read(r, binary.BigEndian, p)
	if nil != err {
		errorLog.Printf("%s: %v\n", FNAME, err)
		return err
	}

	invokeIdAndPriority := tDlmsInvokeIdAndPriority(p[2])

	if bytes.Equal(p[0:2], []byte{0xC0, 0x01}) {
		t.Logf("%s: GetRequestNormal", FNAME)

		err, classId, instanceId, attributeId, accessSelector, accessParameters := decode_GetRequestNormal(r)
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}

		dataAccessResult, data := conn.srv.getData(t, classId, instanceId, attributeId, accessSelector, accessParameters)
		t.Logf("%s: dataAccessResult: %d", FNAME, dataAccessResult)

		var buf bytes.Buffer

		if conn.srv.blockLength <= 0 {
			err = encode_GetResponseNormalBlock(&buf, data)
			if nil != err {
				errorLog.Printf("%s: %v\n", FNAME, err)
				return err
			}
			err = conn.sendEncodedReply(t, 0xC4, 0x01, invokeIdAndPriority, dataAccessResult, buf.Bytes())
			if nil != err {
				return err
			}
		} else {
			err = encode_GetResponseNormalBlock(&buf, data)
			if nil != err {
				errorLog.Printf("%s: %v\n", FNAME, err)
				return err
			}
			err = conn.sendEncodedReply(t, 0xC4, 0x02, invokeIdAndPriority, dataAccessResult, buf.Bytes())
			if nil != err {
				return err
			}
		}

	} else if bytes.Equal(p[0:2], []byte{0xC0, 0x03}) {
		t.Logf("%s: GetRequestWithList", FNAME)

		err, classIds, instanceIds, attributeIds, accessSelectors, accessParameters := decode_GetRequestWithList(r)
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}

		count := len(classIds)
		datas := make([]*DlmsData, count)
		dataAccessResults := make([]DlmsDataAccessResult, count)

		for i := 0; i < count; i += 1 {
			dataAccessResult, data := conn.srv.getData(t, classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
			t.Logf("%s: dataAccessResult[%d]: %d", FNAME, i, dataAccessResult)
			dataAccessResults[i] = dataAccessResult
			datas[i] = data
		}

		var buf bytes.Buffer

		if conn.srv.blockLength <= 0 {
			err = encode_GetResponseWithList(&buf, dataAccessResults, datas)
			if nil != err {
				errorLog.Printf("%s: %v\n", FNAME, err)
				return err
			}
			conn.sendEncodedReply(t, 0xC4, 0x03, invokeIdAndPriority, 0, buf.Bytes())
		} else {
			err = encode_GetResponseWithList(&buf, dataAccessResults, datas)
			if nil != err {
				errorLog.Printf("%s: %v\n", FNAME, err)
				return err
			}
			conn.sendEncodedReply(t, 0xC4, 0x02, invokeIdAndPriority, 0, buf.Bytes())
		}

	} else if bytes.Equal(p[0:2], []byte{0xC0, 0x02}) {
		t.Logf("%s: GetRequestForNextDataBlock", FNAME)

		err, blockNumber := decode_GetRequestForNextDataBlock(r)
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)

		var dataAccessResult DlmsDataAccessResult
		var rawData []byte
		var lastBlock bool

		var buf bytes.Buffer
		buf.Write([]byte{0xC4, 0x02, byte(invokeIdAndPriority)})

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
		err = encode_GetResponsewithDataBlock(&buf, lastBlock, blockNumber, dataAccessResult, rawData)
		if nil != err {
			errorLog.Printf("%s: %v\n", FNAME, err)
			return err
		}
		ch := make(DlmsChannel)
		if conn.srv.blockDelayMsec > 0 {
			if !conn.srv.blockDelayLastBlock || (conn.srv.blockDelayLastBlock && lastBlock) {
				<-time.After(time.Millisecond * time.Duration(conn.srv.blockDelayMsec))
			}
		}
		ipTransportSend(ch, conn.rwc, conn.logicalDevice, conn.applicationClient, buf.Bytes())
		msg := <-ch
		if nil != msg.Err {
			errorLog.Printf("%s: %v\n", FNAME, msg.Err)
			return err
		}
		return nil
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
				err := conn.replyToRequest(t, bytes.NewBuffer(m["pdu"].([]byte)))
				if nil != err {
					errorLog.Printf("%s: %v\n", FNAME, err)
					conn.rwc.Close()
				}
			} else {
				<-time.After(time.Millisecond * time.Duration(conn.srv.replyDelayMsec))
				err := conn.replyToRequest(t, bytes.NewBuffer(m["pdu"].([]byte)))
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

func (srv *tMockCosemServer) getData(t *testing.T, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData) (dataAccessResult DlmsDataAccessResult, data *DlmsData) {
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
	srv.blockLength = 0
	srv.replyDelayMsec = 0
	srv.blockDelayMsec = 0
	srv.blockDelayLastBlock = false
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

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResultResponse)
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
	mockCosemServer.blockLength = 3

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

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResultResponse)
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

func TestX_GetRequestNormal_blockTransfer_timeout(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	mockCosemServer.blockLength = 5
	mockCosemServer.blockDelayMsec = 200
	mockCosemServer.blockDelayLastBlock = true

	data := (new(DlmsData))
	data.Typ = DATA_TYPE_ARRAY
	data.Arr = make([]*DlmsData, 4)

	i := 0
	d := (new(DlmsData))
	d.SetOctetString([]byte{0x00, 0x01, 0x02, 0x03})
	data.Arr[i] = d

	i += 1
	d = (new(DlmsData))
	d.SetLong(10)
	data.Arr[i] = d

	i += 1
	d = (new(DlmsData))
	d.SetLong(20)
	data.Arr[i] = d

	i += 1
	d = (new(DlmsData))
	d.SetOctetString([]byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01})
	data.Arr[i] = d

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

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val

	aconn.SendRequest(ch, 100000, 100, true, vals)
	msg = <-ch
	if ErrorBlockTimeout != msg.Err {
		t.Fatalf("%v\n", msg.Err)
	}

	rep := msg.Data.(DlmsResultResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}

	// even if request timeouted partially received data must be decoded correctly

	rdata := rep.DataAt(0)

	if nil != rdata.Arr[0].Err {
		t.Fatalf("data not parsed")
	}
	if !bytes.Equal(data.Arr[0].GetOctetString(), rdata.Arr[0].GetOctetString()) {
		t.Fatalf("value differs")
	}

	if nil != rdata.Arr[1].Err {
		t.Fatalf("data not parsed")
	}
	if data.Arr[1].GetLong() != rdata.Arr[1].GetLong() {
		t.Fatalf("value differs")
	}

	if nil != rdata.Arr[2].Err {
		t.Fatalf("data not parsed")
	}
	if data.Arr[2].GetLong() != rdata.Arr[2].GetLong() {
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

	vals := make([]*DlmsRequest, 2)

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[1] = val

	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResultResponse)
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

	vals := make([]*DlmsRequest, 2)

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[1] = val

	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResultResponse)
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

	vals := make([]*DlmsRequest, 2)

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[1] = val

	// expect request timeout

	aconn.SendRequest(ch, 500, 10000, true, vals)
	msg = <-ch
	if ErrorRequestTimeout != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}

	// timeouted request must not disable following requests

	mockCosemServer.replyDelayMsec = 0
	aconn.SendRequest(ch, 500, 100, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResultResponse)
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

	vals := make([]*DlmsRequest, 2)

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[1] = val

	// expect block request timeout

	aconn.SendRequest(ch, 10000, 900, true, vals)
	msg = <-ch
	if ErrorBlockTimeout != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}

	// timeouted request must not disable following requests

	aconn.SendRequest(ch, 10000, 2000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep := msg.Data.(DlmsResultResponse)
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

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val

	sink := make(DlmsChannel)
	count := int(1000)

	for i := 0; i < count; i += 1 {
		go func() {
			aconn.SendRequest(ch, 10000, 1000, true, vals)
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
		rep := msg.Data.(DlmsResultResponse)
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
