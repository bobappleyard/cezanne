package linker

import (
	"io"
	"text/template"

	"github.com/bobappleyard/cezanne/must"
)

func Render(w io.Writer, p *Program) error {
	return tmpl.Execute(w, p)
}

var tmpl = must.Be(template.New("").Parse(`

#include <stddef.h>
#include <cz.h>

{{range .Included -}}
	const int cz_classes_{{.Name}} = {{.Offset}};
{{end}}

const cz_class_t cz_classes[{{len .Classes}}] = {
{{- range $i, $item := .Classes}}
	{{- if $i}},{{end}}
	{ .id = {{.ID}}, .fieldc = {{.FieldCount}} }
{{- end}}
};
{{range .Implementations}}
{{if .Symbol}}extern void {{.Symbol}}();{{end}}
{{- end}}

const cz_impl_t cz_impls[{{len .Implementations}}] = {
{{- range $i, $item := .Implementations}}
	{{- if $i}},{{end}}
	{{- if .Symbol}}
		{ .method_id = {{.Method}}, .impl = {{.Symbol}} }
	{{- else}}
		{ .method_id = {{.Method}}, .impl = NULL }
	{{- end}}
{{- end}}
};
{{range .Methods}}
extern void cz_m_{{.Name}}() {
	CZ_METHOD_LOOKUP({{.Offset}});
}
{{end}}

{{- range .Included}}
extern void cz_impl_{{.Name}}();
{{- end}}

extern void cz_init() {
	CZ_PROLOG(0, 0);
{{- range .Included}}
	CZ_CALL(cz_impl_{{.Name}}, 0);
{{- end}}
	CZ_RETURN();
	CZ_EPILOG();
}
`))