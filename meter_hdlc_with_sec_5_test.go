package gocosem

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func printProfile(t *testing.T, data *DlmsData) {
	if DATA_TYPE_ARRAY != data.GetType() {
		t.Fatalf("wrong data type")
	}
	t.Logf("profile entries read:\n")
	for i := 0; i < len(data.Arr); i++ {
		d := data.Arr[i]
		if d.Typ != DATA_TYPE_STRUCTURE {
			t.Logf("\t%d: error: wrong line format: %d, expecting DATA_TYPE_STRUCTURE", i, data.Typ)
			continue
		}
		if len(d.Arr) < 1 {
			t.Logf("\t%d: error: wrong line format: missing first item", i)
			continue
		}
		d0 := d.Arr[0]
		if d0.Typ == DATA_TYPE_NULL {
			t.Logf("\t%d: nil", i)
			continue
		}
		if d0.Typ != DATA_TYPE_OCTET_STRING {
			t.Logf("\t%d: error: wrong first item format: %d", i, d0.Typ)
			continue
		}
		t.Logf("\t%d: %s %s: ", i, DlmsDateTimeFromBytes(d0.GetOctetString()).PrintDateTime(), d.Print())
	}
}

func init_TestMeterHdlc_elektrotikaSecurity5() {
	testMeterIp = "127.0.0.1"
	testHdlcResponseTimeout = time.Duration(1) * time.Hour
	testHdlcCosemWaitTime = time.Duration(5000) * time.Millisecond
	testHdlcSnrmTimeout = time.Duration(45) * time.Second
	testHdlcDiscTimeout = time.Duration(45) * time.Second
}

func TestMeterHdlc_elektrotikaSecurity5_TcpConnect(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	dconn, err := TcpConnect(testMeterIp, 4059)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()
}

func TestMeterHdlc_elektrotikaSecurity5_HdlcConnect(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	physicalDeviceId := uint16(37)
	serverAddressLength := int(4)
	dconn, err := HdlcConnect(testMeterIp, 4059, 3, 1, &physicalDeviceId, &serverAddressLength, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()
}

func TestMeterHdlc_elektrotikaSecurity5_AppConnect_no_security(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()

	applicationClient := uint16(33)
	logicalDevice := uint16(1)
	physicalDeviceId := uint16(37)
	serverAddressLength := int(4)

	dconn, err := HdlcConnect(testMeterIp, 4059, applicationClient, logicalDevice, &physicalDeviceId, &serverAddressLength, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	var buf *bytes.Buffer

	var ireq DlmsInitiateRequest
	ireq.dedicatedKey = nil
	ireq.responseAllowed = true
	ireq.proposedQualityOfService = nil
	ireq.proposedDlmsVersionNumber = 6
	ireq.proposedConformance.bitsUnused = 0
	ireq.proposedConformance.buf = []byte{0xFF, 0xFF, 0xFF}
	ireq.clientMaxReceivePduSize = 0xFFFF

	buf = new(bytes.Buffer)
	err = ireq.encode(buf)
	if nil != err {
		t.Fatalf("DlmsInitiateRequest.encode() failed: %s", err)
	}
	userInformation := buf.Bytes()

	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 1})
	aarq.userInformation = (*tAsn1OctetString)(&userInformation)

	_, _, err = dconn.AppConnect(applicationClient, logicalDevice, 0x0C, &aarq)
	if nil != err {
		t.Fatalf("dconn.AppConnect() failed: %s", err)
	}
}

func TestMeterHdlc_elektrotikaSecurity5_readFrameCounter(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()

	// read frame counter

	applicationClient := uint16(33)
	logicalDevice := uint16(1)
	physicalDeviceId := uint16(37)
	serverAddressLength := int(4)

	dconn, err := HdlcConnect(testMeterIp, 4059, applicationClient, logicalDevice, &physicalDeviceId, &serverAddressLength, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	var buf *bytes.Buffer

	var initiateRequest DlmsInitiateRequest
	initiateRequest.dedicatedKey = nil
	initiateRequest.responseAllowed = true
	initiateRequest.proposedQualityOfService = nil
	initiateRequest.proposedDlmsVersionNumber = 6
	initiateRequest.proposedConformance.bitsUnused = 0
	initiateRequest.proposedConformance.buf = []byte{0xFF, 0xFF, 0xFF}
	initiateRequest.clientMaxReceivePduSize = 0xFFFF

	buf = new(bytes.Buffer)
	err = initiateRequest.encode(buf)
	if nil != err {
		t.Fatalf("DlmsInitiateRequest.encode() failed: %s", err)
	}
	userInformation := buf.Bytes()

	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 1})
	aarq.userInformation = (*tAsn1OctetString)(&userInformation)

	aconn, _, err := dconn.AppConnect(applicationClient, logicalDevice, 0x0C, &aarq)
	if nil != err {
		t.Fatalf("dconn.AppConnect() failed: %s", err)
	}

	val1 := new(DlmsRequest)
	val1.ClassId = 1
	val1.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x01, 0x00, 0xFF}
	val1.AttributeId = 0x02

	val2 := new(DlmsRequest)
	val2.ClassId = 1
	val2.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x01, 0x01, 0xFF}
	val2.AttributeId = 0x02

	vals := make([]*DlmsRequest, 2)
	vals[0] = val1
	vals[1] = val2
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("response delivered: in %v", rep.DeliveredIn())

	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data1 := rep.DataAt(0)
	t.Logf("value read %#v", data1.Val)

	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}
	data2 := rep.DataAt(1)
	t.Logf("value read %#v", data2.Val)

	if data1.Typ != DATA_TYPE_DOUBLE_LONG_UNSIGNED {
		t.Fatalf("wrong data type received")
	}
	frameCounter := data1.Val.(uint32)
	t.Logf("frame counter value: %d", frameCounter)
}

func TestMeterHdlc_elektrotikaSecurity5_AppConnect(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()

	// read frame counter

	applicationClient := uint16(33)
	logicalDevice := uint16(1)
	physicalDeviceId := uint16(37)
	serverAddressLength := int(4)

	dconn, err := HdlcConnect(testMeterIp, 4059, applicationClient, logicalDevice, &physicalDeviceId, &serverAddressLength, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	var buf *bytes.Buffer

	var initiateRequest DlmsInitiateRequest
	initiateRequest.dedicatedKey = nil
	initiateRequest.responseAllowed = true
	initiateRequest.proposedQualityOfService = nil
	initiateRequest.proposedDlmsVersionNumber = 6
	initiateRequest.proposedConformance.bitsUnused = 0
	initiateRequest.proposedConformance.buf = []byte{0xFF, 0xFF, 0xFF}
	initiateRequest.clientMaxReceivePduSize = 0xFFFF

	buf = new(bytes.Buffer)
	err = initiateRequest.encode(buf)
	if nil != err {
		t.Fatalf("DlmsInitiateRequest.encode() failed: %s", err)
	}
	userInformation := buf.Bytes()

	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 1})
	aarq.userInformation = (*tAsn1OctetString)(&userInformation)

	aconn, _, err := dconn.AppConnect(applicationClient, logicalDevice, 0x0C, &aarq)
	if nil != err {
		t.Fatalf("dconn.AppConnect() failed: %s", err)
	}

	val1 := new(DlmsRequest)
	val1.ClassId = 1
	val1.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x01, 0x00, 0xFF}
	val1.AttributeId = 0x02

	val2 := new(DlmsRequest)
	val2.ClassId = 1
	val2.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x01, 0x01, 0xFF}
	val2.AttributeId = 0x02

	vals := make([]*DlmsRequest, 2)
	vals[0] = val1
	vals[1] = val2
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("response delivered: in %v", rep.DeliveredIn())

	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data1 := rep.DataAt(0)
	t.Logf("value read %#v", data1.Val)

	if 0 != rep.DataAccessResultAt(1) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(1))
	}
	data2 := rep.DataAt(1)
	t.Logf("value read %#v", data2.Val)

	if data1.Typ != DATA_TYPE_DOUBLE_LONG_UNSIGNED {
		t.Fatalf("wrong data type received")
	}
	frameCounter := data1.Val.(uint32)
	t.Logf("frame counter value: %d", frameCounter)

	dconn.Close()

	// connect application again with high security

	applicationClient = uint16(3)
	logicalDevice = uint16(1)
	physicalDeviceId = uint16(37)
	serverAddressLength = int(4)

	dconn, err = HdlcConnect(testMeterIp, 4059, applicationClient, logicalDevice, &physicalDeviceId, &serverAddressLength, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, _, err = dconn.AppConnectWithSecurity5(applicationClient, logicalDevice, 0x0C,
		[]byte{0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF},
		[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
		[]uint32{2, 16, 756, 5, 8, 1, 3}, []byte{0x4D, 0x45, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x00}, "ZDXO2;66", &initiateRequest, frameCounter)
	if nil != err {
		t.Fatal(err)
	}

}

// Connect with highest level security 5
func Mikroelectronica_AppConnectWithSec5() (aconn *AppConn, err error) {

	// read frame counter
	debugLog("reading meter's receive frame counter")

	applicationClient := uint16(33)
	logicalDevice := uint16(1)
	physicalDeviceId := uint16(37)
	serverAddressLength := int(4)

	dconn, err := HdlcConnect(testMeterIp, 4059, applicationClient, logicalDevice, &physicalDeviceId, &serverAddressLength, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		errorLog("Microelectronica_AppConnectWithSec5(): HdlcConnect() failed")
		return nil, err
	}
	debugLog("transport connected")

	var buf *bytes.Buffer

	var initiateRequest DlmsInitiateRequest
	initiateRequest.dedicatedKey = nil
	initiateRequest.responseAllowed = true
	initiateRequest.proposedQualityOfService = nil
	initiateRequest.proposedDlmsVersionNumber = 6
	initiateRequest.proposedConformance.bitsUnused = 0
	initiateRequest.proposedConformance.buf = []byte{0xFF, 0xFF, 0xFF}
	initiateRequest.clientMaxReceivePduSize = 0xFFFF

	buf = new(bytes.Buffer)
	err = initiateRequest.encode(buf)
	if nil != err {
		return nil, err
	}
	userInformation := buf.Bytes()

	var aarq AARQapdu

	aarq.applicationContextName = tAsn1ObjectIdentifier([]uint32{2, 16, 756, 5, 8, 1, 1})
	aarq.userInformation = (*tAsn1OctetString)(&userInformation)

	aconn, _, err = dconn.AppConnect(applicationClient, logicalDevice, 0x0C, &aarq)
	if nil != err {
		errorLog("Microelectronica_AppConnectWithSec5(): AppConnect() failed")
		return nil, err
	}

	val1 := new(DlmsRequest)
	val1.ClassId = 1
	val1.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x01, 0x00, 0xFF}
	val1.AttributeId = 0x02

	val2 := new(DlmsRequest)
	val2.ClassId = 1
	val2.InstanceId = &DlmsOid{0x00, 0x00, 0x2B, 0x01, 0x01, 0xFF}
	val2.AttributeId = 0x02

	vals := make([]*DlmsRequest, 2)
	vals[0] = val1
	vals[1] = val2
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		errorLog("Microelectronica_AppConnectWithSec5(): aconn.SendRequest() failed")
		return nil, err
	}

	if 0 != rep.DataAccessResultAt(0) {
		err = fmt.Errorf("Microelectronica_AppConnectWithSec5(): dataAccessResult: %d\n", rep.DataAccessResultAt(0))
		errorLog("%s", err)
		return nil, err
	}
	data1 := rep.DataAt(0)
	debugLog("value read %#v", data1.Val)

	if 0 != rep.DataAccessResultAt(1) {
		err = fmt.Errorf("Microelectronica_AppConnectWithSec5(): dataAccessResult: %d\n", rep.DataAccessResultAt(1))
		errorLog("%s", err)
		return nil, err
	}
	data2 := rep.DataAt(1)
	debugLog("value read %#v", data2.Val)

	if data1.Typ != DATA_TYPE_DOUBLE_LONG_UNSIGNED {
		err = fmt.Errorf("Microelectronica_AppConnectWithSec5(): wrong data type received")
		errorLog("%s", err)
		return nil, err
	}
	frameCounter := data1.Val.(uint32)
	debugLog("meter's receive frame counter value: %d", frameCounter)

	dconn.Close()

	// connect application again with high security
	debugLog("opening application connection with highest level security ....")

	applicationClient = uint16(3)
	logicalDevice = uint16(1)
	physicalDeviceId = uint16(37)
	serverAddressLength = int(4)

	dconn, err = HdlcConnect(testMeterIp, 4059, applicationClient, logicalDevice, &physicalDeviceId, &serverAddressLength, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		errorLog("Microelectronica_AppConnectWithSec5(): HdlcConnect() failed")
		return nil, err
	}
	debugLog("transport connected")

	aconn, _, err = dconn.AppConnectWithSecurity5(applicationClient, logicalDevice, 0x0C,
		[]byte{0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF},
		[]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
		[]uint32{2, 16, 756, 5, 8, 1, 3}, []byte{0x4D, 0x45, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x00}, "ZDXO2;66", &initiateRequest, frameCounter)
	if nil != err {
		errorLog("Microelectronica_AppConnectWithSec5(): failed to open application connection with highest level security")
		return nil, err
	}
	debugLog("application connection with highest level security opened")
	return aconn, nil
}

func TestMeterHdlc_elektrotikaSecurity5_GetTime(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	val := new(DlmsRequest)
	val.ClassId = 8
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x01, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data := rep.DataAt(0)
	t.Logf("value read %#v", data.Val)
	t.Logf("datetime: %s", DlmsDateTimeFromBytes(data.GetOctetString()).PrintDateTime())
}

func TestMeterHdlc_elektrotikaSecurity5_SetTime(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// read time

	val := new(DlmsRequest)
	val.ClassId = 8
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x01, 0x00, 0x00, 0xFF}
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
	data := rep.DataAt(0)
	t.Logf("value read %#v", data.Val)
	t.Logf("datetime: %s", DlmsDateTimeFromBytes(data.GetOctetString()).PrintDateTime())

	// set time

	val = new(DlmsRequest)
	val.ClassId = 8
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x01, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	val.Data = data
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
	t.Logf("time set successfully")

	// read time again

	val = new(DlmsRequest)
	val.ClassId = 8
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x01, 0x00, 0x00, 0xFF}
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
	data = rep.DataAt(0)
	t.Logf("value read %#v", data.Val)
	t.Logf("datetime: %s", DlmsDateTimeFromBytes(data.GetOctetString()).PrintDateTime())

}

func TestMeterHdlc_elektrotikaSecurity5_ProfileCaptureObjects(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// capture objects definitions

	t.Logf("read objects captured by profile...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 3
	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	if DATA_TYPE_ARRAY != data.GetType() {
		t.Fatalf("wrong data type")
	}
	t.Logf("profile captures follwing objects:")
	for i, st := range data.Arr {
		if DATA_TYPE_STRUCTURE != st.GetType() {
			t.Fatalf("wrong data type")
		}
		t.Logf("capture object [%d]:", i)
		t.Logf("\tclass id: %d", st.Arr[0].GetLongUnsigned())
		t.Logf("\tlogical name: %02X", st.Arr[1].GetOctetString())
		t.Logf("\tattribute index: %d", st.Arr[2].GetInteger())
		t.Logf("\tdata index: %02X", st.Arr[3].GetLongUnsigned())
	}
}

func TestMeterHdlc_elektrotikaSecurity5_ProfileEntriesInUse(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// profile entries in use

	t.Logf("read profile entries in use...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 7
	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	t.Logf("profile entries in use: %d", data.GetDoubleLongUnsigned())
}

func TestMeterHdlc_elektrotikaSecurity5_ProfileEntries(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// profile entries

	t.Logf("read maximum profile entries...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 8
	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	t.Logf("maximum profile entries: %d", data.GetDoubleLongUnsigned())
}

func TestMeterHdlc_elektrotikaSecurity5_ProfileSortMethod(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// sort method

	t.Logf("read sort method ...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 5
	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	t.Logf("sort method: %d", data.GetEnum())
}

func TestMeterHdlc_elektrotikaSecurity5_ProfileSortObject(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// sort object

	t.Logf("read sort object ...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 6
	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)

	if DATA_TYPE_STRUCTURE != data.GetType() {
		t.Fatalf("wrong data type")
	}
	t.Logf("sort object:")
	t.Logf("\tclass id: %d", data.Arr[0].GetLongUnsigned())
	t.Logf("\tlogical name: %02X", data.Arr[1].GetOctetString())
	t.Logf("\tattribute index: %d", data.Arr[2].GetInteger())
	t.Logf("\tdata index: %02X", data.Arr[3].GetLongUnsigned())
}

func TestMeterHdlc_elektrotikaSecurity5_ProfileCapturePeriod(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// capture period

	t.Logf("read capture period ...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 4
	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	t.Logf("capture period: %d seconds", data.GetDoubleLongUnsigned())
}

func TestMeterHdlc_elektrotikaSecurity5_ProfileFirstEntries(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// request first 10 entries

	vals := make([]*DlmsRequest, 1)

	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 2
	val.AccessSelector = 2
	val.AccessParameter = new(DlmsData)
	val.AccessParameter.SetStructure(4)
	val.AccessParameter.Arr[0].SetDoubleLongUnsigned(1)  // from_entry
	val.AccessParameter.Arr[1].SetDoubleLongUnsigned(10) // to_entry
	val.AccessParameter.Arr[2].SetLongUnsigned(1)        // from_selected_value
	val.AccessParameter.Arr[3].SetLongUnsigned(0)        // to_selected_value

	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}

	data := rep.DataAt(0) // first request
	printProfile(t, data)
}

func TestMeterHdlc_elektrotikaSecurity5_ProfileLastEntries(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// profile entries in use

	t.Logf("read profile entries ...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 7
	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	entriesInUse := data.GetDoubleLongUnsigned()
	t.Logf("profile entries in use: %d", entriesInUse)

	vals = make([]*DlmsRequest, 1)

	// read last 10 entries

	entriesCnt := uint32(10)

	val = new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 2
	val.AccessSelector = 2
	val.AccessParameter = new(DlmsData)
	val.AccessParameter.SetStructure(4)
	if entriesInUse > 10 {
		val.AccessParameter.Arr[0].SetDoubleLongUnsigned(entriesInUse - entriesCnt + 1) // from_entry
	} else {
		val.AccessParameter.Arr[0].SetDoubleLongUnsigned(1) // from_entry
	}
	val.AccessParameter.Arr[1].SetDoubleLongUnsigned(entriesInUse) // to_entry
	val.AccessParameter.Arr[2].SetLongUnsigned(1)                  // from_selected_value
	val.AccessParameter.Arr[3].SetLongUnsigned(0)                  // to_selected_value

	vals[0] = val

	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult = rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}

	data = rep.DataAt(0) // first request
	printProfile(t, data)
}

func TestMeterHdlc_elektrotikaSecurity5_ProfileTimeRange(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	// read profile entries in use

	t.Logf("read profile entries ...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 7
	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	entriesInUse := data.GetDoubleLongUnsigned()
	t.Logf("profile entries in use: %d", entriesInUse)

	vals = make([]*DlmsRequest, 1)

	// read last 10 entries

	t.Logf("reading last profile entries using the time selector")

	val = new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 2
	val.AccessSelector = 2
	val.AccessParameter = new(DlmsData)
	val.AccessParameter.SetStructure(4)
	if entriesInUse > 10 {
		val.AccessParameter.Arr[0].SetDoubleLongUnsigned(entriesInUse - 10 + 1) // from_entry
	} else {
		val.AccessParameter.Arr[0].SetDoubleLongUnsigned(1) // from_entry
	}
	val.AccessParameter.Arr[1].SetDoubleLongUnsigned(entriesInUse) // to_entry
	val.AccessParameter.Arr[2].SetLongUnsigned(1)                  // from_selected_value
	val.AccessParameter.Arr[3].SetLongUnsigned(0)                  // to_selected_value

	vals[0] = val

	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult = rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}

	data = rep.DataAt(0) // first request
	printProfile(t, data)

	d1 := data.Arr[0]
	if nil != d1.Err {
		t.Fatalf("data error: %v", d1.Err)
	}
	if d1.Typ != DATA_TYPE_STRUCTURE {
		t.Fatalf("data error: unexpected format")
	}
	d2 := data.Arr[len(data.Arr)-1]
	if nil != d2.Err {
		t.Fatalf("data error: %v", d2.Err)
	}
	if d2.Typ != DATA_TYPE_STRUCTURE {
		t.Fatalf("data error: unexpected format")
	}

	// read last 10 entries using time interval selection

	vals = make([]*DlmsRequest, 1)

	val = new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 2
	val.AccessSelector = 1
	val.AccessParameter = new(DlmsData)
	val.AccessParameter.SetStructure(4)

	// selecting according first column which is the time

	restrictingObject := new(DlmsData)
	restrictingObject.SetStructure(4)
	restrictingObject.Arr[0].SetLongUnsigned(8)                                         // class_id
	restrictingObject.Arr[1].SetOctetString([]byte{0x00, 0x00, 0x01, 0x00, 0x00, 0xFF}) // logical_name
	restrictingObject.Arr[2].SetInteger(2)                                              // attribute_index
	restrictingObject.Arr[3].SetLongUnsigned(0)                                         // data_index

	var fromValue, toValue *DlmsData

	if (len(d1.Arr) < 0 || d1.Arr[0].Typ != DATA_TYPE_TIME) || len(d2.Arr) < 0 || d2.Arr[0].Typ != DATA_TYPE_TIME {
		// first column does not contain time, perhaps time is nil
		t.Logf("data error: unexpected format of time column")
		t.Logf("            therefore reading profile entries from yesterday")

		gtim := time.Now()
		gtim = gtim.AddDate(0, 0, -1)
		year, month, day := gtim.Date()
		hour, min, sec := gtim.Clock()

		tim := new(DlmsDateTime)
		tim.Year = uint16(year)
		tim.Month = uint8(month)
		tim.DayOfMonth = uint8(day)
		tim.DayOfWeek = 0xFF
		tim.Hour = uint8(hour)
		tim.Minute = uint8(min)
		tim.Second = uint8(sec)
		tim.Hundredths = 0
		tim.Deviation = 0
		tim.ClockStatus = 0

		t.Logf("time from: %s", tim.PrintDateTime())

		fromValue = new(DlmsData)
		fromValue.SetOctetString(tim.ToBytes())

		gtim = time.Now()
		year, month, day = gtim.Date()
		hour, min, sec = gtim.Clock()

		tim = new(DlmsDateTime)
		tim.Year = uint16(year)
		tim.Month = uint8(month)
		tim.DayOfMonth = uint8(day)
		tim.DayOfWeek = 0xFF
		tim.Hour = uint8(hour)
		tim.Minute = uint8(min)
		tim.Second = uint8(sec)
		tim.Hundredths = 0
		tim.Deviation = 0
		tim.ClockStatus = 0

		t.Logf("time to: %s", tim.PrintDateTime())

		toValue = new(DlmsData)
		toValue.SetOctetString(tim.ToBytes())
	} else {
		// read last profile entries again using time selector

		// read the start time from fisrt profile entry

		tim := DlmsDateTimeFromBytes(d1.Arr[0].GetOctetString())
		t.Logf("time from: %s", tim.PrintDateTime())
		// for some reason deviation and status must be zeroed or else this meter reports error
		tim.Deviation = 0
		tim.ClockStatus = 0
		fromValue = new(DlmsData)
		fromValue.SetOctetString(tim.ToBytes())

		// read the end time from last profile entry

		tim = DlmsDateTimeFromBytes(d2.Arr[0].GetOctetString())
		// for some reason deviation and status must be zeroed or else this meter reports error
		tim.Deviation = 0
		tim.ClockStatus = 0
		t.Logf("time to: %s", tim.PrintDateTime())
		toValue = new(DlmsData)
		toValue.SetOctetString(tim.ToBytes())
	}

	selectedValues := new(DlmsData)
	selectedValues.SetArray(0)

	val.AccessParameter.Arr[0] = restrictingObject
	val.AccessParameter.Arr[1] = fromValue
	val.AccessParameter.Arr[2] = toValue
	val.AccessParameter.Arr[3] = selectedValues

	vals[0] = val

	rep, err = aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult = rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}

	data = rep.DataAt(0) // first request
	printProfile(t, data)
}

func TestMeterHdlc_elektrotikaSecurity5_ActCalendarDaytable(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	t.Logf("read day table...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 20
	val.InstanceId = &DlmsOid{0, 0, 13, 0, 0, 255}
	val.AttributeId = 5
	vals[0] = val

	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("read failed: %s", err)
		return
	}
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	if DATA_TYPE_ARRAY != data.GetType() {
		t.Fatalf("wrong data type")
	}
	t.Logf("daytable:")
	for i, st := range data.Arr {
		if DATA_TYPE_STRUCTURE != st.GetType() {
			t.Fatalf("wrong data type")
		}
		t.Logf("daytable [%d]: ", i)
		t.Logf("\tid: %d", st.Arr[0].GetUnsigned())
		for a, da := range st.Arr[1].Arr {
			if DATA_TYPE_STRUCTURE != da.GetType() {
				t.Fatalf("wrong data type: %v", da.GetType())
			}
			t.Logf("\taction [%d]:", a)
			t.Logf("\t  start: %v", da.Arr[0].GetTime())
			t.Logf("\t  script obis: %v", da.Arr[1].GetOctetString())
			t.Logf("\t  script selector: %d", da.Arr[2].GetLongUnsigned())
		}
	}
}

func TestMeterHdlc_elektrotikaSecurity5_StateOfDisconnector(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	instanceId := &DlmsOid{0x00, 0x00, 0x60, 0x03, 0x0A, 0xFF}
	classId := DlmsClassId(70)
	attributeIdControlState := DlmsAttributeId(3)
	attributeIdControlMode := DlmsAttributeId(4)

	// Read control mode

	val := new(DlmsRequest)
	val.ClassId = classId
	val.InstanceId = instanceId
	val.AttributeId = attributeIdControlMode
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	t.Logf("response took: %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data := rep.DataAt(0)
	if DATA_TYPE_ENUM != data.GetType() {
		t.Fatalf("not integer")
	}
	controlMode := data.GetEnum()
	t.Logf("control mode: %d", controlMode)

	// Check connected state.

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
	t.Logf("response took: %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data = rep.DataAt(0)
	if DATA_TYPE_ENUM != data.GetType() {
		t.Fatalf("not integer")
	}

	controlState := data.GetEnum()
	t.Logf("control state: %d", controlState)

}

// TODO: This test is failing because disconnector control mode is not 2.
func TestMeterHdlc_elektrotikaSecurity5_Disconnector(t *testing.T) {
	init_TestMeterHdlc_elektrotikaSecurity5()
	aconn, err := Mikroelectronica_AppConnectWithSec5()
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()

	instanceId := &DlmsOid{0x00, 0x00, 0x60, 0x03, 0x0A, 0xFF}
	classId := DlmsClassId(70)
	attributeIdControlState := DlmsAttributeId(3)
	attributeIdControlMode := DlmsAttributeId(4)
	methodIdRemoteDisconnect := DlmsMethodId(1)
	methodIdRemoteConnect := DlmsMethodId(2)
	//stateDisconnected := uint8(0)
	//stateConnected := uint8(1)

	// Read control mode

	val := new(DlmsRequest)
	val.ClassId = classId
	val.InstanceId = instanceId
	val.AttributeId = attributeIdControlMode
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	rep, err := aconn.SendRequest(vals)
	if nil != err {
		t.Fatalf("%s\n", err)
	}
	//t.Logf("response took: %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data := rep.DataAt(0)
	if DATA_TYPE_ENUM != data.GetType() {
		t.Fatalf("not integer")
	}
	controlMode := data.GetEnum()
	t.Logf("control mode: %d", data.GetEnum())

	// Check if control mode is acceptable.
	if controlMode != 2 {
		t.Fatalf("unsupported control mode: %v", controlMode)
	}

	// Check connected state.

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
	//t.Logf("response took: %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data = rep.DataAt(0)
	if DATA_TYPE_ENUM != data.GetType() {
		t.Fatalf("not integer")
	}

	controlState := data.GetEnum()
	t.Logf("control state: %d", controlState)

	// Based on current control state try to disconnect or connect.
	// At the end of test always return meter to connected state.

	switch controlState {

	case 0: // disconnected

		// Call remote_connect method.

		method := new(DlmsRequest)
		method.ClassId = classId
		method.InstanceId = instanceId
		method.MethodId = methodIdRemoteConnect
		methodParameters := new(DlmsData)
		methodParameters.SetInteger(1)
		method.MethodParameters = methodParameters
		methods := make([]*DlmsRequest, 1)
		methods[0] = method
		rep, err = aconn.SendRequest(methods)
		if nil != err {
			t.Fatalf("%s\n", err)
		}
		//t.Logf("response took: %v", rep.DeliveredIn())
		if 0 != rep.ActionResultAt(0) {
			t.Fatalf("actionResult: %d\n", rep.ActionResultAt(0))
		}

		// Check connected state.

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
		//t.Logf("response took: %v", rep.DeliveredIn())
		if 0 != rep.DataAccessResultAt(0) {
			t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
		}
		data = rep.DataAt(0)
		if DATA_TYPE_ENUM != data.GetType() {
			t.Fatalf("not integer")
		}

		controlState := data.GetEnum()
		t.Logf("control state: %d", controlState)
		if 1 != controlState {
			t.Fatalf("meter did not connect, control state: %d", controlState)
		}

	case 1: // connected

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
		//t.Logf("response took: %v", rep.DeliveredIn())
		if 0 != rep.ActionResultAt(0) {
			t.Fatalf("actionResult: %d\n", rep.ActionResultAt(0))
		}

		// Check connected state.

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
		//t.Logf("response took: %v", rep.DeliveredIn())
		if 0 != rep.DataAccessResultAt(0) {
			t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
		}
		data = rep.DataAt(0)
		if DATA_TYPE_ENUM != data.GetType() {
			t.Fatalf("not integer")
		}

		controlState := data.GetEnum()
		t.Logf("control state: %d", controlState)
		if 0 != controlState {
			t.Fatalf("meter did not disconnect, control state: %d", controlState)
		}

		// Pause before connect
		time.Sleep(time.Second * 1)

		// Call remote_connect method.

		method = new(DlmsRequest)
		method.ClassId = classId
		method.InstanceId = instanceId
		method.MethodId = methodIdRemoteConnect
		methodParameters = new(DlmsData)
		methodParameters.SetInteger(1)
		method.MethodParameters = methodParameters
		methods = make([]*DlmsRequest, 1)
		methods[0] = method
		rep, err = aconn.SendRequest(methods)
		if nil != err {
			t.Fatalf("%s\n", err)
		}
		//t.Logf("response took: %v", rep.DeliveredIn())
		if 0 != rep.ActionResultAt(0) {
			t.Fatalf("actionResult: %d\n", rep.ActionResultAt(0))
		}

		// Check final state.

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
		//t.Logf("response took: %v", rep.DeliveredIn())
		if 0 != rep.DataAccessResultAt(0) {
			t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
		}
		data = rep.DataAt(0)
		if DATA_TYPE_ENUM != data.GetType() {
			t.Fatalf("not integer")
		}

		controlState = data.GetEnum()
		t.Logf("control state: %d", controlState)
		if 1 != controlState {
			t.Fatalf("meter did not connect, control state: %d", controlState)
		}

	case 3: // ready for connection

		// Call remote_connect method.

		method := new(DlmsRequest)
		method.ClassId = classId
		method.InstanceId = instanceId
		method.MethodId = methodIdRemoteConnect
		methodParameters := new(DlmsData)
		methodParameters.SetInteger(1)
		method.MethodParameters = methodParameters
		methods := make([]*DlmsRequest, 1)
		methods[0] = method
		rep, err = aconn.SendRequest(methods)
		if nil != err {
			t.Fatalf("%s\n", err)
		}
		//t.Logf("response took: %v", rep.DeliveredIn())
		if 0 != rep.ActionResultAt(0) {
			t.Fatalf("actionResult: %d\n", rep.ActionResultAt(0))
		}

		// Check connected state.

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
		//t.Logf("response took: %v", rep.DeliveredIn())
		if 0 != rep.DataAccessResultAt(0) {
			t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
		}
		data = rep.DataAt(0)
		if DATA_TYPE_ENUM != data.GetType() {
			t.Fatalf("not integer")
		}

		controlState := data.GetEnum()
		t.Logf("control state: %d", controlState)
		if 1 != controlState {
			t.Fatalf("meter did not connect, control state: %d", controlState)
		}

	default:
		t.Fatalf("unknown controlState: %d", controlState)
	}
}
