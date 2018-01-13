package gocosem

import (
	"fmt"
	"testing"
	"time"
)

func init_TestMeterHdlc_with_sec_5() {
	testMeterIp = "192.168.1.101"
	testHdlcResponseTimeout = time.Duration(1) * time.Hour
	testHdlcCosemWaitTime = time.Duration(5000) * time.Millisecond
	testHdlcSnrmTimeout = time.Duration(45) * time.Second
	testHdlcDiscTimeout = time.Duration(45) * time.Second
}

func TestMeterHdlc_with_sec_5_TcpConnect(t *testing.T) {
	init_TestMeterHdlc_with_sec_5()
	dconn, err := TcpConnect(testMeterIp, 4059)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()
}

func TestMeterHdlc_with_sec_5_HdlcConnect(t *testing.T) {
	init_TestMeterHdlc_with_sec_5()
	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()
}

func TestMeterHdlc_with_sec_5_AppConnect(t *testing.T) {
	init_TestMeterHdlc_with_sec_5()

	dconn, err := HdlcConnect(testMeterIp, 4059, 1, 1, nil, testHdlcResponseTimeout, &testHdlcCosemWaitTime, testHdlcSnrmTimeout, testHdlcDiscTimeout)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("transport connected")
	defer dconn.Close()

	aconn, err := dconn.AppConnectWithSecurity5(01, 01, 4, []uint32{2, 16, 756, 5, 8, 1, 3}, []byte{0x4D, 0x45, 0x4C, 0x00, 0x00, 0x00, 0x00, 0x01}, ")HB+0F04", []byte{0x21, 0x1F, 0x30, 0x24, 0x50, 0x7E, 0x1E, 0xC4, 0xC0, 0xDB, 0xB9, 0x52, 0xC7, 0x0E, 0x7B, 0x3F, 0xF0, 0xA2, 0x96, 0x2B, 0xB8, 0x86, 0x5A, 0xB9, 0xE5, 0x67, 0xA0, 0xC3, 0x81, 0xD6, 0xEB, 0xF5, 0xC3})
	if nil != err {
		t.Fatalf(fmt.Sprintf("%s\n", err))
	}
	t.Logf("application connected")
	defer aconn.Close()
}
