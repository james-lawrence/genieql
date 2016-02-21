package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	bootstrap := &bootstrap{}
	mapper := &mapper{}
	generator := &generate{crud: &generateCrud{}}
	scanner := &scanners{}

	app := kingpin.New("qlgenie", "query language genie - a tool for interfacing with databases")
	bootstrapCmd := bootstrap.configure(app)
	mapCmd := mapper.configure(app)
	generator.configure(app)
	scanner.configure(app)

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch cmd {
	case bootstrapCmd.FullCommand():
		if err := bootstrap.Bootstrap(); err != nil {
			log.Fatalln(err)
		}
	case mapCmd.FullCommand():
		if err := mapper.Map(); err != nil {
			log.Fatalln(err)
		}
	}
}

func extractPackageType(s string) (string, string) {
	i := strings.LastIndex(s, ".")
	return s[:i], s[i+1:]
}

func configurationDirectory() string {
	var err error
	var defaultPath string
	paths := filepath.SplitList(os.Getenv("GOPATH"))
	if len(paths) == 0 {
		if defaultPath, err = os.Getwd(); err != nil {
			log.Fatalln(err)
		}
	} else {
		defaultPath = paths[0]
	}

	return filepath.Join(defaultPath, ".genieql")
}
