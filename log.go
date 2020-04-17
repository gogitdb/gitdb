package gitdb

import (
	golog "log"

	"github.com/bouggo/log"
)

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

//SetLogLevel sets log level
func SetLogLevel(l LogLevel) {
	log.SetLogLevel(log.Level(l))
}

//SetLogger sets Logger
func SetLogger(l *golog.Logger) {
	log.SetLogger(l)
}
