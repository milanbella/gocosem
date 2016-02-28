package gocosem

import (
	"fmt"
	"testing"
)

func TestX_meter182_TcpConnect(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.Err))
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)
	dconn.Close()
}

func TestX_meter182_AppConnect(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.Err))
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.Err))
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)
	aconn.Close()
}

func TestX_meter182_get_time(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.Err))
	}
	t.Logf("transport connected")
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.Err))
	}
	t.Logf("application connected")
	aconn := msg.Data.(*AppConn)

	//func (aconn *AppConn) getRquest(ch DlmsChannel, msecTimeout int64, highPriority bool, vals []*DlmsValueRequest) {
	val := new(DlmsRequest)
	val.ClassId = 1
	val.InstanceId = &DlmsOid{0x00, 0x00, 0x2A, 0x00, 0x00, 0xFF}
	val.AttributeId = 0x02
	vals := make([]*DlmsRequest, 1)
	vals[0] = val
	aconn.SendRequest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.Err))
	}
	t.Logf("value read")
	aconn.Close()

}
