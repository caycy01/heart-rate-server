package utils

import "time"

func CurrentMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func MillisToTime(millis int64) time.Time {
	return time.Unix(0, millis*int64(time.Millisecond))
}

func TimeToMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
