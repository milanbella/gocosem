package gocosem

import (
	"fmt"
	"testing"
)

func TestX__profileRead(t *testing.T) {

	var FNAME string = "profileRead()"

	ch := make(DlmsChannel)
	TcpConnect(ch, 10000, "172.16.123.182", 4059)
	msg := <-ch
	if nil != msg.Err {
		errorLog.Printf(fmt.Sprintf("%s: %v\n", FNAME, msg.Err))
		return
	}
	debugLog.Printf(fmt.Sprintf("%s: transport connected\n", FNAME))
	dconn := msg.Data.(*DlmsConn)

	dconn.AppConnectWithPassword(ch, 10000, 01, 01, "12345678")
	msg = <-ch
	if nil != msg.Err {
		errorLog.Printf(fmt.Sprintf("%s: %v\n", FNAME, msg.Err))
		return
	}
	debugLog.Printf(fmt.Sprintf("%s: application connected\n", FNAME))
	defer dconn.Close()
	aconn := msg.Data.(*AppConn)

	vals := make([]*DlmsValueRequest, 1)

	val := new(DlmsValueRequest)
	val.ClassId = 7
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 7
	//vals[0] = val

	val = new(DlmsValueRequest)
	val.ClassId = 7
	//0100630100FF
	val.InstanceId = &DlmsOid{1, 0, 99, 1, 0, 255}
	val.AttributeId = 2
	vals[0] = val

	aconn.getRquest(ch, 10000, 0, true, vals)
	msg = <-ch
	if nil != msg.Err {
		errorLog.Printf(fmt.Sprintf("%s: %v\n", FNAME, msg.Err))
		return
	}
	debugLog.Printf(fmt.Sprintf("%s: values read\n", FNAME))

	aconn.Close()
}
