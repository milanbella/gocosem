package gocosem

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

var hdlcMeterIpA = "172.16.123.187"

func TestMeterAHdlcHdlc_TcpConnect(t *testing.T) {
	dconn, err := TcpConnect(hdlcMeterIpA, 4059)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()
}

func TestMeterAHdlcHdlc_HdlcConnect(t *testing.T) {
	dconn, err := HdlcConnect(hdlcMeterIpA, 4059, 1, 1, time.Duration(1)*time.Hour)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()
}

func TestMeterAHdlcHdlc_AppConnect(t *testing.T) {
	b := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}

	docnn, err := HdlcConnect(hdlcMeterIpA, 4059, 1, 1, time.Duration(1)*time.Hour)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aarq := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}
	ch = dconn.AppConnectRaw(01, 01, aarq)
	msg = <-ch
	if nil != msg.Err {
		t.Fatal(msg.Err)
	}
	defer aconn.Close()
	t.Logf("aare: %02X", aare)

	if !bytes.Equal(b, aare) {
		t.Fatalf("failed")
	}

}

func TestMeterAHdlcHdlc_GetTime(t *testing.T) {
	b := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}

	dconn, err := HdlcConnect(hdlcMeterIpA, 4059, 1, 1, time.Duration(1)*time.Hour)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aarq := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}
	aconn, err := dconn.AppConnectRaw(01, 01, aarq)
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()
	t.Logf("aare: %02X", aare)

	if !bytes.Equal(b, aare) {
		t.Fatalf("failed")
	}

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

func TestMeterAHdlcHdlc_SetTime(t *testing.T) {
	b := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}

	dconn, err := HdlcConnect(hdlcMeterIpA, 4059, 1, 1, time.Duration(1)*time.Hour)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aarq := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}
	aconn, err = dconn.AppConnectRaw(01, 01, aarq)
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()
	t.Logf("aare: %02X", aare)

	if !bytes.Equal(b, aare) {
		t.Fatalf("failed")
	}

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
		t.Fatalf("%s\n", msg.Err)
	}
	t.Logf("response delivered: in %v", rep.DeliveredIn())
	if 0 != rep.DataAccessResultAt(0) {
		t.Fatalf("dataAccessResult: %d\n", rep.DataAccessResultAt(0))
	}
	data = rep.DataAt(0)
	t.Logf("value read %#v", data.Val)
	t.Logf("datetime: %s", DlmsDateTimeFromBytes(data.GetOctetString()).PrintDateTime())

}

func TestMeterAHdlcHdlc_ProfileCaptureObjects(t *testing.T) {
	b := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}

	dconn, err := HdlcConnect(hdlcMeterIpA, 4059, 1, 1, time.Duration(1)*time.Hour)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aarq := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}
	aconn, err := dconn.AppConnectRaw(01, 01, aarq)
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()
	t.Logf("aare: %02X", aare)

	if !bytes.Equal(b, aare) {
		t.Fatalf("failed")
	}

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

func TestMeterAHdlcHdlc_ProfileFirstEntries(t *testing.T) {
	b := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}

	dconn, err := HdlcConnect(hdlcMeterIpA, 4059, 1, 1, time.Duration(1)*time.Hour)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aarq := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}
	aconn, err := dconn.AppConnectRaw(01, 01, aarq)
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()
	t.Logf("aare: %02X", aare)

	if !bytes.Equal(b, aare) {
		t.Fatalf("failed")
	}

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
	defer aconn.Close()
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

func TestMeterAHdlcHdlc_ProfileLastEntries(t *testing.T) {
	b := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}

	aconn, err := HdlcConnect(hdlcMeterIpA, 4059, 1, 1, time.Duration(1)*time.Hour)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aarq := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}
	aconn, err := dconn.AppConnectRaw(01, 01, aarq)
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()
	t.Logf("aare: %02X", aare)

	if !bytes.Equal(b, aare) {
		t.Fatalf("failed")
	}

	// profile entries in use

	t.Logf("read profile entries ...")
	vals := make([]*DlmsRequest, 1)
	val := new(DlmsRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 4
	vals[0] = val

	rep, err = aconn.SendRequest(vals)
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
	val.AccessParameter.Arr[0].SetDoubleLongUnsigned(entriesInUse - 10) // from_entry
	val.AccessParameter.Arr[1].SetDoubleLongUnsigned(entriesInUse)      // to_entry
	val.AccessParameter.Arr[2].SetLongUnsigned(1)                       // from_selected_value
	val.AccessParameter.Arr[3].SetLongUnsigned(0)                       // to_selected_value

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

func TestMeterAHdlcHdlc_ProfileTimeRange(t *testing.T) {
	b := []byte{0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3, 0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E, 0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D, 0x00, 0xEF, 0x00, 0x07}

	dconn, err := HdlcConnect(hdlcMeterIpA, 4059, 1, 1, time.Duration(1)*time.Hour)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aarq := []byte{0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80, 0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE, 0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00}
	aconn, err := dconn.AppConnectRaw(01, 01, aarq)
	if nil != err {
		t.Fatal(err)
	}
	defer aconn.Close()
	t.Logf("aare: %02X", aare)

	if !bytes.Equal(b, aare) {
		t.Fatalf("failed")
	}

	// profile entries in use

	t.Logf("read profile entries ...")
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
	val.AccessParameter.Arr[0].SetDoubleLongUnsigned(entriesInUse - 10) // from_entry
	val.AccessParameter.Arr[1].SetDoubleLongUnsigned(entriesInUse)      // to_entry
	val.AccessParameter.Arr[2].SetLongUnsigned(1)                       // from_selected_value
	val.AccessParameter.Arr[3].SetLongUnsigned(0)                       // to_selected_value

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

	rep, err := aconn.SendRequest(vals)
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
