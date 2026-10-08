package postingpolicy

import "time"

func ValidLength(value string, min, max int) bool {
	return len(value) >= min && len(value) <= max
}

func AvailableAt(created time.Time, minutes int) time.Time {
	if minutes <= 0 {
		return time.Time{}
	}
	return created.Add(time.Duration(minutes) * time.Minute)
}
