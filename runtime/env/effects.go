package env

// func (p *Process) installHandlers(handlers api.Object, next int) {
// 	// shift the activation frame up
// 	p.data[p.frame+2] = Int(AsInt(p.data[p.frame]) + 2)
// 	p.data[p.frame+3] = p.data[p.frame+1]

// 	// install the handler context
// 	p.data[p.frame] = Int(p.frame - next)
// 	p.data[p.frame+1] = handlers

// 	// set context registers
// 	p.context = p.frame
// 	p.frame += 2
// }

// func (p *Process) EnterContext(handlers api.Object, body api.Object) {
// 	p.installHandlers(handlers, p.context)

// 	// enter body
// 	p.value = body
// 	p.callMethod(p.env.callMethodID)
// }

// func (p *Process) TriggerEffect(id format.MethodID, argProvider api.Object) {
// 	for ctx := p.context; ; ctx -= AsInt(p.data[ctx]) {
// 		handlers := p.data[ctx+1]
// 		method := p.getMethod(handlers, id)
// 		if method == nil {
// 			continue
// 		}

// 		// establish fake handler context
// 		p.installHandlers(p.env.memory.Alloc(api.Ref(emptyClass)), ctx)

// 		// enter handler
// 		p.data[p.frame+2] = handlers
// 		p.value = argProvider
// 		p.callMethod(p.env.callMethodID)
// 		break
// 	}
// }

// func (p *Process) FastAbortHandler(value api.Object) {
// 	p.context -= AsInt(p.data[p.context])
// 	p.frame = p.context + 2
// 	p.Return(value)
// }

// func (p *Process) FastResumeHandler(value api.Object) {
// 	p.Return(value)
// }

// func (p *Process) ReifyHandlerContext(body api.Object) {
// 	fmt.Println(p.context)

// 	ctx := AsInt(p.data[p.context])
// 	data := make([]api.Object, ctx)
// 	copy(data, p.data[p.context-ctx+4:])

// 	fmt.Println(data)

// 	p.frame -= ctx
// 	p.data[p.frame-2] = Int(AsInt(p.data[p.frame]) - 2)
// 	p.data[p.frame-1] = p.data[p.frame+1]
// 	p.frame -= 2

// 	p.data[p.frame+2] = &contextObject{data: data}

// 	p.value = body
// 	p.callMethod(p.env.callMethodID)
// }

// func (p *Process) SlowResumeHandler(ctx, value api.Object) {

// }
