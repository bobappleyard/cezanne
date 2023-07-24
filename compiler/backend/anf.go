package backend

type pkg struct {
}

type method struct {
	name       string
	argc, varc int
	steps      []step
}

type step interface {
	step()
}

type variable int

func (v variable) offset(x int) variable {
	return v + variable(x)
}

type stringStep struct {
	val  string
	into variable
}

type intStep struct {
	val  int
	into variable
}

type localStep struct {
	from, into variable
}

type fieldStep struct {
	from  variable
	field int
	into  variable
}

type globalStep struct {
	from string
	into variable
}

type globalStoreStep struct {
	into   int
	object variable
}

type createStep struct {
	into    variable
	methods []method
	fields  []variable
}

type returnStep struct {
	val variable
}

type callMethodStep struct {
	into   variable
	object variable
	method string
	params []variable
}

type callFunctionStep struct {
	into   variable
	method string
	params []variable
}

func (stringStep) step()       {}
func (intStep) step()          {}
func (localStep) step()        {}
func (fieldStep) step()        {}
func (globalStep) step()       {}
func (returnStep) step()       {}
func (createStep) step()       {}
func (callMethodStep) step()   {}
func (callFunctionStep) step() {}
func (globalStoreStep) step()  {}

func (b *method) nextVar() variable {
	res := variable(b.varc)
	b.varc++
	return res
}