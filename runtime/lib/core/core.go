package core

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime"
)

func init() {
	runtime.RegisterExt("core:communicate_linkage", communicateLinkage)
	runtime.RegisterExt("core:register_package", registerPackage)
	runtime.RegisterExt("core:enter_context", enterContext)
	runtime.RegisterExt("core:trigger_effect", triggerEffect)
	runtime.RegisterExt("core:fast_abort_handler", fastAbortHandler)
	runtime.RegisterExt("core:fast_resume_handler", fastResumeHandler)
	runtime.RegisterExt("core:reify_handler_context", reifyHandlerContext)
	runtime.RegisterExt("core:slow_resume_handler", slowResumeHandler)
}

func asMethodID(x runtime.Object) format.MethodID {
	return format.MethodID(runtime.AsInt(x))
}

func asClassID(x runtime.Object) format.ClassID {
	return format.ClassID(runtime.AsInt(x))
}

func communicateLinkage(p *runtime.Process) {
	p.Env().CommunicateLinkage(asMethodID(p.Arg(0)))
}

func registerPackage(p *runtime.Process) {
	p.Env().RegisterPackage(p.Arg(0))
}

func enterContext(p *runtime.Process) {
	p.EnterContext(p.Arg(0), p.Arg(1))
}

func triggerEffect(p *runtime.Process) {
	p.TriggerEffect(asMethodID(p.Arg(0)), p.Arg(1))
}

func fastAbortHandler(p *runtime.Process) {
	p.FastAbortHandler(p.Arg(0))
}

func fastResumeHandler(p *runtime.Process) {
	p.FastResumeHandler(p.Arg(0))
}

func reifyHandlerContext(p *runtime.Process) {
	p.ReifyHandlerContext(p.Arg(0))
}

func slowResumeHandler(p *runtime.Process) {
	p.SlowResumeHandler(p.Arg(0), p.Arg(1))
}
