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

	aconn.getRquest(ch, 10000, 0, true, vals)
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

	aconn.getRquest(ch, 10000, 0, true, vals)
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
