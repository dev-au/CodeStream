package plugins

import (
	"time"
)

type DateFormat time.Time

type DateTimeFormat time.Time

func (m DateFormat) MarshalJSON() ([]byte, error) {
	t := time.Time(m)
	return []byte("\"" + t.Format(time.RFC3339) + "\""), nil
}

func (m DateFormat) GetTime() time.Time {
	t := time.Time(m)
	if t.IsZero() {
		return t
	}

	t = t.In(location)
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0,
		t.Location(),
	)
}

func (m DateTimeFormat) MarshalJSON() ([]byte, error) {
	t := time.Time(m)
	return []byte("\"" + t.Format(time.RFC3339) + "\""), nil
}

func (m DateTimeFormat) GetTime() time.Time {
	t := time.Time(m)
	if t.IsZero() {
		return t
	}
	return t.In(location)
}

func (m DateFormat) GetTimePtr() *time.Time {
	t := m.GetTime()
	if t.IsZero() {
		return nil
	}
	return &t
}

func (m DateTimeFormat) GetTimePtr() *time.Time {
	t := m.GetTime()
	if t.IsZero() {
		return nil
	}
	return &t
}
