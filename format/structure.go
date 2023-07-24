package format

type Package struct {
	Name    string
	Imports []string
	Methods []Method
	Classes []Class
}

type Import struct {
	Path string
}

type Class struct {
	FieldCount int
	Methods    []Method
}

type Method struct {
	Name   string
	Offset int
}
