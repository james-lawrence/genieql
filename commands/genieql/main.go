package main

import (
	"log"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	_ "github.com/lib/pq"

	// register the postgresql dialect
	"bitbucket.org/jatone/genieql"
	_ "bitbucket.org/jatone/genieql/internal/postgresql"

	// register the drivers
	_ "bitbucket.org/jatone/genieql/internal/drivers"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	bootstrap := &bootstrap{}
	mapper := &mapper{}
	generator := &generate{}
	scanner := &scanners{}

	app := kingpin.New("genieql", "query language genie - a tool for interfacing with databases")
	bootstrapCmd := bootstrap.configure(app)
	mapCmd := mapper.configure(app)
	generator.configure(app)
	scanner.configure(app)

	if os.Getenv("DEBUGPANIC") != "" {
		defer func() {
			if r := recover(); r != nil {
				log.Println(r)
				log.Fatalln("panic debug", genieql.PrintDebug())
				panic(r)
			}
		}()
	}

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
	if i := strings.LastIndex(s, "."); i > -1 {
		return s[:i], s[i+1:]
	}
	return "", ""
}
