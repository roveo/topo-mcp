package rust

import (
	"strings"
	"testing"
)

func TestLanguageMetadata(t *testing.T) {
	lang := &Language{}

	if lang.Name() != "rust" {
		t.Errorf("expected name 'rust', got %q", lang.Name())
	}

	exts := lang.Extensions()
	if len(exts) != 1 || exts[0] != ".rs" {
		t.Errorf("expected extensions [.rs], got %v", exts)
	}
}

func TestParseFunction(t *testing.T) {
	src := `/// Greet someone by name
pub fn greet(name: &str) -> String {
    format!("Hello, {}", name)
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

	str := sym.String()
	if !strings.Contains(str, "pub") {
		t.Errorf("expected String() to contain 'pub', got %q", str)
	}
	if !strings.Contains(str, "fn greet") {
		t.Errorf("expected String() to contain 'fn greet', got %q", str)
	}
	if !strings.Contains(str, "-> String") {
		t.Errorf("expected String() to contain '-> String', got %q", str)
	}

	// Check doc comment
	if doc, ok := sym.(interface{ DocComment() string }); ok {
		if doc.DocComment() != "Greet someone by name" {
			t.Errorf("expected doc comment 'Greet someone by name', got %q", doc.DocComment())
		}
	}
}

func TestParseStruct(t *testing.T) {
	src := `/// Configuration for the server
pub struct Config {
    host: String,
    port: u16,
}
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	s, ok := symbols[0].(*Struct)
	if !ok {
		t.Fatalf("expected *Struct, got %T", symbols[0])
	}

	if s.Name() != "Config" {
		t.Errorf("expected name 'Config', got %q", s.Name())
	}
	if s.Kind() != "struct" {
		t.Errorf("expected kind 'struct', got %q", s.Kind())
	}

	str := s.String()
	if !strings.Contains(str, "pub struct Config") {
		t.Errorf("expected 'pub struct Config', got %q", str)
	}

	if s.DocComment() != "Configuration for the server" {
		t.Errorf("expected doc comment, got %q", s.DocComment())
	}
}

func TestParseEnum(t *testing.T) {
	src := `/// Possible errors
pub enum Error {
    NotFound,
    PermissionDenied,
    Unknown(String),
}
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	e, ok := symbols[0].(*Enum)
	if !ok {
		t.Fatalf("expected *Enum, got %T", symbols[0])
	}

	if e.Name() != "Error" {
		t.Errorf("expected name 'Error', got %q", e.Name())
	}
	if e.Kind() != "enum" {
		t.Errorf("expected kind 'enum', got %q", e.Kind())
	}
	if !strings.Contains(e.String(), "pub enum Error") {
		t.Errorf("expected 'pub enum Error', got %q", e.String())
	}
}

func TestParseTrait(t *testing.T) {
	src := `/// A handler for requests
pub trait Handler {
    fn handle(&self, req: Request) -> Response;
}
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	tr, ok := symbols[0].(*Trait)
	if !ok {
		t.Fatalf("expected *Trait, got %T", symbols[0])
	}

	if tr.Name() != "Handler" {
		t.Errorf("expected name 'Handler', got %q", tr.Name())
	}
	if tr.Kind() != "trait" {
		t.Errorf("expected kind 'trait', got %q", tr.Kind())
	}
	if !strings.Contains(tr.String(), "pub trait Handler") {
		t.Errorf("expected 'pub trait Handler', got %q", tr.String())
	}
}

func TestParseImpl(t *testing.T) {
	src := `struct Server {
    port: u16,
}

impl Server {
    pub fn new(port: u16) -> Self {
        Server { port }
    }
    
    pub fn start(&self) -> Result<(), Error> {
        Ok(())
    }
}
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have: Server (struct), new (method), start (method)
	if len(symbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(symbols))
	}

	// Find methods
	var methods []*Function
	for _, sym := range symbols {
		if fn, ok := sym.(*Function); ok && fn.receiver != "" {
			methods = append(methods, fn)
		}
	}

	if len(methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(methods))
	}

	// Check first method
	m := methods[0]
	if m.Name() != "new" {
		t.Errorf("expected name 'new', got %q", m.Name())
	}
	if m.Kind() != "method" {
		t.Errorf("expected kind 'method', got %q", m.Kind())
	}
	if m.receiver != "Server" {
		t.Errorf("expected receiver 'Server', got %q", m.receiver)
	}

	str := m.String()
	if !strings.Contains(str, "impl Server") {
		t.Errorf("expected String() to contain 'impl Server', got %q", str)
	}
}

func TestParseTraitImpl(t *testing.T) {
	src := `struct MyHandler;

impl Handler for MyHandler {
    fn handle(&self, req: Request) -> Response {
        Response::ok()
    }
}
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have: MyHandler (struct), handle (method)
	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}

	// Find the method
	var method *Function
	for _, sym := range symbols {
		if fn, ok := sym.(*Function); ok && fn.receiver != "" {
			method = fn
			break
		}
	}

	if method == nil {
		t.Fatal("expected to find a method")
	}

	if method.traitImpl != "Handler" {
		t.Errorf("expected traitImpl 'Handler', got %q", method.traitImpl)
	}

	str := method.String()
	if !strings.Contains(str, "impl Handler for MyHandler") {
		t.Errorf("expected String() to contain 'impl Handler for MyHandler', got %q", str)
	}
}

func TestParseConst(t *testing.T) {
	src := `/// Maximum connections allowed
pub const MAX_CONNECTIONS: usize = 100;
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	c, ok := symbols[0].(*Const)
	if !ok {
		t.Fatalf("expected *Const, got %T", symbols[0])
	}

	if c.Name() != "MAX_CONNECTIONS" {
		t.Errorf("expected name 'MAX_CONNECTIONS', got %q", c.Name())
	}
	if c.Kind() != "const" {
		t.Errorf("expected kind 'const', got %q", c.Kind())
	}
	if !strings.Contains(c.String(), "pub const MAX_CONNECTIONS") {
		t.Errorf("expected 'pub const MAX_CONNECTIONS', got %q", c.String())
	}
}

func TestParseStatic(t *testing.T) {
	src := `static mut COUNTER: u32 = 0;
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	s, ok := symbols[0].(*Static)
	if !ok {
		t.Fatalf("expected *Static, got %T", symbols[0])
	}

	if s.Name() != "COUNTER" {
		t.Errorf("expected name 'COUNTER', got %q", s.Name())
	}
	if s.Kind() != "static" {
		t.Errorf("expected kind 'static', got %q", s.Kind())
	}
}

func TestParseTypeAlias(t *testing.T) {
	src := `pub type Result<T> = std::result::Result<T, Error>;
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	typ, ok := symbols[0].(*TypeAlias)
	if !ok {
		t.Fatalf("expected *TypeAlias, got %T", symbols[0])
	}

	if typ.Name() != "Result" {
		t.Errorf("expected name 'Result', got %q", typ.Name())
	}
	if typ.Kind() != "type" {
		t.Errorf("expected kind 'type', got %q", typ.Kind())
	}
}

func TestParseMod(t *testing.T) {
	src := `pub mod server;
mod internal;
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}

	// First mod (pub)
	m1, ok := symbols[0].(*Mod)
	if !ok {
		t.Fatalf("expected *Mod, got %T", symbols[0])
	}
	if m1.Name() != "server" {
		t.Errorf("expected name 'server', got %q", m1.Name())
	}
	if !strings.Contains(m1.String(), "pub mod server") {
		t.Errorf("expected 'pub mod server', got %q", m1.String())
	}

	// Second mod (private)
	m2, ok := symbols[1].(*Mod)
	if !ok {
		t.Fatalf("expected *Mod, got %T", symbols[1])
	}
	if m2.Name() != "internal" {
		t.Errorf("expected name 'internal', got %q", m2.Name())
	}
	if m2.String() != "mod internal" {
		t.Errorf("expected 'mod internal', got %q", m2.String())
	}
}

func TestParseUse(t *testing.T) {
	src := `use std::collections::HashMap;
use std::io::Read;

fn main() {}
`
	lang := &Language{}
	imports, _, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(imports) != 2 {
		t.Fatalf("expected 2 imports, got %d: %v", len(imports), imports)
	}
}

func TestParseEmptyFile(t *testing.T) {
	src := `// Just a comment`

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

func TestParseFunctionSignatures(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantSig string
	}{
		{
			name:    "no params no return",
			src:     `fn f() {}`,
			wantSig: "fn f()",
		},
		{
			name:    "with params",
			src:     `fn f(x: i32, y: &str) {}`,
			wantSig: "fn f(x: i32, y: &str)",
		},
		{
			name:    "with return",
			src:     `fn f() -> i32 { 0 }`,
			wantSig: "fn f() -> i32",
		},
		// Note: Generics/lifetimes are in type_parameters node, not included in signature for simplicity
		{
			name:    "with lifetime",
			src:     `fn f<'a>(x: &'a str) -> &'a str { x }`,
			wantSig: "fn f(x: &'a str) -> &'a str",
		},
		{
			name:    "with generics",
			src:     `fn f<T: Clone>(x: T) -> T { x }`,
			wantSig: "fn f(x: T) -> T",
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

func TestParseVisibilityModifiers(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantVis string
	}{
		{
			name:    "pub",
			src:     `pub fn f() {}`,
			wantVis: "pub",
		},
		{
			name:    "pub(crate)",
			src:     `pub(crate) fn f() {}`,
			wantVis: "pub(crate)",
		},
		{
			name:    "pub(super)",
			src:     `pub(super) fn f() {}`,
			wantVis: "pub(super)",
		},
		{
			name:    "private (no modifier)",
			src:     `fn f() {}`,
			wantVis: "",
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

			fn, ok := symbols[0].(*Function)
			if !ok {
				t.Fatalf("expected *Function, got %T", symbols[0])
			}

			if fn.visibility != tt.wantVis {
				t.Errorf("expected visibility %q, got %q", tt.wantVis, fn.visibility)
			}
		})
	}
}
