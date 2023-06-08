package compile

import (
	"os"

	"github.com/bobappleyard/cezanne/commands"
	"github.com/bobappleyard/cezanne/commands/compile/ast"
	"github.com/bobappleyard/cezanne/commands/compile/backend"
	"github.com/bobappleyard/cezanne/commands/compile/parser"
	"github.com/bobappleyard/cezanne/format/storage"
	"github.com/bobappleyard/cezanne/format/symtab"
)

type Options struct {
	Output string `option:"o"`
}

func init() {
	commands.Register("compile", Compile)
}

func Compile(options Options, files []string) error {
	var sourceModel ast.Package
	var syms symtab.Symtab

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		err = parser.ParseFile(&syms, &sourceModel, data)
		if err != nil {
			return err
		}
	}

	objectModel, err := backend.BuildPackage(&syms, sourceModel)
	if err != nil {
		return err
	}

	output, err := os.Create(options.Output)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = storage.Write(output, objectModel)
	return err
}
