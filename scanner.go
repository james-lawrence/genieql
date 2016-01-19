package genieql

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

// Matcher takes a set of Mappings and maps the fields
// on the structures to their position in the sql query.
// returns a Array of Matches.
type Matcher struct {
	Maps []Mapping
}

func (t Matcher) Match(columns ...string) []Match {
	matches := make([]Match, 0, len(columns))

	for idx, column := range columns {
		for jdx, m := range t.Maps {
			field, matched := m.FindMatch(column)
			if matched {
				// convert to debug
				// fmt.Printf("%s is an alias %s.%s: position %d\n", column, m.Type, field, idx)
				matches = append(matches, Match{
					Mapping:      m,
					ArgPosition:  jdx,
					Field:        field,
					ScanPosition: idx,
				})
				continue
			}
		}
	}

	return matches
}

type ScannerBuilder struct {
	Name      string
	Arguments []Mapping
	Matches   []Match
}

func (t ScannerBuilder) WriteScanner(dst io.Writer) error {
	fmap := template.FuncMap{
		"FunctionArgs":  FunctionArgs,
		"ArgVariable":   ArgVariable,
		"LocalVariable": LocalVariable,
		"ScannerSuffix": ScannerSuffix,
		"ScanArgs":      ScanArgs,
		"Lowercase":     strings.ToLower,
	}

	tmpl := template.Must(template.New("").Funcs(fmap).Parse(scanner))
	return tmpl.Execute(dst, t)
}

func ScannerSuffix(s string) string {
	return s + "Scanner"
}

func ArgVariable(pos int) string {
	return fmt.Sprintf("arg%d", pos)
}

func LocalVariable(m Match) string {
	return fmt.Sprintf("c%d", m.ScanPosition)
}

func ScanArgs(matches []Match) string {
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, LocalVariable(m))
	}
	return "&" + strings.Join(names, ", &")
}

func FunctionArgs(maps []Mapping) string {
	args := make([]string, 0, len(maps))
	for idx, m := range maps {
		args = append(args, fmt.Sprintf("arg%d *%s.%s", idx, m.Package, m.Type))
	}

	return strings.Join(args, ", ")
}
