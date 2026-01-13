package typescript

import (
	"strings"

	"github.com/roveo/topo-mcp/languages"
)

// Function represents a JS/TS function declaration
type Function struct {
	name      string
	signature string
	isAsync   bool
	doc       string
	loc       languages.Range
}

func (f *Function) Name() string              { return f.name }
func (f *Function) Kind() string              { return "func" }
func (f *Function) Location() languages.Range { return f.loc }
func (f *Function) String() string {
	var sb strings.Builder
	if f.isAsync {
		sb.WriteString("async ")
	}
	sb.WriteString("function ")
	sb.WriteString(f.name)
	sb.WriteString(f.signature)
	return sb.String()
}
func (f *Function) DocComment() string { return f.doc }

// Class represents a JS/TS class declaration
type Class struct {
	name       string
	extends    string
	implements []string
	doc        string
	loc        languages.Range
}

func (c *Class) Name() string              { return c.name }
func (c *Class) Kind() string              { return "class" }
func (c *Class) Location() languages.Range { return c.loc }
func (c *Class) String() string {
	var sb strings.Builder
	sb.WriteString("class ")
	sb.WriteString(c.name)
	if c.extends != "" {
		sb.WriteString(" extends ")
		sb.WriteString(c.extends)
	}
	if len(c.implements) > 0 {
		sb.WriteString(" implements ")
		sb.WriteString(strings.Join(c.implements, ", "))
	}
	return sb.String()
}
func (c *Class) DocComment() string { return c.doc }

// Interface represents a TypeScript interface declaration
type Interface struct {
	name string
	doc  string
	loc  languages.Range
}

func (i *Interface) Name() string              { return i.name }
func (i *Interface) Kind() string              { return "interface" }
func (i *Interface) Location() languages.Range { return i.loc }
func (i *Interface) String() string            { return "interface " + i.name }
func (i *Interface) DocComment() string        { return i.doc }

// TypeAlias represents a TypeScript type alias declaration
type TypeAlias struct {
	name string
	doc  string
	loc  languages.Range
}

func (t *TypeAlias) Name() string              { return t.name }
func (t *TypeAlias) Kind() string              { return "type" }
func (t *TypeAlias) Location() languages.Range { return t.loc }
func (t *TypeAlias) String() string            { return "type " + t.name }
func (t *TypeAlias) DocComment() string        { return t.doc }

// Enum represents a TypeScript enum declaration
type Enum struct {
	name string
	doc  string
	loc  languages.Range
}

func (e *Enum) Name() string              { return e.name }
func (e *Enum) Kind() string              { return "enum" }
func (e *Enum) Location() languages.Range { return e.loc }
func (e *Enum) String() string            { return "enum " + e.name }
func (e *Enum) DocComment() string        { return e.doc }

// Variable represents a JS/TS variable declaration
type Variable struct {
	name string
	kind string // "const", "let", "var"
	loc  languages.Range
}

func (v *Variable) Name() string              { return v.name }
func (v *Variable) Kind() string              { return v.kind }
func (v *Variable) Location() languages.Range { return v.loc }
func (v *Variable) String() string            { return v.kind + " " + v.name }
