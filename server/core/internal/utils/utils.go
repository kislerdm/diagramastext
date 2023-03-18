package utils

import (
	"strconv"
)

func MustParseInt(s string) int {
	o, _ := strconv.Atoi(s)
	return o
}

func MustParseFloat32(s string) float32 {
	o, _ := strconv.ParseFloat(s, 10)
	return float32(o)
}
