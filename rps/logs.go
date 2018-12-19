package rps

import (
	"io"
	"log"
)

var (
	Verbose *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func LogsInit(
	verboseHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer,
) {
	Verbose = log.New(verboseHandle,
		"VERBOSE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
