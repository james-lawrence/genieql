package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin"
	_ "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/debugx"
	"bitbucket.org/jatone/genieql/internal/envx"
	"bitbucket.org/jatone/genieql/internal/stringsx"

	// register the drivers
	_ "bitbucket.org/jatone/genieql/internal/drivers"
	// register the postgresql dialect
	_ "bitbucket.org/jatone/genieql/internal/postgresql"
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
	go debugx.OnSignal(context.Background(), func(ctx context.Context) error {
		dctx, done := context.WithTimeout(ctx, envx.Duration(30*time.Second, "GENIEQL_PROFILING_DURATION"))
		defer done()
		switch envx.String("cpu", "GENIEQL_PROFILING_STRATEGY") {
		case "heap":
			return debugx.Heap(envx.String(os.TempDir(), "CACHE_DIRECTORY"))(dctx)
		case "mem":
			return debugx.Memory(envx.String(os.TempDir(), "CACHE_DIRECTORY"))(dctx)
		default:
			return debugx.CPU(envx.String(os.TempDir(), "CACHE_DIRECTORY"))(dctx)
		}
	}, syscall.SIGUSR1)

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
