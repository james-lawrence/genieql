package main

import (
	"context"
	"log"

	"github.com/egdaemon/eg/runtime/wasi/eg"
	"github.com/egdaemon/eg/runtime/wasi/egenv"
	"github.com/egdaemon/eg/runtime/wasi/eggit"
	"github.com/egdaemon/eg/runtime/wasi/shell"
	"github.com/egdaemon/eg/runtime/x/wasi/egbug"
	"github.com/egdaemon/eg/runtime/x/wasi/eggolang"
	"github.com/egdaemon/eg/runtime/x/wasi/egpostgresql"
)

func Setup(ctx context.Context, id eg.Op) error {
	runtime := shell.Runtime().
		Environ("GOBIN", "/usr/local/bin").
		Environ("USER", "root").
		Environ("LD_LIBRARY_PATH", "/usr/local/lib").
		Environ("GOCACHE", eggolang.CacheBuildDirectory()).
		Environ("GOMODCACHE", eggolang.CacheModuleDirectory())

	return shell.Run(
		ctx,
		runtime.New("go install ./..."),
		runtime.New("genieql bootstrap --queryer=sqlx.Queryer --driver=github.com/jackc/pgx postgres://root@localhost:5432/genieql_examples?sslmode=disable"),
		runtime.New("go generate ./..."),
		runtime.New("go fmt ./..."),
	)
}

func main() {
	ctx, done := context.WithTimeout(context.Background(), egenv.TTL())
	defer done()

	c1 := eg.Container("genieql.ubuntu.24.04")

	err := eg.Perform(
		ctx,
		eggit.AutoClone,
		eg.Build(
			c1.BuildFromFile(".eg/Containerfile"),
		),
		eg.Module(
			ctx,
			c1,
			egbug.Debug,
			egpostgresql.Auto,
			Setup,
			eggolang.AutoCompile(eggolang.CompileOptionTags("no_duckdb_arrow", "duckdb_use_lib")),
			eggolang.AutoTest(eggolang.TestOptionTags("no_duckdb_arrow", "duckdb_use_lib"))),
	)

	if err != nil {
		log.Fatalln(err)
	}
}
