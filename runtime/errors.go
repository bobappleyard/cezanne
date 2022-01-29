package runtime

import "errors"

var (
	ErrUnknownMethod    = errors.New("unknown method")
	ErrWrongArgCount    = errors.New("wrong number of arguments")
	ErrUnknownMember    = errors.New("unknown member")
	ErrWrongMagicNumber = errors.New("wrong magic number")
	ErrWrongUnitVersion = errors.New("wrong unit version")
	ErrBadReturn        = errors.New("host method returns wrong number of results")
)
