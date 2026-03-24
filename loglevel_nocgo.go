//go:build !cgo
// +build !cgo

package mpv

type LogLevel int

const (
	LOG_LEVEL_NONE  LogLevel = 0
	LOG_LEVEL_FATAL LogLevel = 10
	LOG_LEVEL_ERROR LogLevel = 20
	LOG_LEVEL_WARN  LogLevel = 30
	LOG_LEVEL_INFO  LogLevel = 40
	LOG_LEVEL_V     LogLevel = 50
	LOG_LEVEL_DEBUG LogLevel = 60
	LOG_LEVEL_TRACE LogLevel = 70
)

var LOG_LEVEL_STRING = map[LogLevel]string{
	LOG_LEVEL_NONE:  "no",
	LOG_LEVEL_FATAL: "fatal",
	LOG_LEVEL_ERROR: "error",
	LOG_LEVEL_WARN:  "warn",
	LOG_LEVEL_INFO:  "info",
	LOG_LEVEL_V:     "v",
	LOG_LEVEL_DEBUG: "debug",
	LOG_LEVEL_TRACE: "trace",
}
