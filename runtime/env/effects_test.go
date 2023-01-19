package env

/*
func TestEnterContext(t *testing.T) {
	e := &Env{
		code: []byte{
			createOp, 0, 0, 0, 0, 0,
			storeOp, 2,
			createOp, 1, 0, 0, 0, 0,
			storeOp, 3,
			globalOp, 0, 0, 0, 0,
			callOp, 0, 0, 0, 0, 0,

			// expression to be handled
			100: naturalOp, 2, 0, 0, 0,
			retOp,
		},
		classes: make([]format.Class, 3),
		globals: []Object{
			&standardObject{classID: 2},
		},
		methods: []format.Binding{
			{ClassID: -1},
			{ClassID: 1, Start: 100},
			{ClassID: 2, Start: -1},
		},
		extern: []func(*Process){
			func(p *Process) {
				p.EnterContext(p.Arg(0), p.Arg(1))
			},
		},
	}

	p := &Process{env: e}

	p.run()
	assert.Equal(t, Int(2), p.value)
}

func TestTriggerEffect(t *testing.T) {
	e := &Env{
		code: []byte{
			createOp, 0, 0, 0, 0, 0,
			storeOp, 2,
			createOp, 1, 0, 0, 0, 0,
			storeOp, 3,
			globalOp, 0, 0, 0, 0,
			callOp, 0, 0, 0, 0, 0,

			// expression to be handled
			100: createOp, 5, 0, 0, 0, 0,
			storeOp, 2,
			globalOp, 0, 0, 0, 0,
			callOp, 1, 0, 0, 0, 0,

			// trigger
			150: loadOp, 2,
			callOp, 4, 0, 0, 0, 0,

			// handler
			200: naturalOp, 2, 0, 0, 0,
			retOp,
		},
		classes: make([]format.Class, 6),
		globals: []Object{
			&standardObject{classID: 2},
		},
		methods: []format.Binding{
			{ClassID: -1},
			{ClassID: 1, Start: 100},
			{ClassID: 2, Start: -1},
			{ClassID: 2, Start: -2},
			{ClassID: 0, Start: 200},
			{ClassID: 5, Start: 150},
		},
		extern: []func(*Process){
			func(p *Process) {
				p.EnterContext(p.Arg(0), p.Arg(1))
			},
			func(p *Process) {
				p.TriggerEffect(4, p.Arg(0))
			},
		},
	}

	p := &Process{env: e}

	p.run()
	assert.Equal(t, Int(2), p.value)
}

func TestFastAbort(t *testing.T) {
	e := &Env{
		code: []byte{
			createOp, 0, 0, 0, 0, 0,
			storeOp, 2,
			createOp, 1, 0, 0, 0, 0,
			storeOp, 3,
			globalOp, 0, 0, 0, 0,
			callOp, 0, 0, 0, 0, 0,

			// expression to be handled
			100: naturalOp, 2, 0, 0, 0,
			storeOp, 2,
			naturalOp, 127, 0, 0, 0,
			storeOp, 3,
			createOp, 3, 0, 0, 0, 0,
			callOp, 2, 0, 0, 0, 2,
			127: naturalOp, 10, 0, 0, 0,
			retOp,

			// method called by expression
			150: createOp, 7, 0, 0, 0, 0,
			storeOp, 2,
			globalOp, 0, 0, 0, 0,
			callOp, 1, 0, 0, 0, 0,

			// trigger
			175: loadOp, 2,
			callOp, 4, 0, 0, 0, 0,

			// handler
			200: naturalOp, 2, 0, 0, 0,
			storeOp, 2,
			globalOp, 0, 0, 0, 0,
			callOp, 4, 0, 0, 0, 0,
		},
		classes: make([]format.Class, 8),
		globals: []Object{
			&standardObject{classID: 2},
		},
		methods: []format.Binding{
			{ClassID: -1},
			{ClassID: 1, Start: 100},
			{ClassID: 2, Start: -1},
			{ClassID: 2, Start: -2},
			{ClassID: 0, Start: 200},
			{ClassID: 3, Start: 150},
			{ClassID: 2, Start: -3},
			{ClassID: 7, Start: 175},
		},
		extern: []func(*Process){
			func(p *Process) {
				p.EnterContext(p.Arg(0), p.Arg(1))
			},
			func(p *Process) {
				p.TriggerEffect(4, p.Arg(0))
			},
			func(p *Process) {
				p.FastAbortHandler(p.Arg(0))
			},
		},
	}

	p := &Process{env: e}

	p.run()
	assert.Equal(t, Int(2), p.value)
}

func TestFastResume(t *testing.T) {
	e := &Env{
		code: []byte{
			createOp, 0, 0, 0, 0, 0,
			storeOp, 2,
			createOp, 1, 0, 0, 0, 0,
			storeOp, 3,
			globalOp, 0, 0, 0, 0,
			callOp, 0, 0, 0, 0, 0,

			// expression to be handled
			100: naturalOp, 2, 0, 0, 0,
			storeOp, 2,
			naturalOp, 127, 0, 0, 0,
			storeOp, 3,
			createOp, 3, 0, 0, 0, 0,
			callOp, 2, 0, 0, 0, 2,
			127: naturalOp, 10, 0, 0, 0,
			retOp,

			// method called by expression
			150: createOp, 7, 0, 0, 0, 0,
			storeOp, 2,
			globalOp, 0, 0, 0, 0,
			callOp, 1, 0, 0, 0, 0,

			// trigger
			175: loadOp, 2,
			callOp, 4, 0, 0, 0, 0,

			// handler
			200: naturalOp, 2, 0, 0, 0,
			storeOp, 2,
			globalOp, 0, 0, 0, 0,
			callOp, 4, 0, 0, 0, 0,
		},
		classes: make([]format.Class, 8),
		globals: []Object{
			&standardObject{classID: 2},
		},
		methods: []format.Binding{
			{ClassID: -1},
			{ClassID: 1, Start: 100},
			{ClassID: 2, Start: -1},
			{ClassID: 2, Start: -2},
			{ClassID: 0, Start: 200},
			{ClassID: 3, Start: 150},
			{ClassID: 2, Start: -3},
			{ClassID: 7, Start: 175},
		},
		extern: []func(*Process){
			func(p *Process) {
				p.EnterContext(p.Arg(0), p.Arg(1))
			},
			func(p *Process) {
				p.TriggerEffect(4, p.Arg(0))
			},
			func(p *Process) {
				p.FastResumeHandler(p.Arg(0))
			},
		},
	}

	p := &Process{env: e}

	p.run()
	assert.Equal(t, Int(10), p.value)
}

func TestReifyContext(t *testing.T) {
	e := &Env{
		code: []byte{
			createOp, 0, 0, 0, 0, 0,
			storeOp, 2,
			createOp, 1, 0, 0, 0, 0,
			storeOp, 3,
			globalOp, 0, 0, 0, 0,
			callOp, 0, 0, 0, 0, 0,

			// expression to be handled
			100: naturalOp, 2, 0, 0, 0,
			storeOp, 2,
			naturalOp, 127, 0, 0, 0,
			storeOp, 3,
			createOp, 3, 0, 0, 0, 0,
			callOp, 2, 0, 0, 0, 2,
			127: naturalOp, 10, 0, 0, 0,
			retOp,

			// method called by expression
			150: createOp, 7, 0, 0, 0, 0,
			storeOp, 2,
			globalOp, 0, 0, 0, 0,
			callOp, 1, 0, 0, 0, 0,

			// trigger
			175: loadOp, 2,
			callOp, 4, 0, 0, 0, 0,

			// handler
			200: createOp, 8, 0, 0, 0, 0,
			storeOp, 2,
			globalOp, 0, 0, 0, 0,
			callOp, 4, 0, 0, 0, 0,

			// handler body
			250: loadOp, 2,
			retOp,
		},
		classes: make([]format.Class, 9),
		globals: []Object{
			&standardObject{classID: 2},
		},
		methods: []format.Binding{
			{ClassID: -1},
			{ClassID: 1, Start: 100},
			{ClassID: 2, Start: -1},
			{ClassID: 2, Start: -2},
			{ClassID: 0, Start: 200},
			{ClassID: 3, Start: 150},
			{ClassID: 2, Start: -3},
			{ClassID: 7, Start: 175},
			{ClassID: 8, Start: 250},
		},
		extern: []func(*Process){
			func(p *Process) {
				p.EnterContext(p.Arg(0), p.Arg(1))
			},
			func(p *Process) {
				p.TriggerEffect(4, p.Arg(0))
			},
			func(p *Process) {
				p.ReifyHandlerContext(p.Arg(0))
			},
		},
	}

	p := &Process{env: e}

	p.run()
	assert.Equal(t, Int(10), p.value)
}
*/
