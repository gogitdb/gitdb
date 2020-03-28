package gitdb

import (
	"fmt"
	golog "log"
	"path/filepath"
	"runtime"
	"time"
)

var logger *golog.Logger
var verbosity LogLevel

//LogLevel is used to set verbosity of GitDB
type LogLevel int

const (
	//LogLevelNone - log nothing
	LogLevelNone LogLevel = 0
	//LogLevelError - logs only errors
	LogLevelError LogLevel = 1
	//LogLevelWarning  - logs warning and errors
	LogLevelWarning LogLevel = 2
	//LogLevelTest - logs only debug messages
	LogLevelTest LogLevel = 3
	//LogLevelInfo - logs info, warining and errors
	LogLevelInfo LogLevel = 4
)

//SetLogLevel sets log level
func SetLogLevel(l LogLevel) {
	verbosity = l
}

//SetLogger sets Logger
func SetLogger(l *golog.Logger) {
	logger = l
}

func printlog(message string) {
	if logger != nil {
		logger.Println(message)
	} else {
		println("[" + time.Now().Format("2006-01-02-15:04:05.000000") + "] " + message)
	}
}

func log(message string) {
	if verbosity >= LogLevelInfo {
		printlog(message)
	}
}

func logError(message string) {
	if verbosity >= LogLevelError {
		_, fn, line, _ := runtime.Caller(1)
		printlog(fmt.Sprintf("ERROR: %s | %s:%d", message, fn, line))
	}
}

func logWarning(message string) {
	if verbosity >= LogLevelWarning {
		_, fn, line, _ := runtime.Caller(1)
		printlog(fmt.Sprintf("WARNING: %s | %s:%d", message, fn, line))
	}
}

func logTest(message string) {
	if verbosity == LogLevelTest {
		_, fn, line, _ := runtime.Caller(1)
		printlog(fmt.Sprintf("DEBUG: %s (%s:%d)", message, filepath.Base(fn), line))
	}
}
