package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/alecthomas/kingpin"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	// register the postgresql dialect
	"bitbucket.org/jatone/genieql"
	_ "bitbucket.org/jatone/genieql/internal/postgresql"
	"bitbucket.org/jatone/genieql/internal/x/stringsx"

	// register the drivers
	_ "bitbucket.org/jatone/genieql/internal/drivers"
)

func main() {
	defer func() {
		r := recover()
		switch err := r.(type) {
		case runtime.Error:
			log.Println(genieql.PrintDebug())
			log.Fatalln(string(debug.Stack()))
		case error:
			log.Fatalln(errors.Wrap(err, genieql.PrintDebug()))
		}
	}()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	bi := mustBuildInfo()

	bootstrap := &bootstrap{
		buildInfo: bi,
	}
	mapper := &mapper{
		buildInfo: bi,
	}

	astcli := astcli{}

	gg := generator{
		buildInfo: bi,
	}

	generator := &generate{
		buildInfo: bi,
	}
	scanner := &scanners{
		buildInfo: bi,
	}
	app := kingpin.New("genieql", "query language genie - a tool for interfacing with databases")
	app.Flag("debug", "enable debug logging").BoolVar(&bi.DebugEnabled)

	astcli.configure(app)
	bootstrap.configure(app)
	mapper.configure(app)
	generator.configure(app)
	gg.configure(app)
	scanner.configure(app)

	if cmd, err := app.Parse(os.Args[1:]); err != nil {
		fmts := "%s\n"
		if bi.DebugEnabled {
			fmts = "%+v\n"
		}
		log.Printf(fmts, errors.Wrap(err, stringsx.DefaultIfBlank(cmd, fmt.Sprintf("parsing failed: %s", strings.Join(os.Args, " ")))))
		log.Fatalln(genieql.PrintDebug())
	}

}
