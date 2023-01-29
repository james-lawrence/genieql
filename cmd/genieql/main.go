package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/alecthomas/kingpin"
	_ "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	// register the postgresql dialect
	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
	_ "bitbucket.org/jatone/genieql/internal/postgresql"
	"bitbucket.org/jatone/genieql/internal/stringsx"

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
	astcli := astcli{}

	gg := generator{
		buildInfo: &bi,
	}

	app := kingpin.New("genieql", "query language genie - a tool for interfacing with databases")
	app.Command("version", "print version").Action(func(*kingpin.ParseContext) error {
		if bi, ok := debug.ReadBuildInfo(); ok {
			fmt.Println(bi.Main.Version)
		}
		return nil
	})

	app.Flag("verbose", "increase logging").Short('v').Default("0").CounterVar(&bi.Verbosity)

	astcli.configure(app)
	bootstrap.configure(app)
	gg.configure(app)

	if cmd, err := app.Parse(os.Args[1:]); err != nil {
		fmts := "%s\n"
		if bi.Verbosity >= generators.VerbosityDebug {
			fmts = "%+v\n"
		}

		log.Printf(fmts, errors.Wrap(err, stringsx.DefaultIfBlank(cmd, fmt.Sprintf("parsing failed: %s", strings.Join(os.Args, " ")))))
		log.Fatalln(genieql.PrintDebug())
	}
}
