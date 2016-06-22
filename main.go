package main

import (
	"fmt"
	"regexp"
	"sort"
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
	builtin = []string{
		"true", "false", "iota", "nil",
		"append", "cap", "close", "complex", "copy", "delete", "imag",
		"len", "make", "new", "panic", "print", "println", "real", "recover",
		"bool", "byte", "complex128", "complex64", "error", "float32", "float64",
		"int", "int16", "int32", "int64", "int8",
		"rune", "string",
		"uint", "uint16", "uint32", "uint64", "uint8", "uintptr",
	}
)

var (
	Words   = make(map[string]struct{}, 1000)
	wordsRE = regexp.MustCompile(`^([[:alpha:]]+)`)
)

func addWords(words []string) {
	var m []string
	for _, w := range words {
		m = wordsRE.FindStringSubmatch(w)
		if len(m) > 1 {
			Words[m[1]] = struct{}{}
		}
	}
}

func main() {
	addWords(keywords)
	addWords(builtin)

	res := make([]string, 0, len(Words))
	for w := range Words {
		res = append(res, w)
	}
	sort.Strings(res)

	for _, w := range res {
		fmt.Println(w)
	}
}
