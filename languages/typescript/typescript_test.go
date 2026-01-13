package typescript

import (
	"strings"
	"testing"
)

func TestLanguageMetadata(t *testing.T) {
	tests := []struct {
		lang interface {
			Name() string
			Extensions() []string
		}
		wantName string
		wantExts []string
	}{
		{&TSLanguage{}, "typescript", []string{".ts"}},
		{&TSXLanguage{}, "tsx", []string{".tsx"}},
		{&JSLanguage{}, "javascript", []string{".js", ".mjs", ".cjs"}},
		{&JSXLanguage{}, "jsx", []string{".jsx"}},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			if tt.lang.Name() != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, tt.lang.Name())
			}
			exts := tt.lang.Extensions()
			if len(exts) != len(tt.wantExts) {
				t.Errorf("expected %d extensions, got %d", len(tt.wantExts), len(exts))
			}
			for i, ext := range tt.wantExts {
				if exts[i] != ext {
					t.Errorf("extension[%d]: expected %q, got %q", i, ext, exts[i])
				}
			}
		})
	}
}

func TestParseFunction(t *testing.T) {
	src := `/** Greet someone */
function greet(name: string): string {
    return "Hello, " + name;
}
`
	lang := &TSLanguage{}
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
	if !strings.Contains(str, "function greet") {
		t.Errorf("expected String() to contain 'function greet', got %q", str)
	}
	if !strings.Contains(str, ": string") {
		t.Errorf("expected String() to contain ': string', got %q", str)
	}

	// Check doc comment
	if doc, ok := sym.(interface{ DocComment() string }); ok {
		if doc.DocComment() != "Greet someone" {
			t.Errorf("expected doc comment 'Greet someone', got %q", doc.DocComment())
		}
	}
}

func TestParseAsyncFunction(t *testing.T) {
	src := `async function fetchData(url: string): Promise<Data> {
    return fetch(url).then(r => r.json());
}
`
	lang := &TSLanguage{}
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

	if !fn.isAsync {
		t.Error("expected isAsync to be true")
	}

	str := fn.String()
	if !strings.HasPrefix(str, "async function") {
		t.Errorf("expected String() to start with 'async function', got %q", str)
	}
}

func TestParseClass(t *testing.T) {
	src := `/** Server implementation */
class Server extends EventEmitter implements Handler {
    private port: number;
    
    constructor(port: number) {
        this.port = port;
    }
}
`
	lang := &TSLanguage{}
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
	if !strings.Contains(str, "extends EventEmitter") {
		t.Errorf("expected String() to contain 'extends EventEmitter', got %q", str)
	}
	if !strings.Contains(str, "implements Handler") {
		t.Errorf("expected String() to contain 'implements Handler', got %q", str)
	}

	if cls.DocComment() != "Server implementation" {
		t.Errorf("expected doc comment 'Server implementation', got %q", cls.DocComment())
	}
}

func TestParseInterface(t *testing.T) {
	src := `/** Configuration options */
interface Config {
    host: string;
    port: number;
}
`
	lang := &TSLanguage{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	iface, ok := symbols[0].(*Interface)
	if !ok {
		t.Fatalf("expected *Interface, got %T", symbols[0])
	}

	if iface.Name() != "Config" {
		t.Errorf("expected name 'Config', got %q", iface.Name())
	}
	if iface.Kind() != "interface" {
		t.Errorf("expected kind 'interface', got %q", iface.Kind())
	}
	if iface.String() != "interface Config" {
		t.Errorf("expected 'interface Config', got %q", iface.String())
	}
}

func TestParseTypeAlias(t *testing.T) {
	src := `type Handler = (req: Request, res: Response) => void;
`
	lang := &TSLanguage{}
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

	if typ.Name() != "Handler" {
		t.Errorf("expected name 'Handler', got %q", typ.Name())
	}
	if typ.Kind() != "type" {
		t.Errorf("expected kind 'type', got %q", typ.Kind())
	}
	if typ.String() != "type Handler" {
		t.Errorf("expected 'type Handler', got %q", typ.String())
	}
}

func TestParseEnum(t *testing.T) {
	src := `enum Status {
    Pending,
    Active,
    Completed
}
`
	lang := &TSLanguage{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	enum, ok := symbols[0].(*Enum)
	if !ok {
		t.Fatalf("expected *Enum, got %T", symbols[0])
	}

	if enum.Name() != "Status" {
		t.Errorf("expected name 'Status', got %q", enum.Name())
	}
	if enum.Kind() != "enum" {
		t.Errorf("expected kind 'enum', got %q", enum.Kind())
	}
	if enum.String() != "enum Status" {
		t.Errorf("expected 'enum Status', got %q", enum.String())
	}
}

func TestParseVariables(t *testing.T) {
	src := `const PORT = 8080;
let counter = 0;
var legacy = "old";
`
	lang := &TSLanguage{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(symbols))
	}

	tests := []struct {
		name string
		kind string
		str  string
	}{
		{"PORT", "const", "const PORT"},
		{"counter", "let", "let counter"},
		{"legacy", "var", "var legacy"},
	}

	for i, tt := range tests {
		sym := symbols[i]
		if sym.Name() != tt.name {
			t.Errorf("symbol[%d]: expected name %q, got %q", i, tt.name, sym.Name())
		}
		if sym.Kind() != tt.kind {
			t.Errorf("symbol[%d]: expected kind %q, got %q", i, tt.kind, sym.Kind())
		}
		if sym.String() != tt.str {
			t.Errorf("symbol[%d]: expected %q, got %q", i, tt.str, sym.String())
		}
	}
}

func TestParseImports(t *testing.T) {
	src := `import { useState } from 'react';
import express from 'express';
import * as fs from 'fs';

function main() {}
`
	lang := &TSLanguage{}
	imports, _, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := []string{"react", "express", "fs"}
	if len(imports) != len(expected) {
		t.Fatalf("expected %d imports, got %d: %v", len(expected), len(imports), imports)
	}

	for i, exp := range expected {
		if imports[i] != exp {
			t.Errorf("import[%d]: expected %q, got %q", i, exp, imports[i])
		}
	}
}

func TestParseExportedDeclarations(t *testing.T) {
	src := `export function publicFunc() {}
export class PublicClass {}
export interface PublicInterface {}
export type PublicType = string;
export const PUBLIC_CONST = 1;
`
	lang := &TSLanguage{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 5 {
		t.Fatalf("expected 5 symbols, got %d", len(symbols))
	}

	// Just verify we got all types
	kinds := make(map[string]bool)
	for _, sym := range symbols {
		kinds[sym.Kind()] = true
	}

	expectedKinds := []string{"func", "class", "interface", "type", "const"}
	for _, k := range expectedKinds {
		if !kinds[k] {
			t.Errorf("expected to find kind %q", k)
		}
	}
}

func TestParseJavaScript(t *testing.T) {
	src := `function greet(name) {
    return "Hello, " + name;
}

class Server {
    constructor(port) {
        this.port = port;
    }
}

const PORT = 8080;
`
	lang := &JSLanguage{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(symbols))
	}

	// Function
	if symbols[0].Name() != "greet" || symbols[0].Kind() != "func" {
		t.Errorf("expected func 'greet', got %s %q", symbols[0].Kind(), symbols[0].Name())
	}

	// Class
	if symbols[1].Name() != "Server" || symbols[1].Kind() != "class" {
		t.Errorf("expected class 'Server', got %s %q", symbols[1].Kind(), symbols[1].Name())
	}

	// Const
	if symbols[2].Name() != "PORT" || symbols[2].Kind() != "const" {
		t.Errorf("expected const 'PORT', got %s %q", symbols[2].Kind(), symbols[2].Name())
	}
}

func TestParseEmptyFile(t *testing.T) {
	src := `// Just a comment`

	lang := &TSLanguage{}
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
			src:     `function f() {}`,
			wantSig: "function f()",
		},
		{
			name:    "with params",
			src:     `function f(x: number, y: string) {}`,
			wantSig: "function f(x: number, y: string)",
		},
		{
			name:    "with return type",
			src:     `function f(): boolean { return true; }`,
			wantSig: "function f(): boolean",
		},
		{
			name:    "optional param",
			src:     `function f(x?: number) {}`,
			wantSig: "function f(x?: number)",
		},
		{
			name:    "rest params",
			src:     `function f(...args: string[]) {}`,
			wantSig: "function f(...args: string[])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := &TSLanguage{}
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
