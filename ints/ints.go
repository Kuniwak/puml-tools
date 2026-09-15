package ints

import (
	"strconv"
	"strings"
)

func Replaced(t []int, i, v int) []int {
	out := append([]int(nil), t...)
	out[i] = v
	return out
}

func Key(t []int) string {
	var b strings.Builder
	for i, v := range t {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(v))
	}
	return b.String()
}
