package gocosem

import (
	"testing"
)

func TestX__profileRead_captureObjects(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect tcp: %s", msg.Err)
		return
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect app: %s", msg.Err)
		return
	}
	t.Logf("application connected")
	defer dconn.Close()
	aconn := msg.Data.(*AppConn)

	// capture objects definitions

	t.Logf("read objects captured by profile...")
	vals := make([]*DlmsValueRequest, 1)
	val := new(DlmsValueRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 3
	vals[0] = val

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep := msg.Data.(DlmsResponse)
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

	aconn.Close()
}

func TestX__profileRead_profileEntriesInUse(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect tcp: %s", msg.Err)
		return
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect app: %s", msg.Err)
		return
	}
	t.Logf("application connected")
	defer dconn.Close()
	aconn := msg.Data.(*AppConn)

	// profile entries in use

	t.Logf("read profile entries in use...")
	vals := make([]*DlmsValueRequest, 1)
	val := new(DlmsValueRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 7
	vals[0] = val

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep := msg.Data.(DlmsResponse)
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	t.Logf("profile entries in use: %d", data.GetDoubleLongUnsigned())

	aconn.Close()
}

func TestX__profileRead_sortMethod(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect tcp: %s", msg.Err)
		return
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect app: %s", msg.Err)
		return
	}
	t.Logf("application connected")
	defer dconn.Close()
	aconn := msg.Data.(*AppConn)

	// sort method

	t.Logf("read sort method ...")
	vals := make([]*DlmsValueRequest, 1)
	val := new(DlmsValueRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 5
	vals[0] = val

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep := msg.Data.(DlmsResponse)
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	t.Logf("sort method: %d", data.GetEnum())

	aconn.Close()
}

func TestX__profileRead_sortObject(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect tcp: %s", msg.Err)
		return
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect app: %s", msg.Err)
		return
	}
	t.Logf("application connected")
	defer dconn.Close()
	aconn := msg.Data.(*AppConn)

	// sort object

	t.Logf("read sort object ...")
	vals := make([]*DlmsValueRequest, 1)
	val := new(DlmsValueRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 6
	vals[0] = val

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep := msg.Data.(DlmsResponse)
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

	aconn.Close()
}

func TestX__profileRead_capturePeriod(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect tcp: %s", msg.Err)
		return
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect app: %s", msg.Err)
		return
	}
	t.Logf("application connected")
	defer dconn.Close()
	aconn := msg.Data.(*AppConn)

	// capture period

	t.Logf("read capture period ...")
	vals := make([]*DlmsValueRequest, 1)
	val := new(DlmsValueRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 4
	vals[0] = val

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep := msg.Data.(DlmsResponse)
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	t.Logf("capture period: %d seconds", data.GetDoubleLongUnsigned())

	aconn.Close()
}

func TestX__profileRead_first_entries(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect tcp: %s", msg.Err)
		return
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect app: %s", msg.Err)
		return
	}
	t.Logf("application connected")
	defer dconn.Close()
	aconn := msg.Data.(*AppConn)

	// request first 10 entries

	vals := make([]*DlmsValueRequest, 1)

	val := new(DlmsValueRequest)
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

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep := msg.Data.(DlmsResponse)
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

	aconn.Close()
}

func TestX__profileRead_last_entries(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect tcp: %s", msg.Err)
		return
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect app: %s", msg.Err)
		return
	}
	t.Logf("application connected")
	defer dconn.Close()
	aconn := msg.Data.(*AppConn)

	// profile entries in use

	t.Logf("read profile entries ...")
	vals := make([]*DlmsValueRequest, 1)
	val := new(DlmsValueRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 4
	vals[0] = val

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep := msg.Data.(DlmsResponse)
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	entriesInUse := data.GetDoubleLongUnsigned()
	t.Logf("profile entries in use: %d", entriesInUse)

	vals = make([]*DlmsValueRequest, 1)

	// read last 10 entries

	val = new(DlmsValueRequest)
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

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep = msg.Data.(DlmsResponse)
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

	aconn.Close()
}

func TestX__profileRead_timeInterval(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect tcp: %s", msg.Err)
		return
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("cannot connect app: %s", msg.Err)
		return
	}
	t.Logf("application connected")
	defer dconn.Close()
	aconn := msg.Data.(*AppConn)

	// profile entries in use

	t.Logf("read profile entries ...")
	vals := make([]*DlmsValueRequest, 1)
	val := new(DlmsValueRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 4
	vals[0] = val

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep := msg.Data.(DlmsResponse)
	dataAccessResult := rep.DataAccessResultAt(0)
	if 0 != dataAccessResult {
		t.Fatalf("data access result: %d", dataAccessResult)
	}
	data := rep.DataAt(0)
	entriesInUse := data.GetDoubleLongUnsigned()
	t.Logf("profile entries in use: %d", entriesInUse)

	vals = make([]*DlmsValueRequest, 1)

	// read last 10 entries

	val = new(DlmsValueRequest)
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

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep = msg.Data.(DlmsResponse)
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

	vals = make([]*DlmsValueRequest, 1)

	val = new(DlmsValueRequest)
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

	aconn.GetRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf("read failed: %s", msg.Err)
		return
	}
	rep = msg.Data.(DlmsResponse)
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

	aconn.Close()
}
