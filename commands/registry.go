package commands

import (
	"errors"
	"fmt"
)

var ErrUnknownCommand = errors.New("unknown command")

func Register[T any](name string, proc func(options T, args []string) error) {
	registry[name] = func(ctx Context) error {
		var options T
		ctx.BindOptions(&options)
		return proc(options, ctx.Args())
	}
}

func Execute(context Context) error {
	cmd := registry[context.Name()]
	if cmd == nil {
		return fmt.Errorf("%s: %w", context.Name(), ErrUnknownCommand)
	}
	return cmd(context)
}

type Context interface {
	Name() string
	Args() []string
	BindOptions(options any)
}

var registry = map[string]func(ctx Context) error{}
