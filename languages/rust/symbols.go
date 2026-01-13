//go:build lang_rust || lang_all

package rust

import (
	"strings"

	"github.com/roveo/topo-mcp/languages"
)

// Function represents a Rust function or method
type Function struct {
	name       string
	signature  string
	receiver   string // For methods in impl blocks
	traitImpl  string // Trait being implemented (if any)
	visibility string
	doc        string
	loc        languages.Range
}

func (f *Function) Name() string { return f.name }
func (f *Function) Kind() string {
	if f.receiver != "" {
		return "method"
	}
	return "func"
}
func (f *Function) Location() languages.Range { return f.loc }
func (f *Function) String() string {
	var sb strings.Builder
	if f.visibility != "" {
		sb.WriteString(f.visibility)
		sb.WriteString(" ")
	}
	if f.receiver != "" {
		if f.traitImpl != "" {
			sb.WriteString("impl ")
			sb.WriteString(f.traitImpl)
			sb.WriteString(" for ")
		} else {
			sb.WriteString("impl ")
		}
		sb.WriteString(f.receiver)
		sb.WriteString(": ")
	}
	sb.WriteString("fn ")
	sb.WriteString(f.name)
	sb.WriteString(f.signature)
	return sb.String()
}
func (f *Function) DocComment() string { return f.doc }

// Struct represents a Rust struct
type Struct struct {
	name       string
	visibility string
	doc        string
	loc        languages.Range
}

func (s *Struct) Name() string              { return s.name }
func (s *Struct) Kind() string              { return "struct" }
func (s *Struct) Location() languages.Range { return s.loc }
func (s *Struct) String() string {
	var sb strings.Builder
	if s.visibility != "" {
		sb.WriteString(s.visibility)
		sb.WriteString(" ")
	}
	sb.WriteString("struct ")
	sb.WriteString(s.name)
	return sb.String()
}
func (s *Struct) DocComment() string { return s.doc }

// Enum represents a Rust enum
type Enum struct {
	name       string
	visibility string
	doc        string
	loc        languages.Range
}

func (e *Enum) Name() string              { return e.name }
func (e *Enum) Kind() string              { return "enum" }
func (e *Enum) Location() languages.Range { return e.loc }
func (e *Enum) String() string {
	var sb strings.Builder
	if e.visibility != "" {
		sb.WriteString(e.visibility)
		sb.WriteString(" ")
	}
	sb.WriteString("enum ")
	sb.WriteString(e.name)
	return sb.String()
}
func (e *Enum) DocComment() string { return e.doc }

// Trait represents a Rust trait
type Trait struct {
	name       string
	visibility string
	doc        string
	loc        languages.Range
}

func (t *Trait) Name() string              { return t.name }
func (t *Trait) Kind() string              { return "trait" }
func (t *Trait) Location() languages.Range { return t.loc }
func (t *Trait) String() string {
	var sb strings.Builder
	if t.visibility != "" {
		sb.WriteString(t.visibility)
		sb.WriteString(" ")
	}
	sb.WriteString("trait ")
	sb.WriteString(t.name)
	return sb.String()
}
func (t *Trait) DocComment() string { return t.doc }

// Const represents a Rust const item
type Const struct {
	name       string
	visibility string
	doc        string
	loc        languages.Range
}

func (c *Const) Name() string              { return c.name }
func (c *Const) Kind() string              { return "const" }
func (c *Const) Location() languages.Range { return c.loc }
func (c *Const) String() string {
	var sb strings.Builder
	if c.visibility != "" {
		sb.WriteString(c.visibility)
		sb.WriteString(" ")
	}
	sb.WriteString("const ")
	sb.WriteString(c.name)
	return sb.String()
}
func (c *Const) DocComment() string { return c.doc }

// Static represents a Rust static item
type Static struct {
	name       string
	visibility string
	doc        string
	loc        languages.Range
}

func (s *Static) Name() string              { return s.name }
func (s *Static) Kind() string              { return "static" }
func (s *Static) Location() languages.Range { return s.loc }
func (s *Static) String() string {
	var sb strings.Builder
	if s.visibility != "" {
		sb.WriteString(s.visibility)
		sb.WriteString(" ")
	}
	sb.WriteString("static ")
	sb.WriteString(s.name)
	return sb.String()
}
func (s *Static) DocComment() string { return s.doc }

// TypeAlias represents a Rust type alias
type TypeAlias struct {
	name       string
	visibility string
	doc        string
	loc        languages.Range
}

func (t *TypeAlias) Name() string              { return t.name }
func (t *TypeAlias) Kind() string              { return "type" }
func (t *TypeAlias) Location() languages.Range { return t.loc }
func (t *TypeAlias) String() string {
	var sb strings.Builder
	if t.visibility != "" {
		sb.WriteString(t.visibility)
		sb.WriteString(" ")
	}
	sb.WriteString("type ")
	sb.WriteString(t.name)
	return sb.String()
}
func (t *TypeAlias) DocComment() string { return t.doc }

// Mod represents a Rust module declaration
type Mod struct {
	name       string
	visibility string
	doc        string
	loc        languages.Range
}

func (m *Mod) Name() string              { return m.name }
func (m *Mod) Kind() string              { return "mod" }
func (m *Mod) Location() languages.Range { return m.loc }
func (m *Mod) String() string {
	var sb strings.Builder
	if m.visibility != "" {
		sb.WriteString(m.visibility)
		sb.WriteString(" ")
	}
	sb.WriteString("mod ")
	sb.WriteString(m.name)
	return sb.String()
}
func (m *Mod) DocComment() string { return m.doc }
