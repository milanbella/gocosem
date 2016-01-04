package gocosem

import (
	"fmt"
	"testing"
)

const ipAddr = "172.16.123.182"
const port = 4059

func TestX_meter182_TcpConnect(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, ipAddr, 4059)
	msg := <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.err))
	}
	t.Logf("transport connected")
	dconn := msg.data.(*DlmsConn)
	dconn.Close()
}

func TestX_meter182_AppConnect(t *testing.T) {

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, ipAddr, 4059)
	msg := <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.err))
	}
	t.Logf("transport connected")
	dconn := msg.data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.err {
		t.Fatalf(fmt.Sprintf("%s\n", msg.err))
	}
	t.Logf("application connected")
	aconn := msg.data.(*AppConn)
	aconn.Close()
}
