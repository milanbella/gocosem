package gocosem

import (
	"log"
	"os"
)

func getErrorLogger() *log.Logger {
	return log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
}

func getDebugLogger() *log.Logger {
	return log.New(os.Stderr, "DEBUG ", log.Ldate|log.Ltime|log.Lshortfile)
}
