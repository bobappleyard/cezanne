package backend

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/compiler/ast"
)

func TestBuildPackage(t *testing.T) {

	body, meta, err := BuildPackage(ast.Package{
		Name: "main",
		Imports: []ast.Import{
			{Name: "test", Path: "test"},
		},
		Funcs: []ast.Method{
			{
				Name: "main",
				Body: ast.Invoke{
					Object: ast.Ref{Name: "test"},
					Name:   "print",
					Args: []ast.Expr{
						ast.Invoke{
							Object: ast.Ref{Name: "fac"},
							Name:   "call",
							Args:   []ast.Expr{ast.Int{Value: 4}},
						},
					},
				},
			},
			{
				Name: "fac",
				Args: []string{"x"},
				Body: ast.Invoke{
					Object: ast.Invoke{
						Object: ast.Ref{Name: "test"},
						Name:   "lte",
						Args: []ast.Expr{
							ast.Ref{Name: "x"},
							ast.Int{Value: 1},
						},
					},
					Name: "match",
					Args: []ast.Expr{ast.Create{Methods: []ast.Method{
						{
							Name: "true",
							Body: ast.Ref{Name: "x"},
						},
						{
							Name: "false",
							Body: ast.Invoke{
								Object: ast.Ref{Name: "test"},
								Name:   "mul",
								Args: []ast.Expr{
									ast.Ref{Name: "x"},
									ast.Invoke{
										Object: ast.Ref{Name: "fac"},
										Name:   "call",
										Args: []ast.Expr{
											ast.Invoke{
												Object: ast.Ref{Name: "test"},
												Name:   "sub",
												Args: []ast.Expr{
													ast.Ref{Name: "x"},
													ast.Int{Value: 1},
												},
											},
										},
									},
								},
							},
						},
					}}},
				},
			},
		},
	})
	assert.Nil(t, err)

	fmt.Printf("%s", body)

	assert.Equal(t, body, []byte{})
	assert.Equal(t, meta, nil)

}
