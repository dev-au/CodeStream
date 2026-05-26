package plugins

import (
	"fmt"
	"strconv"
)

func RoundTo(a float64, precision int) float64 {
	format := "%." + strconv.Itoa(precision) + "f"

	str := fmt.Sprintf(format, a)
	s, _ := strconv.ParseFloat(str, 64)
	return s
}

func SafeEqualString(s1, s2 *string) bool {
	if s1 == nil || s2 == nil {
		return false
	}
	return *s1 == *s2
}

func Ptr[T any](v T) *T {
	return &v
}
