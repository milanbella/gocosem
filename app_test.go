package gocosem

import (
	"bytes"
	"testing"
)

func TestApp_TcpConnect(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	defer mockCosemServer.Close()

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

}

func TestApp_AppConnect(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	defer mockCosemServer.Close()

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
	defer aconn.Close()

}

func TestApp_app_GetRequestNormal(t *testing.T) {
	ensureMockCosemServer(t)
	mockCosemServer.Init()
	defer mockCosemServer.Close()

	data := (new(DlmsData))
	data.SetOctetString([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
	mockCosemServer.setAttribute(&DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}, 1, 0x02, data)

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
	defer aconn.Close()

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
	defer aconn.Close()

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
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

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
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

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
	defer aconn.Close()

	// read value

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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
	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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
	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
	defer aconn.Close()

	// read value

	vals := make([]*DlmsRequest, 1)

	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals[0] = val
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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
	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
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

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
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

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
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

func noTestApp_ActionRequestNormal(t *testing.T) {

	ensureMockCosemServer(t)
	defer mockCosemServer.Close()
	mockCosemServer.Init()

	// Test is based on class id 70 (disconnect control) behaviour.

	instanceId := &DlmsOid{0x00, 0x00, 0x60, 0x03, 0x0A, 0xFF}
	classId := DlmsClassId(70)
	attributeIdControlState := DlmsAttributeId(3)
	methodIdRemoteDisconnect := DlmsMethodId(1)
	stateDisconnected := uint8(0)
	stateConnected := uint8(1)

	// Set initial state to connected.

	data := (new(DlmsData))
	data.SetEnum(stateConnected)
	mockCosemServer.setAttribute(instanceId, classId, attributeIdControlState, data)

	// Define remote_disconnect method.

	methodRemoteDisconnect := func(obj *tMockCosemObject, methodParameters *DlmsData) (DlmsActionResult, *DlmsDataAccessResult, *DlmsData) {

		if nil == methodParameters {
			t.Fatal("no method parameter")
		}

		if DATA_TYPE_INTEGER != methodParameters.GetType() {
			t.Fatal("parameter type is not integer")
		}

		data := (new(DlmsData))
		data.SetEnum(stateDisconnected)
		obj.attributes[attributeIdControlState] = data

		return 0, nil, nil
	}

	mockCosemServer.setMethod(instanceId, classId, methodIdRemoteDisconnect, methodRemoteDisconnect)

	// Mock disconnect control set up. Start testing.

	dconn, err := TcpConnect("localhost", 4059)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 0, "12345678")
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("application connected")
	defer aconn.Close()

	// Read value, expect connected state.

	val := new(DlmsRequest)
	val.ClassId = classId
	val.InstanceId = instanceId
	val.AttributeId = attributeIdControlState
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data = rep.DataAt(0)
	if DATA_TYPE_ENUM != data.GetType() {
		t.Fatalf("not integer")
	}
	if stateConnected != data.GetEnum() {
		t.Fatalf("not connected state")
	}

	// Call remote_disconnect method.

	method := new(DlmsRequest)
	method.ClassId = classId
	method.InstanceId = instanceId
	method.MethodId = methodIdRemoteDisconnect
	methodParameters := new(DlmsData)
	methodParameters.SetInteger(1)
	method.MethodParameters = methodParameters
	methods := make([]*DlmsRequest, 1)
	methods[0] = method
	rep, err = aconn.SendRequest(methods)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.ActionResultAt(0) {
		t.Fatalf("actionResult: %d\n", rep.ActionResultAt(0))
	}
	if nil != rep.DataAt(0) {
		t.Fatal("method returned data")
	}

	// Verify if state is disconncted now.

	val = new(DlmsRequest)
	val.ClassId = classId
	val.InstanceId = instanceId
	val.AttributeId = attributeIdControlState
	vals = make([]*DlmsRequest, 1)
	vals[0] = val
	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data = rep.DataAt(0)
	if DATA_TYPE_ENUM != data.GetType() {
		t.Fatalf("not integer")
	}
	if stateDisconnected != data.GetEnum() {
		t.Fatalf("connected state")
	}

}
