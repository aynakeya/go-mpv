package mpv

import (
	"fmt"
	"testing"
)

func TestEvent(t *testing.T) {
	fmt.Println(EventName(EVENT_SEEK))
}
