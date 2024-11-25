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

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/generators"
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
			log.Fatalln(errors.Wrap(err, genieql.PrintDebug()))
		}
	}()

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
	// go debugx.OnSignal(context.Background(), func(ctx context.Context) error {
	// 	dctx, done := context.WithTimeout(ctx, envx.Duration(30*time.Second, "GENIEQL_PROFILING_DURATION"))
	// 	defer done()
	// 	switch envx.String("cpu", "GENIEQL_PROFILING_STRATEGY") {
	// 	case "heap":
	// 		return debugx.Heap(envx.String(os.TempDir(), "CACHE_DIRECTORY"))(dctx)
	// 	case "mem":
	// 		return debugx.Memory(envx.String(os.TempDir(), "CACHE_DIRECTORY"))(dctx)
	// 	default:
	// 		return debugx.CPU(envx.String(os.TempDir(), "CACHE_DIRECTORY"))(dctx)
	// 	}
	// }, syscall.SIGUSR1)

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
	duckdb.configure(app)

	if cmd, err := app.Parse(os.Args[1:]); err != nil {
		fmts := "%s\n"
		if bi.Verbosity >= generators.VerbosityDebug {
			fmts = "%+v\n"
		}

		log.Printf(fmts, errors.Wrap(err, stringsx.DefaultIfBlank(cmd, fmt.Sprintf("parsing failed: %s", strings.Join(os.Args, " ")))))
		log.Fatalln(genieql.PrintDebug())
	}
}
