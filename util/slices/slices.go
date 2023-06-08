package slices

import (
	"sort"

	"golang.org/x/exp/constraints"
)

func FlatMap[T, U any](xs []T, f func(T) []U) []U {
	var res []U
	for _, x := range xs {
		res = append(res, f(x)...)
	}
	return res
}

func Map[T, U any](xs []T, f func(T) U) []U {
	res := make([]U, len(xs))
	for i, x := range xs {
		res[i] = f(x)
	}
	return res
}

func Filter[T any](xs []T, match func(T) bool) []T {
	var res []T
	for _, x := range xs {
		if match(x) {
			res = append(res, x)
		}
	}
	return res
}

func IndexOf[T comparable](xs []T, need T) int {
	for i, x := range xs {
		if x == need {
			return i
		}
	}
	return -1
}

func Unique[T constraints.Ordered](xs []T) []T {
	if len(xs) == 0 {
		return nil
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i] < xs[j]
	})
	res := []T{xs[0]}
	for i, x := range xs[1:] {
		if x == xs[i] {
			continue
		}
		res = append(res, x)
	}
	return res
}
