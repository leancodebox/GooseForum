package app

import (
	"time"
)

var start time.Time
var unitTime time.Time

func InitStart() {
	start = time.Now()
	unitTime = time.Now()
}

func GetRunTime() time.Duration {
	return time.Now().Sub(start)
}

func GetUnitTime() time.Duration {
	unit := time.Now().Sub(unitTime)
	unitTime = time.Now()
	return unit
}
