package mpv

/*
#include <mpv/client.h>
*/
import "C"

import (
	"fmt"
)

type Error int

const (
	ERROR_SUCCESS              Error = C.MPV_ERROR_SUCCESS
	ERROR_EVENT_QUEUE_FULL     Error = C.MPV_ERROR_EVENT_QUEUE_FULL
	ERROR_NOMEM                Error = C.MPV_ERROR_NOMEM
	ERROR_UNINITIALIZED        Error = C.MPV_ERROR_UNINITIALIZED
	ERROR_INVALID_PARAMETER    Error = C.MPV_ERROR_INVALID_PARAMETER
	ERROR_OPTION_NOT_FOUND     Error = C.MPV_ERROR_OPTION_NOT_FOUND
	ERROR_OPTION_FORMAT        Error = C.MPV_ERROR_OPTION_FORMAT
	ERROR_OPTION_ERROR         Error = C.MPV_ERROR_OPTION_ERROR
	ERROR_PROPERTY_NOT_FOUND   Error = C.MPV_ERROR_PROPERTY_NOT_FOUND
	ERROR_PROPERTY_FORMAT      Error = C.MPV_ERROR_PROPERTY_FORMAT
	ERROR_PROPERTY_UNAVAILABLE Error = C.MPV_ERROR_PROPERTY_UNAVAILABLE
	ERROR_PROPERTY_ERROR       Error = C.MPV_ERROR_PROPERTY_ERROR
	ERROR_COMMAND              Error = C.MPV_ERROR_COMMAND
	ERROR_LOADING_FAILED       Error = C.MPV_ERROR_LOADING_FAILED
	ERROR_AO_INIT_FAILED       Error = C.MPV_ERROR_AO_INIT_FAILED
	ERROR_VO_INIT_FAILED       Error = C.MPV_ERROR_VO_INIT_FAILED
	ERROR_NOTHING_TO_PLAY      Error = C.MPV_ERROR_NOTHING_TO_PLAY
	ERROR_UNKNOWN_FORMAT       Error = C.MPV_ERROR_UNKNOWN_FORMAT
	ERROR_UNSUPPORTED          Error = C.MPV_ERROR_UNSUPPORTED
	ERROR_NOT_IMPLEMENTED      Error = C.MPV_ERROR_NOT_IMPLEMENTED
	MPV_ERROR_GENERIC          Error = C.MPV_ERROR_GENERIC
)

// const char *mpv_error_string(int error);
func ErrorString(err Error) string {
	return C.GoString(C.mpv_error_string(C.int(err)))
}

func newError(err C.int) error {
	if err == C.MPV_ERROR_SUCCESS {
		return nil
	}
	return Error(err)
}

func (m Error) Error() string {
	return fmt.Sprintf("%s (%d)", ErrorString(m), int(m))
}
