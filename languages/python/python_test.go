package python

import (
	"strings"
	"testing"
)

func TestLanguageMetadata(t *testing.T) {
	lang := &Language{}

	if lang.Name() != "python" {
		t.Errorf("expected name 'python', got %q", lang.Name())
	}

	exts := lang.Extensions()
	if len(exts) != 1 || exts[0] != ".py" {
		t.Errorf("expected extensions [.py], got %v", exts)
	}
}

func TestParseFunction(t *testing.T) {
	src := `def greet(name: str) -> str:
    """Say hello to someone."""
    return f"Hello, {name}"
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

	str := sym.String()
	if !strings.Contains(str, "def greet") {
		t.Errorf("expected String() to contain 'def greet', got %q", str)
	}
	if !strings.Contains(str, "-> str") {
		t.Errorf("expected String() to contain '-> str', got %q", str)
	}

	// Check doc comment
	if doc, ok := sym.(interface{ DocComment() string }); ok {
		if doc.DocComment() != "Say hello to someone." {
			t.Errorf("expected doc comment 'Say hello to someone.', got %q", doc.DocComment())
		}
	}
}

func TestParseClass(t *testing.T) {
	src := `class Server(BaseServer):
    """HTTP server implementation."""
    
    def __init__(self, port: int):
        self.port = port
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	cls, ok := symbols[0].(*Class)
	if !ok {
		t.Fatalf("expected *Class, got %T", symbols[0])
	}

	if cls.Name() != "Server" {
		t.Errorf("expected name 'Server', got %q", cls.Name())
	}
	if cls.Kind() != "class" {
		t.Errorf("expected kind 'class', got %q", cls.Kind())
	}

	str := cls.String()
	if !strings.Contains(str, "class Server") {
		t.Errorf("expected String() to contain 'class Server', got %q", str)
	}
	if !strings.Contains(str, "BaseServer") {
		t.Errorf("expected String() to contain 'BaseServer', got %q", str)
	}

	if cls.DocComment() != "HTTP server implementation." {
		t.Errorf("expected doc comment, got %q", cls.DocComment())
	}
}

func TestParseDecoratedFunction(t *testing.T) {
	src := `@decorator
@another(arg=1)
def decorated_func():
    pass
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	fn, ok := symbols[0].(*Function)
	if !ok {
		t.Fatalf("expected *Function, got %T", symbols[0])
	}

	if fn.Name() != "decorated_func" {
		t.Errorf("expected name 'decorated_func', got %q", fn.Name())
	}

	str := fn.String()
	if !strings.Contains(str, "@decorator") {
		t.Errorf("expected String() to contain '@decorator', got %q", str)
	}
	if !strings.Contains(str, "@another") {
		t.Errorf("expected String() to contain '@another', got %q", str)
	}
}

func TestParseDecoratedClass(t *testing.T) {
	src := `@dataclass
class Config:
    name: str
    value: int
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	cls, ok := symbols[0].(*Class)
	if !ok {
		t.Fatalf("expected *Class, got %T", symbols[0])
	}

	str := cls.String()
	if !strings.Contains(str, "@dataclass") {
		t.Errorf("expected String() to contain '@dataclass', got %q", str)
	}
	if !strings.Contains(str, "class Config") {
		t.Errorf("expected String() to contain 'class Config', got %q", str)
	}
}

func TestParseImports(t *testing.T) {
	src := `import os
import sys
from typing import List, Optional
from dataclasses import dataclass

def main():
    pass
`
	lang := &Language{}
	imports, _, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// We expect: os, sys, typing, dataclasses
	expected := []string{"os", "sys", "typing", "dataclasses"}
	if len(imports) != len(expected) {
		t.Fatalf("expected %d imports, got %d: %v", len(expected), len(imports), imports)
	}

	for i, exp := range expected {
		if imports[i] != exp {
			t.Errorf("import[%d]: expected %q, got %q", i, exp, imports[i])
		}
	}
}

func TestParseModuleLevelVariable(t *testing.T) {
	src := `VERSION = "1.0.0"
_private = "hidden"

def main():
    pass
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have VERSION (public) and main(), but not _private
	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}

	// Find the variable
	var foundVar *Variable
	for _, sym := range symbols {
		if v, ok := sym.(*Variable); ok {
			foundVar = v
			break
		}
	}

	if foundVar == nil {
		t.Fatal("expected to find a Variable")
	}

	if foundVar.Name() != "VERSION" {
		t.Errorf("expected name 'VERSION', got %q", foundVar.Name())
	}
	if foundVar.Kind() != "var" {
		t.Errorf("expected kind 'var', got %q", foundVar.Kind())
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
			src:     `def f(): pass`,
			wantSig: "def f()",
		},
		{
			name:    "with params",
			src:     `def f(x, y): pass`,
			wantSig: "def f(x, y)",
		},
		{
			name:    "with type hints",
			src:     `def f(x: int, y: str) -> bool: pass`,
			wantSig: "def f(x: int, y: str) -> bool",
		},
		{
			name:    "with default",
			src:     `def f(x: int = 0): pass`,
			wantSig: "def f(x: int = 0)",
		},
		{
			name:    "with args kwargs",
			src:     `def f(*args, **kwargs): pass`,
			wantSig: "def f(*args, **kwargs)",
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

func TestParseAsyncFunction(t *testing.T) {
	src := `async def fetch(url: str) -> bytes:
    """Fetch URL content."""
    pass
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	// async functions are still parsed as function_definition
	if symbols[0].Kind() != "func" {
		t.Errorf("expected kind 'func', got %q", symbols[0].Kind())
	}
}

func TestParseEmptyFile(t *testing.T) {
	src := `# Just a comment`

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

func TestParseClassWithMultipleBases(t *testing.T) {
	src := `class MyClass(Base1, Base2, metaclass=ABCMeta):
    pass
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	cls, ok := symbols[0].(*Class)
	if !ok {
		t.Fatalf("expected *Class, got %T", symbols[0])
	}

	// Should include Base1, Base2 but not metaclass=...
	str := cls.String()
	if !strings.Contains(str, "Base1") {
		t.Errorf("expected String() to contain 'Base1', got %q", str)
	}
	if !strings.Contains(str, "Base2") {
		t.Errorf("expected String() to contain 'Base2', got %q", str)
	}
}
