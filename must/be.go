package must

func Be[T any](x T, err error) T {
	Succeed(err)
	return x
}

func Succeed(err error) {
	if err != nil {
		panic(err)
	}
}
