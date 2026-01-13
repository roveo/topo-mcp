//go:build lang_go || lang_all || (!lang_python && !lang_typescript && !lang_rust)

package golang

import (
	"fmt"

	"github.com/roveo/topo-mcp/languages"
)

// Function represents a Go function declaration
type Function struct {
	name      string
	signature string
	doc       string
	loc       languages.Range
}

func (f *Function) Name() string              { return f.name }
func (f *Function) Kind() string              { return "func" }
func (f *Function) Location() languages.Range { return f.loc }
func (f *Function) String() string {
	return fmt.Sprintf("%s%s", f.name, f.signature)
}
func (f *Function) DocComment() string { return f.doc }

// Method represents a Go method declaration
type Method struct {
	name      string
	receiver  string
	signature string
	doc       string
	loc       languages.Range
}

func (m *Method) Name() string              { return m.name }
func (m *Method) Kind() string              { return "method" }
func (m *Method) Location() languages.Range { return m.loc }
func (m *Method) String() string {
	return fmt.Sprintf("(%s) %s%s", m.receiver, m.name, m.signature)
}
func (m *Method) DocComment() string { return m.doc }

// Type represents a Go type declaration
type Type struct {
	name     string
	typeKind string
	doc      string
	loc      languages.Range
}

func (t *Type) Name() string              { return t.name }
func (t *Type) Kind() string              { return "type" }
func (t *Type) Location() languages.Range { return t.loc }
func (t *Type) String() string {
	return fmt.Sprintf("type %s %s", t.name, t.typeKind)
}
func (t *Type) DocComment() string { return t.doc }

// Const represents a Go const declaration
type Const struct {
	name string
	doc  string
	loc  languages.Range
}

func (c *Const) Name() string              { return c.name }
func (c *Const) Kind() string              { return "const" }
func (c *Const) Location() languages.Range { return c.loc }
func (c *Const) String() string {
	return fmt.Sprintf("const %s", c.name)
}
func (c *Const) DocComment() string { return c.doc }

// Var represents a Go var declaration
type Var struct {
	name string
	doc  string
	loc  languages.Range
}

func (v *Var) Name() string              { return v.name }
func (v *Var) Kind() string              { return "var" }
func (v *Var) Location() languages.Range { return v.loc }
func (v *Var) String() string {
	return fmt.Sprintf("var %s", v.name)
}
func (v *Var) DocComment() string { return v.doc }
