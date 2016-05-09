package gocosem

import "flag"

var (
	realMeter bool
	meterIp   string
	hdlcDebug bool
)

func init() {
	Log.SetPrefix("")

	var lvl string
	flag.BoolVar(&realMeter, "real", false, "test against real meter")
	flag.StringVar(&meterIp, "ip", "172.16.123.182", "meter ip address")
	flag.StringVar(&lvl, "log", "INFO", "log level [ALL|TRACE|DEBUG|INFO|WARN|ERROR|FATAL|OFF]")
	flag.BoolVar(&hdlcDebug, "hdlcDebug", false, "dsplay hdlc debug messages")
	flag.Parse()

	SetLogLevel(lvl)
}
