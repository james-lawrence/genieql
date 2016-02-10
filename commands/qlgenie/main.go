package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
)

func main() {
	bootstrap := &bootstrap{}
	mapper := &mapper{}
	generator := &generate{crud: &generateCrud{}}

	app := kingpin.New("qlgenie", "query language genie - a tool for interfacing with databases")
	bootstrapCmd := bootstrap.configure(app)
	mapCmd := mapper.configure(app)
	generator.configure(app)

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch cmd {
	case bootstrapCmd.FullCommand():
		if err := genieql.Bootstrap(bootstrap.outputfilepath, bootstrap.dburi); err != nil {
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
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	return filepath.Join(wd, ".genieql")
}
