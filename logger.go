package gocosem

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

var (
	DebugEnabled = false
	Log          = log.New(os.Stderr, "[gocosem] ", log.Lmicroseconds)
)

func logf(f string, a ...interface{}) {
	if DebugEnabled {
		Log.Printf("%s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func errlogf(f string, a ...interface{}) {
	if DebugEnabled {
		Log.Printf("ERROR: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

var stripFnPreamble = regexp.MustCompile(`^.*\.(.*)$`)

func funcInfo() string {
	name := "<unknown>"
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		name = stripFnPreamble.ReplaceAllString(runtime.FuncForPC(pc).Name(), "$1")
	}
	file = filepath.Base(file)
	return fmt.Sprintf("%s:%d: %s()", file, line, name)
}
