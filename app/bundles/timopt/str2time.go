package timeopt

import (
	"errors"
	"time"
)

var formats = []string{
	time.DateTime,
	"2006-01-02",
	"2006/01/02 15:04:05",
	"2006-01-02",
	"2006/01/02",
	time.RFC3339,
	time.RFC3339Nano,
	time.DateOnly,
	time.TimeOnly,
	time.Layout,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
	time.DateTime,
}

func Str2Time(input string) time.Time {
	res, _ := Str2TimeE(input)
	return res
}
func Str2TimeE(input string) (time.Time, error) {
	var t time.Time
	var err error
	for _, layout := range formats {
		t, err = time.Parse(layout, input)
		if err == nil {
			return t, nil
		}
	}

	return t, errors.New("unable to parse time")
}
