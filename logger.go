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
	DebugEnabled = true
	Log          = log.New(os.Stderr, "[gocosem] ", log.Lmicroseconds)
)

const (
	LOG_LEVEL_ALL   = 8
	LOG_LEVEL_DEBUG = 7
	LOG_LEVEL_ERROR = 6
	LOG_LEVEL_FATAL = 5
	LOG_LEVEL_INFO  = 4
	LOG_LEVEL_OFF   = 3
	LOG_LEVEL_TRACE = 2
	LOG_LEVEL_WARN  = 1
)

func debugLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_DEBUG {
		Log.Printf("DEBUG: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func errorLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_ERROR {
		Log.Printf("ERROR: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func fatalLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_FATAL {
		Log.Printf("FATAL: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func infoLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_INFO {
		Log.Printf("INFO: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func traceLog(f string, a ...interface{}) {
	Log.Printf("TRACE: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
}

func warnLog(f string, a ...interface{}) {
	Log.Printf("WARN: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
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
