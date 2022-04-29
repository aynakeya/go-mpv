package mpv

import (
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	fmt.Println(ErrorString(ERROR_COMMAND))
	fmt.Println(newError(1))
	fmt.Println(newError(-1) == ERROR_EVENT_QUEUE_FULL)
}
