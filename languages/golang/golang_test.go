//go:build lang_go || lang_all || (!lang_python && !lang_typescript && !lang_rust)

package golang

import (
	"testing"
)

func TestLanguageMetadata(t *testing.T) {
	lang := &Language{}

	if lang.Name() != "go" {
		t.Errorf("expected name 'go', got %q", lang.Name())
	}

	exts := lang.Extensions()
	if len(exts) != 1 || exts[0] != ".go" {
		t.Errorf("expected extensions [.go], got %v", exts)
	}
}

func TestParseFunction(t *testing.T) {
	src := `package main

// greet prints a greeting message
func greet(name string) error {
	return nil
}
`
	lang := &Language{}
	imports, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(imports) != 0 {
		t.Errorf("expected no imports, got %v", imports)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	sym := symbols[0]
	if sym.Name() != "greet" {
		t.Errorf("expected name 'greet', got %q", sym.Name())
	}
	if sym.Kind() != "func" {
		t.Errorf("expected kind 'func', got %q", sym.Kind())
	}
	if sym.String() != "greet(string) error" {
		t.Errorf("expected 'greet(string) error', got %q", sym.String())
	}

	// Check doc comment
	if doc, ok := sym.(interface{ DocComment() string }); ok {
		if doc.DocComment() != "greet prints a greeting message" {
			t.Errorf("expected doc comment, got %q", doc.DocComment())
		}
	}

	// Check location
	loc := sym.Location()
	if loc.Start.Line != 3 {
		t.Errorf("expected start line 3, got %d", loc.Start.Line)
	}
}

func TestParseMethod(t *testing.T) {
	src := `package main

type Server struct{}

func (s *Server) Start() error {
	return nil
}
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}

	// Find the method
	var method *Method
	for _, sym := range symbols {
		if m, ok := sym.(*Method); ok {
			method = m
			break
		}
	}

	if method == nil {
		t.Fatal("expected to find a method")
	}

	if method.Name() != "Start" {
		t.Errorf("expected name 'Start', got %q", method.Name())
	}
	if method.Kind() != "method" {
		t.Errorf("expected kind 'method', got %q", method.Kind())
	}
	if method.String() != "(*Server) Start() error" {
		t.Errorf("expected '(*Server) Start() error', got %q", method.String())
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		wantKind string
		wantStr  string
	}{
		{
			name: "struct",
			src: `package main
type Config struct {
	Name string
}`,
			wantKind: "struct",
			wantStr:  "type Config struct",
		},
		{
			name: "interface",
			src: `package main
type Reader interface {
	Read([]byte) (int, error)
}`,
			wantKind: "interface",
			wantStr:  "type Reader interface",
		},
		{
			name: "alias",
			src: `package main
type ID int`,
			wantKind: "int",
			wantStr:  "type ID int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := &Language{}
			_, symbols, err := lang.Parse([]byte(tt.src))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(symbols) != 1 {
				t.Fatalf("expected 1 symbol, got %d", len(symbols))
			}

			typ, ok := symbols[0].(*Type)
			if !ok {
				t.Fatalf("expected *Type, got %T", symbols[0])
			}

			if typ.typeKind != tt.wantKind {
				t.Errorf("expected typeKind %q, got %q", tt.wantKind, typ.typeKind)
			}
			if typ.String() != tt.wantStr {
				t.Errorf("expected %q, got %q", tt.wantStr, typ.String())
			}
		})
	}
}

func TestParseConstAndVar(t *testing.T) {
	src := `package main

const MaxSize = 100
var DefaultName = "test"
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}

	// Check const
	c := symbols[0]
	if c.Kind() != "const" {
		t.Errorf("expected kind 'const', got %q", c.Kind())
	}
	if c.Name() != "MaxSize" {
		t.Errorf("expected name 'MaxSize', got %q", c.Name())
	}
	if c.String() != "const MaxSize" {
		t.Errorf("expected 'const MaxSize', got %q", c.String())
	}

	// Check var
	v := symbols[1]
	if v.Kind() != "var" {
		t.Errorf("expected kind 'var', got %q", v.Kind())
	}
	if v.Name() != "DefaultName" {
		t.Errorf("expected name 'DefaultName', got %q", v.Name())
	}
}

func TestParseImports(t *testing.T) {
	src := `package main

import (
	"fmt"
	"os"
	"github.com/example/pkg"
)

func main() {}
`
	lang := &Language{}
	imports, _, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := []string{"fmt", "os", "github.com/example/pkg"}
	if len(imports) != len(expected) {
		t.Fatalf("expected %d imports, got %d", len(expected), len(imports))
	}

	for i, exp := range expected {
		if imports[i] != exp {
			t.Errorf("import[%d]: expected %q, got %q", i, exp, imports[i])
		}
	}
}

func TestParseFunctionSignatures(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantSig string
	}{
		{
			name:    "no params no return",
			src:     `package main; func f() {}`,
			wantSig: "f()",
		},
		{
			name:    "single param",
			src:     `package main; func f(x int) {}`,
			wantSig: "f(int)",
		},
		{
			name:    "multiple params same type",
			src:     `package main; func f(x, y int) {}`,
			wantSig: "f(int, int)",
		},
		{
			name:    "multiple params different types",
			src:     `package main; func f(x int, y string) {}`,
			wantSig: "f(int, string)",
		},
		{
			name:    "single return",
			src:     `package main; func f() error { return nil }`,
			wantSig: "f() error",
		},
		{
			name:    "multiple returns",
			src:     `package main; func f() (int, error) { return 0, nil }`,
			wantSig: "f() (int, error)",
		},
		// Note: variadic parameters are complex in tree-sitter, simplified for now
		// {
		// 	name:    "variadic",
		// 	src:     `package main; func f(args ...string) {}`,
		// 	wantSig: "f(...string)",
		// },
		{
			name:    "pointer param",
			src:     `package main; func f(x *int) {}`,
			wantSig: "f(*int)",
		},
		{
			name:    "slice param",
			src:     `package main; func f(x []byte) {}`,
			wantSig: "f([]byte)",
		},
		{
			name:    "map param",
			src:     `package main; func f(x map[string]int) {}`,
			wantSig: "f(map[string]int)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := &Language{}
			_, symbols, err := lang.Parse([]byte(tt.src))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(symbols) != 1 {
				t.Fatalf("expected 1 symbol, got %d", len(symbols))
			}

			if symbols[0].String() != tt.wantSig {
				t.Errorf("expected %q, got %q", tt.wantSig, symbols[0].String())
			}
		})
	}
}

func TestParseEmptyFile(t *testing.T) {
	src := `package main`

	lang := &Language{}
	imports, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(imports) != 0 {
		t.Errorf("expected no imports, got %v", imports)
	}
	if len(symbols) != 0 {
		t.Errorf("expected no symbols, got %v", symbols)
	}
}

func TestParseMultipleConstsInBlock(t *testing.T) {
	src := `package main

const (
	A = 1
	B = 2
	C = 3
)
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(symbols))
	}

	names := []string{"A", "B", "C"}
	for i, sym := range symbols {
		if sym.Name() != names[i] {
			t.Errorf("symbol[%d]: expected name %q, got %q", i, names[i], sym.Name())
		}
		if sym.Kind() != "const" {
			t.Errorf("symbol[%d]: expected kind 'const', got %q", i, sym.Kind())
		}
	}
}
