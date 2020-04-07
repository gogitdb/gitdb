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
var verbositySet bool

//LogLevel is used to set verbosity of GitDB
type LogLevel int

const (
	//LogLevelNone - log nothing
	LogLevelNone LogLevel = iota
	//LogLevelError - logs only errors
	LogLevelError
	//LogLevelWarning  - logs warning and errors
	LogLevelWarning
	//LogLevelTest - logs only debug messages
	LogLevelTest
	//LogLevelInfo - logs info, warining and errors
	LogLevelInfo
)

//getVerbosity defaults verbosity to LogLevelWarning if a verbosity is not set
func getVerbosity() LogLevel {
	if !verbositySet {
		return LogLevelWarning
	}
	return verbosity
}

//SetLogLevel sets log level
func SetLogLevel(l LogLevel) {
	verbosity = l
	verbositySet = true
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
	if getVerbosity() >= LogLevelInfo {
		printlog(message)
	}
}

func logError(message string) {
	if getVerbosity() >= LogLevelError {
		_, fn, line, _ := runtime.Caller(1)
		printlog(fmt.Sprintf("ERROR: %s (%s:%d)", message, filepath.Base(fn), line))
	}
}

func logWarning(message string) {
	if getVerbosity() >= LogLevelWarning {
		_, fn, line, _ := runtime.Caller(1)
		printlog(fmt.Sprintf("WARNING: %s (%s:%d)", message, filepath.Base(fn), line))
	}
}

func logTest(message string) {
	if getVerbosity() == LogLevelTest {
		_, fn, line, _ := runtime.Caller(1)
		printlog(fmt.Sprintf("DEBUG: %s (%s:%d)", message, filepath.Base(fn), line))
	}
}
