package runtime

import "reflect"

type Env struct {
	nextClass ClassID
	space     Space
	pkgs      map[string]Object
	wrappers  map[reflect.Type]wrappingProtocol
	load      func(name string) (*Unit, error)
}

func New(load func(name string) (*Unit, error)) *Env {
	e := &Env{
		space:    Space{methods: map[string]MethodID{}},
		wrappers: make(map[reflect.Type]wrappingProtocol),
		pkgs:     map[string]Object{},
		load:     load,
	}
	e.init()
	return e
}

func (e *Env) Method(name string) MethodID {
	return e.space.Method(name)
}

func (e *Env) Import(name string) (Object, error) {
	pkg := e.pkgs[name]
	if pkg != nil {
		return pkg, nil
	}

	unit, err := e.load(name)
	if err != nil {
		return nil, err
	}

	pkg, err = unit.exec(e.Process())
	if err != nil {
		return nil, err
	}

	e.pkgs[name] = pkg
	return pkg, nil
}

func (e *Env) Process() *Process {
	return &Process{env: e}
}

// Built in methods

const (
	callMethod MethodID = iota
	importMethod
	extendMethod
)

func (e *Env) init() {
	e.Method("call")
	e.Method("import")
	e.Method("extend")

	e.initTypes()
}

func (e *Env) initTypes() {
	e.wrappers[reflect.TypeOf(0)] = &wrapperFns{
		func(x reflect.Value) Object {
			return FromInt(int(x.Int()))
		},
		func(x Object) (reflect.Value, error) {
			xv, ok := x.(*intObject)
			if !ok {
				return reflect.Value{}, ErrWrongType
			}
			return reflect.ValueOf(xv.value), nil
		},
	}
	e.wrappers[reflect.TypeOf("")] = &wrapperFns{
		func(x reflect.Value) Object {
			return FromString(x.String())
		},
		func(x Object) (reflect.Value, error) {
			xv, ok := x.(*stringObject)
			if !ok {
				return reflect.Value{}, ErrWrongType
			}
			return reflect.ValueOf(xv.value), nil
		},
	}
}
