package logging

import (
	"io"
	"log"
	"os"
)

func NewDebugLogger(enabled bool) *log.Logger {
	writer := io.Discard
	if enabled {
		writer = os.Stdout
	}

	return log.New(writer, "DEBUG: ", log.Lshortfile)
}

func NewInfoLogger() *log.Logger {
	return log.New(os.Stdout, "INFO: ", log.Default().Flags())
}
