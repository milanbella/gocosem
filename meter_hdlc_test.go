package gocosem

import (
	"fmt"
	"testing"
	"time"
)

func init_TestMeterHdlc() {
	testMeterIp = "172.16.123.182"
	testHdlcResponseTimeout = time.Duration(1) * time.Hour
	testHdlcCosemWaitTime = time.Duration(5000) * time.Millisecond
	testHdlcSnrmTimeout = time.Duration(45) * time.Second
	testHdlcDiscTimeout = time.Duration(45) * time.Second
}

func TestMeterHdlc_TcpConnect(t *testing.T) {
	init_TestMeterHdlc()
	dconn, err := TcpConnect(testMeterIp, 4059)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()
}

func TestMeterHdlc_HdlcConnect(t *testing.T) {
	init_TestMeterHdlc()
	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()
}

func TestMeterHdlc_AppConnect(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
	defer aconn.Close()
}

func TestMeterHdlc_GetTime(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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

func TestMeterHdlc_SetTime(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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

func TestMeterHdlc_ProfileCaptureObjects(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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
		t.Logf("\tlogical name: % 02X", st.Arr[1].GetOctetString())
		t.Logf("\tattribute index: %d", st.Arr[2].GetInteger())
		t.Logf("\tdata index: % 02X", st.Arr[3].GetLongUnsigned())
	}
}

//@@@@@@@@@@@@@@@@@@
func TestMeterHdlc_ProfileEntriesInUse(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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

func TestMeterHdlc_ProfileEntries(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
	defer aconn.Close()

	// profile entries

	t.Logf("read profile entries in use...")
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

func TestMeterHdlc_ProfileSortMethod(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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

func TestMeterHdlc_ProfileSortObject(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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
	t.Logf("\tlogical name: % 02X", data.Arr[1].GetOctetString())
	t.Logf("\tattribute index: %d", data.Arr[2].GetInteger())
	t.Logf("\tdata index: % 02X", data.Arr[3].GetLongUnsigned())
}

func TestMeterHdlc_ProfileCapturePeriod(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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

//@@@@@@@@@@@@@@@@@@

func TestMeterHdlc_ProfileFirstEntries(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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
	if DATA_TYPE_ARRAY != data.GetType() {
		t.Fatalf("wrong data type")
	}
	t.Logf("profile entries read:\n")
	for i := 0; i < len(data.Arr); i++ {
		d := data.Arr[i]
		d0 := d.Arr[0]
		t.Logf("\t%d: %s %s: ", i, DlmsDateTimeFromBytes(d0.GetOctetString()).PrintDateTime(), d.Print())
	}
}

func TestMeterHdlc_ProfileLastEntries(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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
	if DATA_TYPE_ARRAY != data.GetType() {
		t.Fatalf("wrong data type")
	}
	t.Logf("profile entries read:\n")
	for i := 0; i < len(data.Arr); i++ {
		d := data.Arr[i]
		d0 := d.Arr[0]
		t.Logf("\t%d: %s %s: ", i, DlmsDateTimeFromBytes(d0.GetOctetString()).PrintDateTime(), d.Print())
	}
}

func TestMeterHdlc_ProfileTimeRange(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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
	if DATA_TYPE_ARRAY != data.GetType() {
		t.Fatalf("wrong data type")
	}
	t.Logf("profile entries read:\n")
	for i := 0; i < len(data.Arr); i++ {
		d := data.Arr[i]
		d0 := d.Arr[0]
		t.Logf("\t%d: %s %s: ", i, DlmsDateTimeFromBytes(d0.GetOctetString()).PrintDateTime(), d.Print())
	}

	d1 := data.Arr[0]
	if nil != d1.Err {
		t.Fatalf("data error: %v", d1.Err)
	}
	d2 := data.Arr[len(data.Arr)-1]
	if nil != d2.Err {
		t.Fatalf("data error: %v", d2.Err)
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

	tim := DlmsDateTimeFromBytes(d1.Arr[0].GetOctetString())
	/*
		tim := new(DlmsDateTime)

		tim.Year = 2016
		tim.Month = 2
		tim.DayOfMonth = 22
		tim.DayOfWeek = 1
		tim.Hour = 4
		tim.Minute = 16
		tim.Second = 39
		tim.Hundredths = 0
		tim.Deviation = 0
		tim.ClockStatus = 0
	*/

	t.Logf("time from: %s", tim.PrintDateTime())

	// for some reason deviation and status must be zeroed or else thi meter reports error
	tim.Deviation = 0
	tim.ClockStatus = 0

	fromValue := new(DlmsData)
	fromValue.SetOctetString(tim.ToBytes())

	tim = DlmsDateTimeFromBytes(d2.Arr[0].GetOctetString())
	/*
		tim = new(DlmsDateTime)

		tim.Year = 2016
		tim.Month = 2
		tim.DayOfMonth = 22
		tim.DayOfWeek = 1
		tim.Hour = 5
		tim.Minute = 16
		tim.Second = 39
		tim.Hundredths = 0
		tim.Deviation = 0
		tim.ClockStatus = 0
	*/

	// for some reason deviation and status must be zeroed or else thi meter reports error
	tim.Deviation = 0
	tim.ClockStatus = 0

	t.Logf("time to: %s", tim.PrintDateTime())

	toValue := new(DlmsData)
	toValue.SetOctetString(tim.ToBytes())

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
	if DATA_TYPE_ARRAY != data.GetType() {
		t.Fatalf("wrong data type")
	}
	t.Logf("profile entries read:\n")
	for i := 0; i < len(data.Arr); i++ {
		d := data.Arr[i]
		d0 := d.Arr[0]
		t.Logf("\t%d: %s %s: ", i, DlmsDateTimeFromBytes(d0.GetOctetString()).PrintDateTime(), d.Print())
	}
}

func TestMeterHdlc_ActCalendarDaytable(t *testing.T) {
	init_TestMeterHdlc()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithPassword(01, 01, 4, "12345678")
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
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
