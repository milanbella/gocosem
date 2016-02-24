package gocosem

import (
	"io"
	"log"
	"os"
)

var errorLogWriter io.Writer = os.Stderr
var debugLogWriter io.Writer = os.Stderr

func getErrorLogger() *log.Logger {
	return log.New(errorLogWriter, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
}

func getDebugLogger() *log.Logger {
	return log.New(debugLogWriter, "DEBUG ", log.Ldate|log.Ltime|log.Lshortfile)
}

func SetErrorLogger(out io.Writer) {
	errorLogWriter = out
}

func SetDebugLogger(out io.Writer) {
	debugLogWriter = out
}
