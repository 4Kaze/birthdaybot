package common

import "time"

type SystemClock struct{}

func (clock SystemClock) Now() time.Time {
	return time.Now()
}
