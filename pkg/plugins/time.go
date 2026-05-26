package plugins

import (
	"time"
)

var location = time.UTC

func SetLocation(loc *time.Location) {
	location = loc
}

func GetNow() time.Time {
	return time.Now().In(location)
}
