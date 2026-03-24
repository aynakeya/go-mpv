//go:build !cgo
// +build !cgo

package mpv

import (
	"errors"
	"fmt"
)

var ErrCGODisabled = errors.New("mpv backend unavailable without cgo (failed to load libmpv)")

type Error int

const (
	ERROR_SUCCESS              Error = 0
	ERROR_EVENT_QUEUE_FULL     Error = -1
	ERROR_NOMEM                Error = -2
	ERROR_UNINITIALIZED        Error = -3
	ERROR_INVALID_PARAMETER    Error = -4
	ERROR_OPTION_NOT_FOUND     Error = -5
	ERROR_OPTION_FORMAT        Error = -6
	ERROR_OPTION_ERROR         Error = -7
	ERROR_PROPERTY_NOT_FOUND   Error = -8
	ERROR_PROPERTY_FORMAT      Error = -9
	ERROR_PROPERTY_UNAVAILABLE Error = -10
	ERROR_PROPERTY_ERROR       Error = -11
	ERROR_COMMAND              Error = -12
	ERROR_LOADING_FAILED       Error = -13
	ERROR_AO_INIT_FAILED       Error = -14
	ERROR_VO_INIT_FAILED       Error = -15
	ERROR_NOTHING_TO_PLAY      Error = -16
	ERROR_UNKNOWN_FORMAT       Error = -17
	ERROR_UNSUPPORTED          Error = -18
	ERROR_NOT_IMPLEMENTED      Error = -19
	MPV_ERROR_GENERIC          Error = -20
)

func ErrorString(err Error) string {
	if ensureBackend() == nil && backend.errorString != nil {
		if p := backend.errorString(int32(err)); p != nil {
			return goString(p)
		}
	}
	return err.Error()
}

func newError(err int32) error {
	if err >= 0 {
		return nil
	}
	return Error(err)
}

func (m Error) Error() string {
	if m == ERROR_NOT_IMPLEMENTED {
		return "not implemented (nocgo backend)"
	}
	return fmt.Sprintf("mpv error (%d)", int(m))
}
