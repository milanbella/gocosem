package gocosem

import (
	"flag"
	"time"
)

var realMeter bool
var timeout time.Duration

var testMeterIp = "192.168.1.101"

var testHdlcResponseTimeout = time.Duration(1) * time.Hour
var testHdlcCosemWaitTime = time.Duration(5000) * time.Millisecond
var testHdlcSnrmTimeout = time.Duration(45) * time.Second
var testHdlcDiscTimeout = time.Duration(45) * time.Second

func init() {
	Log.SetPrefix("")

	var lvl string
	flag.BoolVar(&realMeter, "real", false, "test against real meter")
	flag.StringVar(&testMeterIp, "ip", "172.16.123.182", "meter ip address")
	flag.DurationVar(&timeout, "t", 15*time.Second, "timeout")
	flag.StringVar(&lvl, "log", "INFO", "log level [ALL|TRACE|DEBUG|INFO|WARN|ERROR|FATAL|OFF]")
	flag.BoolVar(&HdlcDebug, "hdlcDebug", false, "dsplay hdlc debug messages")
	flag.Parse()

	SetLogLevel(lvl)
}
