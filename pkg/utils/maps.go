package utils

import (
	"cmp"
)

func Values[K cmp.Ordered, V any](m map[K]V) []V {
	var vs = make([]V, 0, len(m))

	for _, v := range m {
		vs = append(vs, v)
	}

	return vs
}
