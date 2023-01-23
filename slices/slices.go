package slices

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
