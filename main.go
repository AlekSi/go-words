package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	Words = make(map[string]struct{}, 7000)

	// https://golang.org/ref/spec#Keywords
	Keywords = []string{
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
	}

	// https://golang.org/pkg/builtin/
	Builtins = []string{
		"true", "false", "iota", "nil",
		"append", "cap", "close", "complex", "copy", "delete", "imag",
		"len", "make", "new", "panic", "print", "println", "real", "recover",
		"bool", "byte", "complex128", "complex64", "error", "float32", "float64",
		"int", "int16", "int32", "int64", "int8",
		"rune", "string",
		"uint", "uint16", "uint32", "uint64", "uint8", "uintptr",
	}

	Extra = []string{
		"omitempty", // popular tag value in std
		"gopath", "goroot",
		"goroutine", "goroutines",

		// build tags
		"386", "amd64", "amd64p32", "arm", "arm64", "mips64", "mips64le", "ppc64", "ppc64le", // okgoarch
		"darwin", "dragonfly", "linux", "android", "solaris", "freebsd", "nacl", "netbsd", "openbsd", "plan9", "windows", // okgoos
		"gc", "gcc", "cgo",
		"go1.1", "go1.2", "go1.3", "go1.4", "go1.5", "go1.6", "go1.7",
	}
)

var (
	DebugF = flag.Bool("debug", false, "Enable debug output")
)

func debugf(format string, v ...interface{}) {
	if *DebugF {
		log.Printf(format, v...)
	}
}

// addWords adds words to Words map.
func addWords(words ...string) {
	for _, w := range words {
		Words[w] = struct{}{}
	}
}

// processIdent extracts words from ident and adds them to Words.
func processIdent(ident *ast.Ident) {
	if ident == nil {
		return
	}

	if strings.Contains(ident.Name, ".") {
		log.Fatalf("unhandled ident %q", ident.Name)
	}

	if ast.IsExported(ident.Name) {
		addWords(ident.Name)
	}
}

// processAST extracts words from f.
func processAST(f *ast.File) {
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ImportSpec:
					processIdent(spec.Name)

				case *ast.ValueSpec:
					for _, n := range spec.Names {
						processIdent(n)
					}

				case *ast.TypeSpec:
					processIdent(spec.Name)

				default:
					log.Fatalf("unhandled spec %#v", spec)
				}
			}

		case *ast.FuncDecl:
			processIdent(decl.Name)

		default:
			log.Fatalf("unhandled decl %#v", decl)
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	// add keywords, builtins and some extra words
	addWords(Keywords...)
	addWords(Builtins...)
	addWords(Extra...)

	// get std packages
	cmd := exec.Command("go", "list", "std")
	b, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	packages := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")

	// process packages
	for _, p := range packages {
		pack, err := build.Import(p, "", 0)
		if err != nil {
			log.Fatal(err)
		}

		// skip internal packages like "internal/trace" or "net/http/internal"
		if strings.Contains(pack.ImportPath, "internal") {
			debugf("skipping internal package %q", pack.ImportPath)
			continue
		}
		debugf("processing package %q", pack.ImportPath)

		addWords(pack.Name)

		fset := token.NewFileSet()
		for _, f := range pack.GoFiles {
			ast, err := parser.ParseFile(fset, filepath.Join(pack.Dir, f), nil, 0)
			if err != nil {
				log.Fatal(err)
			}
			processAST(ast)
		}
	}

	// sort and print result
	log.Printf("found %d words", len(Words))
	res := make([]string, 0, len(Words))
	for w := range Words {
		res = append(res, w)
	}
	sort.Strings(res)
	for _, w := range res {
		fmt.Println(w)
	}
}
