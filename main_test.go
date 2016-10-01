package gocosem

import (
	"flag"
	"time"
)

var (
	realMeter bool
	meterIp   string
	timeout   time.Duration
)

func init() {
	Log.SetPrefix("")

	var lvl string
	flag.BoolVar(&realMeter, "real", false, "test against real meter")
	flag.StringVar(&meterIp, "ip", "172.16.123.182", "meter ip address")
	flag.DurationVar(&timeout, "t", 15*time.Second, "timeout")
	flag.StringVar(&lvl, "log", "INFO", "log level [ALL|TRACE|DEBUG|INFO|WARN|ERROR|FATAL|OFF]")
	flag.BoolVar(&HdlcDebug, "hdlcDebug", false, "dsplay hdlc debug messages")
	flag.Parse()

	SetLogLevel(lvl)
}
