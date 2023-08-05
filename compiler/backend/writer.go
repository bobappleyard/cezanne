package backend

import (
	"bytes"
	"fmt"

	"github.com/bobappleyard/cezanne/format"
)

type assemblyWriter struct {
	prefix bytes.Buffer
	code   bytes.Buffer
	funcs  []string
	meta   format.Package
}

func (w *assemblyWriter) ImplementFunction(name string, argc, varc int, block func()) {
	fmt.Fprintf(&w.code, "extern void cz_impl_%s_%s() {\n", w.meta.Name, name)
	fmt.Fprintf(&w.code, "    CZ_PROLOG(%d, %d);\n", argc, varc)
	block()
	fmt.Fprintln(&w.code, "    CZ_EPILOG();")
	fmt.Fprintln(&w.code, "}")
	fmt.Fprintln(&w.code)
}

func (w *assemblyWriter) ImplementMethod(classID int, method string, argc, varc int, block func()) {
	w.addMethod(method)
	fmt.Fprintf(&w.code, "extern void cz_impl_%s_%d_%s() {\n", w.meta.Name, classID, method)
	fmt.Fprintf(&w.code, "    CZ_PROLOG(%d, %d);\n", argc, varc)
	block()
	fmt.Fprintln(&w.code, "    CZ_EPILOG();")
	fmt.Fprintln(&w.code, "}")
	fmt.Fprintln(&w.code)
}

func (w *assemblyWriter) Natural(val int) {
	w.call("CZ_INT", val)
}

func (w *assemblyWriter) Load(from variable) {
	w.call("CZ_LOAD", from)
}

func (w *assemblyWriter) Store(into variable) {
	w.call("CZ_STORE", into)
}

func (w *assemblyWriter) Field(f int) {
	w.call("CZ_FIELD", f)
}

func (w *assemblyWriter) GlobalLoad(from int) {
	w.call("CZ_GLOBAL", from)
}

func (w *assemblyWriter) Create(classID int, base variable) {
	w.call("CZ_CREATE", fmt.Sprintf("cz_classes_%s + %d", w.meta.Name, classID), base)
}

func (w *assemblyWriter) Return() {
	w.call("CZ_RETURN")
}

func (w *assemblyWriter) FunctionCall(name string, base variable) {
	w.addFunction(name)
	w.call("CZ_CALL", name, base)
}

func (w *assemblyWriter) FunctionCallTail(name string) {
	w.addFunction(name)
	w.call("CZ_CALL_TAIL", name)
}

func (w *assemblyWriter) call(name string, args ...any) {
	fmt.Fprintf(&w.code, "    %s(", name)
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(&w.code, ", ")
		}
		fmt.Fprint(&w.code, arg)
	}
	fmt.Fprintln(&w.code, ");")
}

func (w *assemblyWriter) addMethod(name string) {
	for _, n := range w.meta.Methods {
		if n.Name == name {
			return
		}
	}
	w.meta.Methods = append(w.meta.Methods, format.Method{Name: name})
	fmt.Fprintf(&w.prefix, "extern void cz_m_%s();\n", name)
}

func (w *assemblyWriter) addFunction(name string) {
	for _, n := range w.funcs {
		if n == name {
			return
		}
	}
	w.funcs = append(w.funcs, name)
	fmt.Fprintf(&w.prefix, "extern void %s();\n", name)
}
