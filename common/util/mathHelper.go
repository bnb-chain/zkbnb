package util

import "strconv"

func Uint64ToString(a uint64) string {
	return strconv.FormatUint(a, Base)
}

func Int64ToString(a int64) string {
	return strconv.FormatInt(a, Base)
}

func StringToUint64(a string) (uint64, error) {
	return strconv.ParseUint(a, Base, 64)
}

func StringToInt64(a string) (int64, error) {
	return strconv.ParseInt(a, Base, 64)
}

func MinInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
