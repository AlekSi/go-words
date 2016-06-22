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
	"regexp"
	"sort"
	"strings"
)

var (
	// https://golang.org/ref/spec#Keywords
	keywords = []string{
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
	}

	// https://golang.org/pkg/builtin/
	builtins = []string{
		"true", "false", "iota", "nil",
		"append", "cap", "close", "complex", "copy", "delete", "imag",
		"len", "make", "new", "panic", "print", "println", "real", "recover",
		"bool", "byte", "complex128", "complex64", "error", "float32", "float64",
		"int", "int16", "int32", "int64", "int8",
		"rune", "string",
		"uint", "uint16", "uint32", "uint64", "uint8", "uintptr",
	}

	extra = []string{
		"omitempty", // popular tag value in std
	}
)

var (
	Words   = make(map[string]struct{}, 1000)
	wordsRE = regexp.MustCompile(`^([[:alpha:]]+)`)
)

var DebugF = flag.Bool("debug", false, "Enable debug output")

func debugf(format string, v ...interface{}) {
	if *DebugF {
		log.Printf(format, v...)
	}
}

// addWords adds words to Words map, stripping non-alpha suffixes.
// For example, "uint8" and "uint16" both result in "uint".
func addWords(words ...string) {
	var m []string
	for _, w := range words {
		m = wordsRE.FindStringSubmatch(w)
		if len(m) > 1 {
			debugf("adding %q", m[1])
			Words[m[1]] = struct{}{}
		}
	}
}

// processIdent extracts words from ident and adds them to Words.
func processIdent(ident *ast.Ident) {
	if ident == nil {
		return
	}

	if strings.Contains(ident.Name, ".") {
		log.Fatal("unhandled ident %q", ident.Name)
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
	addWords(keywords...)
	addWords(builtins...)
	addWords(extra...)

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
	res := make([]string, 0, len(Words))
	for w := range Words {
		res = append(res, w)
	}
	sort.Strings(res)
	for _, w := range res {
		fmt.Println(w)
	}
}
