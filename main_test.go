package gocosem

import (
	"flag"
	"os"
	"testing"
)

var (
	realMeter bool = true
	meterIp   string
)

func init() {
	flag.BoolVar(&realMeter, "real", false, "test against real meter")
	flag.StringVar(&meterIp, "ip", "172.16.123.182", "meter ip address")
	flag.Parse()

}

func TestMain(m *testing.M) {
	flag.Parse()
	Log.SetPrefix("")
	os.Exit(m.Run())
}
