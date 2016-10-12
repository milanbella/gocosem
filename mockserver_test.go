package gocosem

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

type tMockCosemObjectMethod func(*DlmsData) (DlmsActionResult, *DlmsDataAccessResult, *DlmsData)

type tMockCosemObject struct {
	classId    DlmsClassId
	attributes map[DlmsAttributeId]*DlmsData           // all attributes with their values
	methods    map[DlmsMethodId]tMockCosemObjectMethod // all methods
}

type tMockCosemServer struct {
	//closed              bool
	ln                  net.Listener
	connections         *list.List  // list of *tMockCosemServerConnection
	connections_mtx     *sync.Mutex // TODO: avoid mutex
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
	//TODO: refactor, create request structure instead indexing each item by invokeId
	blocks           map[uint8][][]byte             // blocks to be sent in case of outbound block transfer (key is invokeId)
	rawData          map[uint8]*bytes.Buffer        // raw data received so far in case of inbound block transfer (key is invokeId)
	classIds         map[uint8][]DlmsClassId        // key is invokeId
	instanceIds      map[uint8][]*DlmsOid           // key is invokeId
	attributeIds     map[uint8][]DlmsAttributeId    // key is invokeId
	methodIds        map[uint8][]DlmsAttributeId    // key is invokeId
	accessSelectors  map[uint8][]DlmsAccessSelector // key is invokeId
	accessParameters map[uint8][]*DlmsData          // key is invokeId
}

//TODO: refactor
func (conn *tMockCosemServerConnection) sendEncodedReply(t *testing.T, b0 byte, b1 byte, invokeIdAndPriority tDlmsInvokeIdAndPriority, dataAccessResult DlmsDataAccessResult, reply []byte) (err error) {
	var buf bytes.Buffer

	invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)
	l := conn.srv.blockLength // block length
	//if len(reply) > l {
	if (0xC4 == b0) && (0x02 == b1) {
		// use block transfer
		t.Logf("outbound block transfer")

		blocks := make([][]byte, len(reply)/l+1)
		b := reply[0:]
		var i int
		for i = 0; len(b) > l; i += 1 {
			blocks[i] = b[0:l]
			b = b[l:]
		}
		blocks[i] = b
		blocks = blocks[0 : i+1] // truncate since we may have allocated more
		conn.blocks[invokeId] = blocks

		t.Logf("blocks count: %d", len(blocks))
		/*
			for i = 0; i < len(blocks); i += 1 {
				t.Logf("block[%d]: %02X", i, blocks[i])
			}
		*/

		_, err := buf.Write([]byte{b0, b1, byte(invokeIdAndPriority)})
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		err = encode_GetResponsewithDataBlock(&buf, false, 1, dataAccessResult, blocks[0])
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		err = ipTransportSend(conn.rwc, conn.logicalDevice, conn.applicationClient, buf.Bytes())
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

	} else {
		t.Logf("outbound normal transfer")
		_, err := buf.Write([]byte{b0, b1, byte(invokeIdAndPriority)})
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		if (0xC4 == b0) && (0x03 != b1) { // only  Get responses except to Get response with list
			_, err = buf.Write([]byte{byte(dataAccessResult)})
			if nil != err {
				t.Errorf("%v\n", err)
				return err
			}
		}
		_, err = buf.Write(reply)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		err = ipTransportSend(conn.rwc, conn.logicalDevice, conn.applicationClient, buf.Bytes())
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
	}
	return nil
}

func (conn *tMockCosemServerConnection) setBlockReply(t *testing.T, invokeIdAndPriority tDlmsInvokeIdAndPriority, lastBlock bool, blockNumber uint32) (err error) {
	invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)

	if lastBlock {
		buf := conn.rawData[invokeId]

		data := new(DlmsData)
		err = data.Decode(buf)
		if nil != err {
			return err
		}

		dataAccessResult := conn.srv.setData(t, conn.classIds[invokeId][0], conn.instanceIds[invokeId][0], conn.attributeIds[invokeId][0], conn.accessSelectors[invokeId][0], conn.accessParameters[invokeId][0], data)
		t.Logf("dataAccessResult: %d", dataAccessResult)

		delete(conn.rawData, invokeId)
		delete(conn.classIds, invokeId)
		delete(conn.instanceIds, invokeId)
		delete(conn.attributeIds, invokeId)
		delete(conn.accessSelectors, invokeId)
		delete(conn.accessParameters, invokeId)

		buf = new(bytes.Buffer)
		err = encode_SetResponseForLastDataBlock(buf, dataAccessResult, blockNumber)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		err = conn.sendEncodedReply(t, 0xC5, 0x03, invokeIdAndPriority, 0, buf.Bytes())
		if nil != err {
			return err
		}
	} else {
		var buf bytes.Buffer

		err = encode_SetResponseForDataBlock(&buf, blockNumber)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		err = conn.sendEncodedReply(t, 0xC5, 0x02, invokeIdAndPriority, 0, buf.Bytes())
		if nil != err {
			return err
		}
	}

	return nil

}

func (conn *tMockCosemServerConnection) setBlockListReply(t *testing.T, invokeIdAndPriority tDlmsInvokeIdAndPriority, lastBlock bool, blockNumber uint32) (err error) {
	invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)

	if lastBlock {
		buf := conn.rawData[invokeId]

		var count uint8
		err = binary.Read(buf, binary.BigEndian, &count)
		if nil != err {
			t.Errorf("binary.Read() failed, err: %v", err)
			return err
		}

		dataAccessResults := make([]DlmsDataAccessResult, count)

		for i := 0; i < int(count); i++ {
			data := new(DlmsData)
			err = data.Decode(buf)
			if nil != err {
				return err
			}

			dataAccessResults[i] = conn.srv.setData(t, conn.classIds[invokeId][i], conn.instanceIds[invokeId][i], conn.attributeIds[invokeId][i], conn.accessSelectors[invokeId][i], conn.accessParameters[invokeId][i], data)
			t.Logf("dataAccessResults[i]: %d", dataAccessResults[i])
		}

		delete(conn.rawData, invokeId)
		delete(conn.classIds, invokeId)
		delete(conn.instanceIds, invokeId)
		delete(conn.attributeIds, invokeId)
		delete(conn.accessSelectors, invokeId)
		delete(conn.accessParameters, invokeId)

		buf = new(bytes.Buffer)
		err = encode_SetResponseForLastDataBlockWithList(buf, dataAccessResults, blockNumber)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		err = conn.sendEncodedReply(t, 0xC5, 0x04, invokeIdAndPriority, 0, buf.Bytes())
		if nil != err {
			return err
		}
	} else {
		var buf bytes.Buffer

		err = encode_SetResponseForDataBlock(&buf, blockNumber)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		err = conn.sendEncodedReply(t, 0xC5, 0x02, invokeIdAndPriority, 0, buf.Bytes())
		if nil != err {
			return err
		}
	}

	return nil
}

func (conn *tMockCosemServerConnection) replyToRequest(t *testing.T, r io.Reader) (err error) {
	p := make([]byte, 3)
	err = binary.Read(r, binary.BigEndian, p)
	if nil != err {
		t.Errorf("%v\n", err)
		return err
	}

	invokeIdAndPriority := tDlmsInvokeIdAndPriority(p[2])
	invokeId := uint8((invokeIdAndPriority & 0xF0) >> 4)

	if bytes.Equal(p[0:2], []byte{0xC0, 0x01}) {
		t.Logf("GetRequestNormal")

		err, classId, instanceId, attributeId, accessSelector, accessParameters := decode_GetRequestNormal(r)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		dataAccessResult, data := conn.srv.getData(t, classId, instanceId, attributeId, accessSelector, accessParameters)
		t.Logf("dataAccessResult: %d", dataAccessResult)

		var buf bytes.Buffer

		if conn.srv.blockLength <= 0 {
			err = encode_GetResponseNormalBlock(&buf, data)
			if nil != err {
				t.Errorf("%v\n", err)
				return err
			}
			err = conn.sendEncodedReply(t, 0xC4, 0x01, invokeIdAndPriority, dataAccessResult, buf.Bytes())
			if nil != err {
				return err
			}
		} else {
			err = encode_GetResponseNormalBlock(&buf, data)
			if nil != err {
				t.Errorf("%v\n", err)
				return err
			}
			err = conn.sendEncodedReply(t, 0xC4, 0x02, invokeIdAndPriority, dataAccessResult, buf.Bytes())
			if nil != err {
				return err
			}
		}

	} else if bytes.Equal(p[0:2], []byte{0xC0, 0x03}) {
		t.Logf("GetRequestWithList")

		err, classIds, instanceIds, attributeIds, accessSelectors, accessParameters := decode_GetRequestWithList(r)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		count := len(classIds)
		datas := make([]*DlmsData, count)
		dataAccessResults := make([]DlmsDataAccessResult, count)

		for i := 0; i < count; i += 1 {
			dataAccessResult, data := conn.srv.getData(t, classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i])
			t.Logf("dataAccessResult[%d]: %d", i, dataAccessResult)
			dataAccessResults[i] = dataAccessResult
			datas[i] = data
		}

		var buf bytes.Buffer

		if conn.srv.blockLength <= 0 {
			err = encode_GetResponseWithList(&buf, dataAccessResults, datas)
			if nil != err {
				t.Errorf("%v\n", err)
				return err
			}
			conn.sendEncodedReply(t, 0xC4, 0x03, invokeIdAndPriority, 0, buf.Bytes())
		} else {
			err = encode_GetResponseWithList(&buf, dataAccessResults, datas)
			if nil != err {
				t.Errorf("%v\n", err)
				return err
			}
			conn.sendEncodedReply(t, 0xC4, 0x02, invokeIdAndPriority, 0, buf.Bytes())
		}

	} else if bytes.Equal(p[0:2], []byte{0xC0, 0x02}) {
		t.Logf("GetRequestForNextDataBlock")

		err, blockNumber := decode_GetRequestForNextDataBlock(r)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

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
		t.Logf("dataAccessResult: %d", dataAccessResult)

		if (len(conn.blocks[invokeId]) - 1) == int(blockNumber) {
			lastBlock = true
		} else {
			lastBlock = false
		}

		if lastBlock {
			delete(conn.blocks, invokeId)
		}

		if !lastBlock {
			blockNumber += 1
		}
		err = encode_GetResponsewithDataBlock(&buf, lastBlock, blockNumber, dataAccessResult, rawData)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		if conn.srv.blockDelayMsec > 0 {
			if !conn.srv.blockDelayLastBlock || (conn.srv.blockDelayLastBlock && lastBlock) {
				<-time.After(time.Millisecond * time.Duration(conn.srv.blockDelayMsec))
			}
		}
		err = ipTransportSend(conn.rwc, conn.logicalDevice, conn.applicationClient, buf.Bytes())
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		return nil
	} else if bytes.Equal(p[0:2], []byte{0xC1, 0x01}) {
		t.Logf("SetRequestNormal")

		err, classId, instanceId, attributeId, accessSelector, accessParameters, data := decode_SetRequestNormal(r)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		dataAccessResult := conn.srv.setData(t, classId, instanceId, attributeId, accessSelector, accessParameters, data)
		t.Logf("dataAccessResult: %d", dataAccessResult)

		var buf bytes.Buffer

		err = encode_SetResponseNormal(&buf, dataAccessResult)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		err = conn.sendEncodedReply(t, 0xC5, 0x01, invokeIdAndPriority, 0, buf.Bytes())
		if nil != err {
			return err
		}
	} else if bytes.Equal(p[0:2], []byte{0xC1, 0x04}) {
		t.Logf("SetRequestNormalWithList")

		err, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, datas := decode_SetRequestWithList(r)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		count := len(classIds)
		dataAccessResults := make([]DlmsDataAccessResult, count)

		for i := 0; i < count; i++ {
			dataAccessResults[i] = conn.srv.setData(t, classIds[i], instanceIds[i], attributeIds[i], accessSelectors[i], accessParameters[i], datas[i])
			t.Logf("dataAccessResult[%d]: %d", i, dataAccessResults[i])
		}

		var buf bytes.Buffer

		err = encode_SetResponseWithList(&buf, dataAccessResults)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}
		err = conn.sendEncodedReply(t, 0xC5, 0x05, invokeIdAndPriority, 0, buf.Bytes())
		if nil != err {
			return err
		}

	} else if bytes.Equal(p[0:2], []byte{0xC1, 0x02}) {
		t.Logf("SetRequestNormalBlock")

		err, classId, instanceId, attributeId, accessSelector, accessParameters, lastBlock, blockNumber, rawData := decode_SetRequestNormalBlock(r)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		conn.rawData[invokeId] = new(bytes.Buffer)

		conn.classIds[invokeId] = make([]DlmsClassId, 1)
		conn.instanceIds[invokeId] = make([]*DlmsOid, 1)
		conn.attributeIds[invokeId] = make([]DlmsAttributeId, 1)
		conn.accessSelectors[invokeId] = make([]DlmsAccessSelector, 1)
		conn.accessParameters[invokeId] = make([]*DlmsData, 1)

		_, err = conn.rawData[invokeId].Write(rawData)
		if nil != err {
			return err
		}
		conn.classIds[invokeId][0] = classId
		conn.instanceIds[invokeId][0] = instanceId
		conn.attributeIds[invokeId][0] = attributeId
		conn.accessSelectors[invokeId][0] = accessSelector
		conn.accessParameters[invokeId][0] = accessParameters

		err = conn.setBlockReply(t, invokeIdAndPriority, lastBlock, blockNumber)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

	} else if bytes.Equal(p[0:2], []byte{0xC1, 0x05}) {
		t.Logf("SetRequestWithListBlock")

		err, classIds, instanceIds, attributeIds, accessSelectors, accessParameters, lastBlock, blockNumber, rawData := decode_SetRequestWithListBlock(r)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		conn.rawData[invokeId] = new(bytes.Buffer)

		conn.classIds[invokeId] = classIds
		conn.instanceIds[invokeId] = instanceIds
		conn.attributeIds[invokeId] = attributeIds
		conn.accessSelectors[invokeId] = accessSelectors
		conn.accessParameters[invokeId] = accessParameters

		_, err = conn.rawData[invokeId].Write(rawData)
		if nil != err {
			return err
		}

		err = conn.setBlockListReply(t, invokeIdAndPriority, lastBlock, blockNumber)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

	} else if bytes.Equal(p[0:2], []byte{0xC1, 0x03}) {
		t.Logf("SetRequestWithDataBlock")

		err, lastBlock, blockNumber, rawData := decode_SetRequestWithDataBlock(r)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		_, err = conn.rawData[invokeId].Write(rawData)
		if nil != err {
			return err
		}

		if conn.srv.blockDelayMsec > 0 {
			if !conn.srv.blockDelayLastBlock || (conn.srv.blockDelayLastBlock && lastBlock) {
				<-time.After(time.Millisecond * time.Duration(conn.srv.blockDelayMsec))
			}
		}

		isList := len(conn.classIds[invokeId]) > 1

		if isList {
			err = conn.setBlockListReply(t, invokeIdAndPriority, lastBlock, blockNumber)
			if nil != err {
				t.Errorf("%v\n", err)
				return err
			}
		} else {
			err = conn.setBlockReply(t, invokeIdAndPriority, lastBlock, blockNumber)
			if nil != err {
				t.Errorf("%v\n", err)
				return err
			}
		}
	} else if bytes.Equal(p[0:2], []byte{0xC3, 0x01}) {
		t.Logf("received ActionRequestNormal")

		err, classId, instanceId, methodId, methodParameters := decode_ActionRequestNormal(r)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		actionResult, dataAccessResult, data := conn.srv.callMethod(t, classId, instanceId, methodId, methodParameters)
		t.Logf("actionResult: %d", actionResult)

		t.Logf("sending ActionResponseNormal")

		var buf bytes.Buffer

		_, err = buf.Write([]byte{0xC7, 0x01, byte(invokeIdAndPriority)})
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		err = encode_ActionResponseNormal(&buf, actionResult, dataAccessResult, data)
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

		err = ipTransportSend(conn.rwc, conn.logicalDevice, conn.applicationClient, buf.Bytes())
		if nil != err {
			t.Errorf("%v\n", err)
			return err
		}

	} else {
		panic("assertion failed")
	}
	return nil
}

func (conn *tMockCosemServerConnection) receiveAndReply(t *testing.T) {
	for {

		pdu, _, _, err := ipTransportReceive(conn.rwc, &conn.applicationClient, &conn.logicalDevice)
		if nil != err {
			if io.EOF != err {
				t.Errorf("%v\n", err)
			}
			conn.rwc.Close()
			break
		}

		if conn.srv.replyDelayMsec <= 0 {
			err := conn.replyToRequest(t, bytes.NewBuffer(pdu))
			if nil != err {
				t.Errorf("%v\n", err)
				conn.rwc.Close()
				break
			}
		} else {
			<-time.After(time.Millisecond * time.Duration(conn.srv.replyDelayMsec))
			err := conn.replyToRequest(t, bytes.NewBuffer(pdu))
			if nil != err {
				t.Errorf("%v\n", err)
				conn.rwc.Close()
				break
			}
		}
	}
}

func (srv *tMockCosemServer) objectKey(instanceId *DlmsOid) string {
	return fmt.Sprintf("%d_%d_%d_%d_%d_%d", instanceId[0], instanceId[1], instanceId[2], instanceId[3], instanceId[4], instanceId[5])
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

func (srv *tMockCosemServer) setData(t *testing.T, classId DlmsClassId, instanceId *DlmsOid, attributeId DlmsAttributeId, accessSelector DlmsAccessSelector, accessParameters *DlmsData, data *DlmsData) (dataAccessResult DlmsDataAccessResult) {
	if nil == instanceId {
		panic("assertion failed")
	}
	key := srv.objectKey(instanceId)
	obj, ok := srv.objects[key]
	if !ok {
		t.Logf("no such instance id: setting dataAccessResult to 1")
		return 1
	} else {
		if obj.classId == classId {
			_, ok = obj.attributes[attributeId]
			if !ok {
				t.Logf("no such instance attribute: setting dataAccessResult to 1")
				return 1
			}
			obj.attributes[attributeId] = data
			return 0
		} else {
			t.Logf("instance class mismatch: setting dataAccessResult to 1")
			return 1
		}
	}
}

func (srv *tMockCosemServer) callMethod(t *testing.T, classId DlmsClassId, instanceId *DlmsOid, methodId DlmsMethodId, methodParameters *DlmsData) (actionResult DlmsActionResult, dataAccessResult *DlmsDataAccessResult, data *DlmsData) {
	if nil == instanceId {
		panic("assertion failed")
	}
	key := srv.objectKey(instanceId)
	obj, ok := srv.objects[key]
	if !ok {
		t.Logf("no such instance id: setting actionResult to 1")
		return 1, nil, nil
	} else {
		if obj.classId == classId {
			method, ok := obj.methods[methodId]
			if !ok {
				t.Logf("no such instance method: setting actionResult to 1")
				return 1, nil, nil
			}
			return method(methodParameters)
		} else {
			t.Logf("instance class mismatch: setting actionResult to 1")
			return 1, nil, nil
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

func (srv *tMockCosemServer) setMethod(instanceId *DlmsOid, classId DlmsClassId, methodId DlmsMethodId, method tMockCosemObjectMethod) {

	key := srv.objectKey(instanceId)
	obj := srv.objects[key]
	if nil == obj {
		obj = new(tMockCosemObject)
		srv.objects[key] = obj
	}
	obj.classId = classId
	methods := obj.methods
	if nil == methods {
		methods = make(map[DlmsMethodId]tMockCosemObjectMethod)
		obj.methods = methods
	}
	if nil != method {
		methods[methodId] = noopMethod
	} else {
		methods[methodId] = method
	}
}

func noopMethod(methodParameters *DlmsData) (DlmsActionResult, *DlmsDataAccessResult, *DlmsData) {
	return 0, nil, nil
}

func (srv *tMockCosemServer) acceptApp(t *testing.T, rwc io.ReadWriteCloser, aare []byte) (err error) {
	t.Logf("mock server waiting for client to connect")

	// receive aarq
	_, src, dst, err := ipTransportReceive(rwc, nil, nil)
	if nil != err {
		if io.EOF != err {
			t.Errorf("%v\n", err)
		}
		return err
	}

	logicalDevice := dst
	applicationClient := src

	// reply with aare
	err = ipTransportSend(rwc, logicalDevice, applicationClient, aare)
	if nil != err {
		t.Errorf("%v\n", err)
		rwc.Close()
		return err
	}

	conn := new(tMockCosemServerConnection)
	conn.srv = srv
	conn.rwc = rwc
	conn.logicalDevice = logicalDevice
	conn.applicationClient = applicationClient

	conn.blocks = make(map[uint8][][]byte)

	conn.rawData = make(map[uint8]*bytes.Buffer)
	conn.classIds = make(map[uint8][]DlmsClassId)
	conn.instanceIds = make(map[uint8][]*DlmsOid)
	conn.attributeIds = make(map[uint8][]DlmsAttributeId)
	conn.methodIds = make(map[uint8][]DlmsAttributeId)
	conn.accessSelectors = make(map[uint8][]DlmsAccessSelector)
	conn.accessParameters = make(map[uint8][]*DlmsData)

	mockCosemServer.connections_mtx.Lock()
	srv.connections.PushBack(conn)
	mockCosemServer.connections_mtx.Unlock()

	go conn.receiveAndReply(t)
	return nil
}

func (srv *tMockCosemServer) accept(t *testing.T, tcpAddr string, aare []byte) (err error) {
	ln, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		t.Errorf("%v\n", err)
		return err
	}
	srv.ln = ln

	t.Logf("mock server bound to %s", tcpAddr)

	go func() {
		for {
			conn, err := srv.ln.Accept()
			if err != nil {
				t.Errorf("%v\n", err)
				return
			}
			go srv.acceptApp(t, conn, aare)
		}
	}()
	return nil
}

var mockCosemServer *tMockCosemServer

func startMockCosemServer(t *testing.T, addr string, port int, aare []byte) {

	tcpAddr := fmt.Sprintf("%s:%d", addr, port)

	mockCosemServer = new(tMockCosemServer)
	mockCosemServer.connections = list.New()
	mockCosemServer.connections_mtx = &sync.Mutex{}
	err := mockCosemServer.accept(t, tcpAddr, aare)
	if nil != err {
		t.Fatal(err)
	}
}

func (srv *tMockCosemServer) Close() {
	srv.connections_mtx.Lock()
	for e := srv.connections.Front(); e != nil; e = e.Next() {
		sconn := e.Value.(*tMockCosemServerConnection)
		if !sconn.closed {
			sconn.closed = true
		}
	}
	srv.connections = list.New()
	srv.connections_mtx.Unlock()
}

func (srv *tMockCosemServer) Init() {
	srv.Close()

	srv.connections_mtx.Lock()
	srv.connections = list.New()
	srv.connections_mtx.Unlock()
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
		startMockCosemServer(t, c_TEST_ADDR, c_TEST_PORT, c_TEST_AARE)
	}
}
