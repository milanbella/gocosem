package gocosem

import (
	"log"
	"os"
)

func getLogger() *log.Logger {
	return log.New(os.Stdin, "cosemgo", log.Ldate|log.Ltime|log.Lshortfile)
}
