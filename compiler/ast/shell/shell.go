package shell

type File struct {
	Stmts []Stmt
}

type Expr interface {
	expr()
}

type IntLit struct {
	Value int
}

type StrLit struct {
	Value string
}

type FltLit struct {
	Value float64
}

type Builtins struct{}

type VarRef struct {
	Name string
}

type MemberRef struct {
	Object Expr
	Name   string
}

type Call struct {
	Fn   Expr
	Args []Expr
}

type Lambda struct {
	Args []string
	Body []Stmt
}

type Op struct {
	Name  string
	Left  Expr
	Right Expr
}

func (IntLit) expr()    {}
func (StrLit) expr()    {}
func (FltLit) expr()    {}
func (Builtins) expr()  {}
func (VarRef) expr()    {}
func (MemberRef) expr() {}
func (Call) expr()      {}
func (Lambda) expr()    {}
func (Op) expr()        {}

type Stmt interface {
	stmt()
}

type Def interface {
	def()
}

type VarDef struct {
	Name  string
	Value Expr
}

type FuncDef struct {
	Name string
	Args []string
	Body []Stmt
}

type ClassDef struct {
	Name    string
	Extend  Expr
	Members []MemberDef
}

type ExprStmt struct {
	Expr Expr
}

type Return struct {
	Value Expr
}

type If struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}

type Assign struct {
	Target Expr
	Value  Expr
}

type DecoratedDef struct {
	Value Expr
	Def   Def
}

func (VarDef) stmt()       {}
func (FuncDef) stmt()      {}
func (ClassDef) stmt()     {}
func (ExprStmt) stmt()     {}
func (Return) stmt()       {}
func (If) stmt()           {}
func (Assign) stmt()       {}
func (DecoratedDef) stmt() {}

func (FuncDef) def()  {}
func (ClassDef) def() {}
func (VarDef) def()   {}

type MemberDef interface {
	member()
}

type Field struct {
	Name string
	Init Expr
}

type Method struct {
	Name string
	Args []string
	Body []Stmt
}

type DecoratedMember struct {
	Value  Expr
	Member MemberDef
}

func (Field) member()           {}
func (Method) member()          {}
func (DecoratedMember) member() {}
