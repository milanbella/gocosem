package gocosem

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	DebugEnabled = testing.Verbose()
	Log.SetPrefix("")
	os.Exit(m.Run())
}
