package go_fritzbox_api

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func Reverse(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}
	return
}

func TimeFromDM(day int, month int) time.Time {
	return time.Date(1, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func TimeFromDMH(day int, month int, hour int) time.Time {
	return time.Date(1, time.Month(month), day, hour, 0, 0, 0, time.UTC)
}

func ToUrlValue(value interface{}) []string {
	var v string

	switch value.(type) {
	case int:
		v = strconv.Itoa(value.(int))
	case string:
		v = value.(string)
	case float32:
		return ToUrlValue(float64(value.(float32)))
	case float64:
		v = strconv.FormatFloat(value.(float64), 'f', 2, 64)
	case time.Month:
		return ToUrlValue(int(value.(time.Month)))
	default:
		panic("unsupported type: " + fmt.Sprintf("%T", value))
	}

	return []string{v}
}

// Pow returns n^m
func Pow(n, m int) int {
	return int(math.Pow(float64(n), float64(m)))
}
