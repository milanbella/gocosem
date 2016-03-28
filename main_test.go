package gocosem

import (
	"flag"
	"os"
	"testing"
)

var (
	realMeter bool
	meterIp   string
	logLevel  int
)

func init() {
	var _logLevel string
	flag.BoolVar(&realMeter, "real", false, "test against real meter")
	flag.StringVar(&meterIp, "ip", "172.16.123.182", "meter ip address")
	flag.StringVar(&_logLevel, "log", "INFO", "log level [ALL|DEBUG|FATAL|INFO|OFF]")
	flag.Parse()

	/*
		LOG_LEVEL_ALL   = 8
		LOG_LEVEL_DEBUG = 7
		LOG_LEVEL_ERROR = 6
		LOG_LEVEL_FATAL = 5
		LOG_LEVEL_INFO  = 4
		LOG_LEVEL_OFF   = 3
		LOG_LEVEL_TRACE = 2
		LOG_LEVEL_WARN  = 1
	*/

	if "ALL" == _logLevel {
		logLevel = LOG_LEVEL_ALL
	} else if "DEBUG" == _logLevel {
		logLevel = LOG_LEVEL_DEBUG
	} else if "ERROR" == _logLevel {
		logLevel = LOG_LEVEL_ERROR
	} else if "FATAL" == _logLevel {
		logLevel = LOG_LEVEL_FATAL
	} else if "INFO" == _logLevel {
		logLevel = LOG_LEVEL_INFO
	} else if "OFF" == _logLevel {
		logLevel = LOG_LEVEL_OFF
	} else {
		panic("incorrect value of command lone flag 'logLevel'")
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	DebugEnabled = testing.Verbose()
	Log.SetPrefix("")
	os.Exit(m.Run())
}
