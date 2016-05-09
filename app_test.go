package gocosem

import (
	"bytes"
	"sync"
	"testing"
)

func TestApp_TcpConnect(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	defer mockCosemServer.Close()

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

}

func TestApp_AppConnect(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	defer mockCosemServer.Close()

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

}

func TestApp_app_GetRequestNormal(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	defer mockCosemServer.Close()

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	ch = aconn.SendRequest(vals)
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
}

func TestApp_GetRequestNormal_blockTransfer(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	defer mockCosemServer.Close()
	mockCosemServer.blockLength = 3

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	ch = aconn.SendRequest(vals)
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

}

func TestApp_GetRequestWithList(t *testing.T) {
	ensureMockCosemServer(t)
	defer mockCosemServer.Close()
	mockCosemServer.Init()

	data1 := (new(DlmsData))
	data1.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data1)

	data2 := (new(DlmsData))
	data2.SetOctetString([]byte{0x06, 0x07, 0x08, 0x08, 0x0A})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}, 1, 0x02, data2)

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

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

	ch = aconn.SendRequest(vals)
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
}

func TestApp_GetRequestWithList_blockTransfer(t *testing.T) {
	ensureMockCosemServer(t)
	defer mockCosemServer.Close()
	mockCosemServer.Init()
	mockCosemServer.blockLength = 5

	data1 := (new(DlmsData))
	data1.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data1)

	data2 := (new(DlmsData))
	data2.SetOctetString([]byte{0x06, 0x07, 0x08, 0x08, 0x0A})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}, 1, 0x02, data2)

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

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

	ch = aconn.SendRequest(vals)
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
}

func TestApp_SetRequestNormal(t *testing.T) {
	ensureMockCosemServer(t)
	defer mockCosemServer.Close()
	mockCosemServer.Init()

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

	// read value

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	ch = aconn.SendRequest(vals)
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
	ch = aconn.SendRequest(vals)
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
	ch = aconn.SendRequest(vals)
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
}

func TestApp_SetRequestNormal_blockTransfer(t *testing.T) {
	ensureMockCosemServer(t)
	defer mockCosemServer.Close()
	mockCosemServer.Init()
	mockCosemServer.blockLength = 3

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

	// read value

	vals := make([]*DlmsRequest, 1)

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val
	ch = aconn.SendRequest(vals)
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

	ch = aconn.SendRequest(vals)
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
	ch = aconn.SendRequest(vals)
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
}

func TestApp_SetRequestWithList(t *testing.T) {
	ensureMockCosemServer(t)
	defer mockCosemServer.Close()
	mockCosemServer.Init()

	data1 := (new(DlmsData))
	data1.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data1)

	data2 := (new(DlmsData))
	data2.SetOctetString([]byte{0x06, 0x07, 0x08, 0x08, 0x0A})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}, 1, 0x02, data2)

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

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

	ch = aconn.SendRequest(vals)
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

	ch = aconn.SendRequest(vals)
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

	ch = aconn.SendRequest(vals)
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
}

func TestApp_SetRequestWithList_blockTransfer(t *testing.T) {
	ensureMockCosemServer(t)
	defer mockCosemServer.Close()
	mockCosemServer.Init()
	mockCosemServer.blockLength = 5

	data1 := (new(DlmsData))
	data1.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data1)

	data2 := (new(DlmsData))
	data2.SetOctetString([]byte{0x06, 0x07, 0x08, 0x08, 0x0A})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2B, 0x00, 0x00, 0xFF}, 1, 0x02, data2)

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

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

	ch = aconn.SendRequest(vals)
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

	ch = aconn.SendRequest(vals)
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

	ch = aconn.SendRequest(vals)
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
}

func TestApp_1000parallelRequests(t *testing.T) {
	ensureMockCosemServer(t)
	defer mockCosemServer.Close()
	mockCosemServer.Init()

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	ch := TcpConnect("localhost", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	defer dconn.Close()

	ch = dconn.AppConnectWithPassword(01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	defer aconn.Close()

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val

	sink := make(chan *DlmsMessage)
	count := int(1000)

	n1 := count
	countSent := 0
	countSentMtx := new(sync.Mutex)
	for i := 0; i < count; i += 1 {
		go func() {
			ch := aconn.SendRequest(vals)
			msg := <-ch
			sink <- msg
			countSentMtx.Lock()
			countSent++
			if countSent == n1 {
				close(sink)
			}
			countSentMtx.Unlock()
		}()
	}

	n2 := 0
	for msg := range sink {
		n2++
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
	}
	if n2 != count {
		t.Fatalf("wrong reamining count")
	}

}
