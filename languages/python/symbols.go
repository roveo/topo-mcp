//go:build lang_python || lang_all

package python

import (
	"strings"

	"github.com/roveo/topo-mcp/languages"
)

// Function represents a Python function definition
type Function struct {
	name       string
	signature  string
	decorators []string
	doc        string
	loc        languages.Range
}

func (f *Function) Name() string              { return f.name }
func (f *Function) Kind() string              { return "func" }
func (f *Function) Location() languages.Range { return f.loc }
func (f *Function) String() string {
	var sb strings.Builder
	for _, dec := range f.decorators {
		sb.WriteString("@")
		sb.WriteString(dec)
		sb.WriteString(" ")
	}
	sb.WriteString("def ")
	sb.WriteString(f.name)
	sb.WriteString(f.signature)
	return sb.String()
}
func (f *Function) DocComment() string { return f.doc }

// Class represents a Python class definition
type Class struct {
	name       string
	bases      []string
	decorators []string
	doc        string
	loc        languages.Range
}

func (c *Class) Name() string              { return c.name }
func (c *Class) Kind() string              { return "class" }
func (c *Class) Location() languages.Range { return c.loc }
func (c *Class) String() string {
	var sb strings.Builder
	for _, dec := range c.decorators {
		sb.WriteString("@")
		sb.WriteString(dec)
		sb.WriteString(" ")
	}
	sb.WriteString("class ")
	sb.WriteString(c.name)
	if len(c.bases) > 0 {
		sb.WriteString("(")
		sb.WriteString(strings.Join(c.bases, ", "))
		sb.WriteString(")")
	}
	return sb.String()
}
func (c *Class) DocComment() string { return c.doc }

// Variable represents a Python module-level variable
type Variable struct {
	name string
	loc  languages.Range
}

func (v *Variable) Name() string              { return v.name }
func (v *Variable) Kind() string              { return "var" }
func (v *Variable) Location() languages.Range { return v.loc }
func (v *Variable) String() string            { return v.name }
