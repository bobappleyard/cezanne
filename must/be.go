package must

func Be[T any](x T, err error) T {
	Succeed(err)
	return x
}

func Be2[T, U any](x T, y U, err error) (T, U) {
	Succeed(err)
	return x, y
}

func Succeed(err error) {
	if err != nil {
		panic(err)
	}
}
