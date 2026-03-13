package telemetry

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	enabledOnce sync.Once
	enabled     bool
)

func Enabled() bool {
	enabledOnce.Do(func() {
		v, ok := os.LookupEnv("FRESH_PROFILE")
		if !ok {
			enabled = false
			return
		}
		parsed, err := strconv.ParseBool(v)
		enabled = err == nil && parsed
	})
	return enabled
}

func Short(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}
