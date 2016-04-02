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
	Log      = log.New(os.Stderr, "[gocosem] ", log.Lmicroseconds)
	logLevel int
)

const (
	LOG_LEVEL_OFF = iota
	LOG_LEVEL_FATAL
	LOG_LEVEL_ERROR
	LOG_LEVEL_WARN
	LOG_LEVEL_INFO
	LOG_LEVEL_DEBUG
	LOG_LEVEL_TRACE
	LOG_LEVEL_ALL
)

func SetLogLevel(lvl string) {
	switch lvl {
	case "ALL":
		logLevel = LOG_LEVEL_ALL
	case "TRACE":
		logLevel = LOG_LEVEL_TRACE
	case "DEBUG":
		logLevel = LOG_LEVEL_DEBUG
	case "INFO":
		logLevel = LOG_LEVEL_INFO
	case "WARN":
		logLevel = LOG_LEVEL_WARN
	case "ERROR":
		logLevel = LOG_LEVEL_ERROR
	case "FATAL":
		logLevel = LOG_LEVEL_FATAL
	case "OFF":
		logLevel = LOG_LEVEL_OFF
	default:
		panic("invalid log level: " + lvl)
	}
}

func fatalLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_FATAL {
		Log.Printf("FATAL: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func errorLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_ERROR {
		Log.Printf("ERROR: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func warnLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_WARN {
		Log.Printf("WARN: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func infoLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_INFO {
		Log.Printf("INFO: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func debugLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_DEBUG {
		Log.Printf("DEBUG: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
	}
}

func traceLog(f string, a ...interface{}) {
	if logLevel >= LOG_LEVEL_TRACE {
		Log.Printf("TRACE: %s: %s", funcInfo(), fmt.Sprintf(f, a...))
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
