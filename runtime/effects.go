package runtime

import "github.com/bobappleyard/cezanne/format"

func (p *Process) installHandlers(handlers Object, next int) {
	// shift the activation frame up
	p.data[p.frame+2] = Int(AsInt(p.data[p.frame]) + 2)
	p.data[p.frame+3] = p.data[p.frame+1]

	// install the handler context
	p.data[p.frame] = Int(p.frame - next)
	p.data[p.frame+1] = handlers

	// set context registers
	p.context = p.frame
	p.frame += 2
}

func (p *Process) EnterContext(handlers Object, body Object) {
	p.installHandlers(handlers, p.context)

	// enter body
	p.value = body
	p.callMethod(p.env.callMethodID)
}

func (p *Process) TriggerEffect(id format.MethodID, argCount int) {
	for ctx := p.context; ctx != 0; ctx -= AsInt(p.data[ctx]) {
		handlers := p.data[ctx+1]
		method := p.getMethod(handlers, id)
		if method == nil {
			continue
		}

		// shift args + establish fake handler context
		copy(p.data[p.frame+4:], p.data[p.frame+2:p.frame+2+argCount])
		p.installHandlers(&standardObject{classID: p.env.emptyClassID}, ctx)

		// enter handler
		p.value = handlers
		p.callMethod(id)
	}
}

func (p *Process) FastAbortHandler(value Object) {

}

func (p *Process) FastResumeHandler(value Object) {

}

func (p *Process) ReifyHandlerContext(body Object) {

}

func (p *Process) SlowResumeHandler(ctx, value Object) {

}
