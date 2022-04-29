package mpv

/*
#include <mpv/client.h>
*/
import "C"

type LogLevel int

const (
	LOG_LEVEL_NONE  LogLevel = C.MPV_LOG_LEVEL_NONE  /// "no"    - disable absolutely all messages
	LOG_LEVEL_FATAL LogLevel = C.MPV_LOG_LEVEL_FATAL /// "fatal" - critical/aborting errors
	LOG_LEVEL_ERROR LogLevel = C.MPV_LOG_LEVEL_ERROR /// "error" - simple errors
	LOG_LEVEL_WARN  LogLevel = C.MPV_LOG_LEVEL_WARN  /// "warn"  - possible problems
	LOG_LEVEL_INFO  LogLevel = C.MPV_LOG_LEVEL_INFO  /// "info"  - informational message
	LOG_LEVEL_V     LogLevel = C.MPV_LOG_LEVEL_V     /// "v"     - noisy informational message
	LOG_LEVEL_DEBUG LogLevel = C.MPV_LOG_LEVEL_DEBUG /// "debug" - very noisy technical information
	LOG_LEVEL_TRACE LogLevel = C.MPV_LOG_LEVEL_TRACE /// "trace" - extremely noisy
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
