package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/alecthomas/kingpin"
	_ "github.com/duckdb/duckdb-go/v2"
	_ "github.com/jackc/pgx/v5"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/generators"
	"github.com/james-lawrence/genieql/internal/debugx"
	"github.com/james-lawrence/genieql/internal/envx"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/stringsx"

	// register the drivers
	_ "github.com/james-lawrence/genieql/internal/drivers"
	// register the postgresql dialect
	_ "github.com/james-lawrence/genieql/internal/postgresql"
	// register the duckdb dialect
	_ "github.com/james-lawrence/genieql/internal/duckdb"
)

func main() {
	defer func() {
		r := recover()
		switch err := r.(type) {
		case runtime.Error:
			log.Println(genieql.PrintDebug())
			log.Fatalln(string(debug.Stack()))
		case error:
			log.Fatalln(errorsx.Wrap(err, genieql.PrintDebug()))
		}
	}()

	var (
		pmode string
	)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	bi := errorsx.Must(genieql.NewBuildInfo())

	bootstrap := &bootstrap{
		BuildInfo: bi,
	}
	astcli := astcli{}

	gg := generator{
		BuildInfo: &bi,
	}

	duckdb := duckdb{}

	bg := &sync.WaitGroup{}
	defer bg.Wait()

	ctx, done := context.WithCancelCause(context.Background())
	defer func() {
		errorsx.Log(errorsx.Ignore(context.Cause(ctx), context.Canceled))
	}()
	defer done(nil)

	app := kingpin.New("genieql", "query language genie - a tool for interfacing with databases")
	app.Flag("profile", "enable profiler with the specified mode: trace,heap,mem,alloc,block,cpu").Short('p').Action(func(pc *kingpin.ParseContext) error {
		bg.Add(1)
		go func() {
			defer bg.Done()
			done(errorsx.Wrap(debugx.Profile(ctx, envx.String(pmode, "GENIEQL_PROFILING_STRATEGY")), "profiling failed"))
		}()
		return nil
	}).StringVar(&pmode)

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
	duckdb.configure(app)

	if cmd, err := app.Parse(os.Args[1:]); err != nil {
		fmts := "%s\n"
		if bi.Verbosity >= generators.VerbosityDebug {
			fmts = "%+v\n"
		}

		log.Printf(fmts, errorsx.Wrap(err, stringsx.DefaultIfBlank(cmd, fmt.Sprintf("parsing failed: %s", strings.Join(os.Args, " ")))))
		log.Fatalln(genieql.PrintDebug())
	}
}
