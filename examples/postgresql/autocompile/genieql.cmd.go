package autocompile

// pkga must be generated first: CombinedScanner2 in genieql.input.go references
// pkga.Example1, and genieql auto (unlike genieql auto graph) only resolves
// mapping info for types already generated on disk.
//go:generate go generate ./pkga/...
//go:generate genieql auto -o "genieql.gen.go"
