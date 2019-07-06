package gitdb

import (
	"time"
	"runtime"
	"fmt"
	golog "log"
)

var logger *golog.Logger
var verbosity LogLevel

type LogLevel int

const (
	LOGLEVEL_NONE    LogLevel = 0
	LOGLEVEL_ERROR   LogLevel = 1
	LOGLEVEL_WARNING LogLevel = 2
	LOGLEVEL_TEST    LogLevel = 3
	LOGLEVEL_INFO    LogLevel = 4
)

func SetLogLevel(l LogLevel) {
	verbosity = l
}

func printlog(message string){
		if logger != nil {
			logger.Println(message)
		} else {
			println("[" + time.Now().Format("2006-01-02-15:04:05.000000") + "] " + message)
		}
}

func log(message string){
	if verbosity >= LOGLEVEL_INFO{
		printlog(message)
	}
}

func logError(message string){
	if verbosity >= LOGLEVEL_ERROR {
		_, fn, line, _ := runtime.Caller(1)
		printlog(fmt.Sprintf("ERROR: %s | %s:%d",message, fn, line))
	}
}

func logWarning(message string){
	if verbosity >= LOGLEVEL_WARNING {
		_, fn, line, _ := runtime.Caller(1)
		printlog(fmt.Sprintf("WARNING: %s | %s:%d",message, fn, line))
	}
}

func logTest(message string){
	if verbosity == LOGLEVEL_TEST {
		_, fn, line, _ := runtime.Caller(1)
		printlog(fmt.Sprintf("DEBUG: %s",message))
		printlog(fmt.Sprintf("  |__ %s:%d", fn, line))
	}
}