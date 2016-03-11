package gocosem

import (
	"bytes"
	"testing"
)

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
	mockCosemServer.blockLength = 3
	mockCosemServer.blockDelayMsec = 300

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
	aconn.SendRequest(ch, 10000, 100, true, vals)
	msg = <-ch
	if ErrorBlockTimeout != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}

	aconn.Close()

	mockCosemServer.Close()
}

func TestX_GetRequestNormal_blockTransfer_blockTimeout(t *testing.T) {
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
	mockCosemServer.blockLength = 5

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
	mockCosemServer.blockLength = 5
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
	mockCosemServer.blockLength = 5
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

//TODO: Test is failing due to concurrency bugs on client side  (concurrent map access and writing on close channel).
// We need to change the way the invokeId is parallely handled. Perhaps we should have one go routine per invokeId one main routine receiving packets and disttributing requests accoriding invokeId.
func noTestX_1000parallelRequests(t *testing.T) {
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

func TestX_SetRequestNormal(t *testing.T) {
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

	// read value

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

	// set value

	data = (new(DlmsData))
	data.SetOctetString([]byte{0x06, 0x07, 0x08, 0x09, 0x0A})

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data // new value to set
	vals = make([]*DlmsRequest, 1)
	vals[0] = val
	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep = msg.Data.(DlmsResultResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}

	// verify if value was really set

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals = make([]*DlmsRequest, 1)
	vals[0] = val
	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep = msg.Data.(DlmsResultResponse)
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

func TestX_SetRequestNormal_blockTransfer(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	mockCosemServer.blockLength = 3

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07})
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

	// read value

	vals := make([]*DlmsRequest, 1)

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
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

	// set value using the block transfer

	data = (new(DlmsData))
	data.SetOctetString([]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17})

	vals = make([]*DlmsRequest, 1)

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data
	vals[0] = val

	vals[0].BlockSize = 5 // setting BlockSize at vals[0] will force the block transfer

	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep = msg.Data.(DlmsResultResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}

	// read value again to verify that it was set correctly

	vals = make([]*DlmsRequest, 1)

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val
	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep = msg.Data.(DlmsResultResponse)
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

func TestX_SetRequestWithList(t *testing.T) {
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

	// read values

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

	// set values

	data1 = (new(DlmsData))
	data1.SetOctetString([]byte{0x11, 0x12, 0x13, 0x14, 0x15})

	data2 = (new(DlmsData))
	data2.SetOctetString([]byte{0x16, 0x17, 0x18, 0x18, 0x1A})

	vals = make([]*DlmsRequest, 2)

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data1
	vals[0] = val

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data2
	vals[1] = val

	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep = msg.Data.(DlmsResultResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}

	// verify if values are set correctly

	vals = make([]*DlmsRequest, 2)

	val = new(DlmsRequest)
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
	rep = msg.Data.(DlmsResultResponse)
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

func TestX_SetRequestWithList_blockTransfer(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	mockCosemServer.blockLength = 5

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

	// read values

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

	// set values (using block transfer)

	data1 = (new(DlmsData))
	data1.SetOctetString([]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18})

	data2 = (new(DlmsData))
	data2.SetOctetString([]byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28})

	vals = make([]*DlmsRequest, 2)

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data1
	vals[0] = val

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data2
	vals[1] = val

	vals[0].BlockSize = 5 // setting BlockSize at vals[0] will force the block transfer

	aconn.SendRequest(ch, 10000, 1000, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	rep = msg.Data.(DlmsResultResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}

	// read same values again to verify that values were set correctly

	vals = make([]*DlmsRequest, 2)

	val = new(DlmsRequest)
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
	rep = msg.Data.(DlmsResultResponse)
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	if !bytes.Equal(data1.GetOctetString(), rep.DataAt(0).GetOctetString()) {
		t.Logf("%X", data1.GetOctetString())
		t.Logf("%X", rep.DataAt(0).GetOctetString())
		t.Fatalf("value differs: %X", rep.DataAt(0).GetOctetString())
	}
	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}
	if !bytes.Equal(data2.GetOctetString(), rep.DataAt(1).GetOctetString()) {
		t.Logf("%X", data2.GetOctetString())
		t.Logf("%X", rep.DataAt(1).GetOctetString())
		t.Fatalf("value differs: %X", rep.DataAt(1).GetOctetString())
	}

	aconn.Close()

	mockCosemServer.Close()
}

func TestX_SetRequestWithList_blockTransfer_timeout(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	mockCosemServer.blockLength = 4
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

	// try to set values
	data1 = (new(DlmsData))
	data1.SetOctetString([]byte{0x11, 0x12, 0x13, 0x14, 0x15})

	data2 = (new(DlmsData))
	data2.SetOctetString([]byte{0x16, 0x17, 0x18, 0x18, 0x1A})

	vals := make([]*DlmsRequest, 2)

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data1
	vals[0] = val

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data2
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
	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}

	aconn.Close()

	mockCosemServer.Close()
}

func TestX_SetRequestWithList_blockTransfer_blockTimeout(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
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

	// try to set values

	data1 = (new(DlmsData))
	data1.SetOctetString([]byte{0x11, 0x12, 0x13, 0x14, 0x15})

	data2 = (new(DlmsData))
	data2.SetOctetString([]byte{0x16, 0x17, 0x18, 0x18, 0x1A})

	vals := make([]*DlmsRequest, 2)

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data1
	vals[0] = val

	val = new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data2
	vals[1] = val

	vals[0].BlockSize = 5 // setting BlockSize at vals[0] will force the block transfer

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
	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}

	aconn.Close()

	mockCosemServer.Close()
}
