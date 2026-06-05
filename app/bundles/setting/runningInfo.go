package setting

import (
	"time"
)

var start time.Time
var unitTime time.Time

func init() {
	start = time.Now()
	unitTime = time.Now()
}

// GetRunTime returns the duration since application startup.
func GetRunTime() time.Duration {
	return time.Now().Sub(start)
}

// GetUnitTime returns the duration since the previous GetUnitTime call.
func GetUnitTime() time.Duration {
	unit := time.Now().Sub(unitTime)
	unitTime = time.Now()
	return unit
}

// GetStart returns the application startup time.
func GetStart() time.Time {
	return start
}
