package genieql

// Mapping - structure that holds a mapping of the aliases to fields
type Mapping struct {
	Package string
	Type    string
	Fields  []Field
}

func (t Mapping) FindMatch(column string) (Field, bool) {
	for _, field := range t.Fields {
		if field.IsAlias(column) {
			return field, true
		}
	}

	return Field{}, false
}

type Field struct {
	Type        string
	StructField string
	Aliases     []string
}

func (t Field) IsAlias(s string) bool {
	for _, alias := range t.Aliases {
		if alias == s {
			return true
		}
	}

	return false
}

type Match struct {
	Mapping
	ArgPosition int
	Field
	ScanPosition int
}
