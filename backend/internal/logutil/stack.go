package logutil

import (
	"log"
	"os"
	"runtime/debug"
)

// Fatalf logs the message with stack trace and exits with status 1.
func Fatalf(format string, args ...interface{}) {
	log.Printf(format, args...)
	debug.PrintStack()
	os.Exit(1)
}

// Errorf logs the message with stack trace but does not exit.
func Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
	debug.PrintStack()
}
