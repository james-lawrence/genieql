package main

import (
	"context"
	"log"

	"github.com/egdaemon/eg/runtime/wasi/eg"
	"github.com/egdaemon/eg/runtime/wasi/egenv"
	"github.com/egdaemon/eg/runtime/wasi/eggit"
	"github.com/egdaemon/eg/runtime/wasi/shell"
	"github.com/egdaemon/eg/runtime/x/wasi/eggolang"
	"github.com/egdaemon/eg/runtime/x/wasi/egpostgresql"
)

func Setup(ctx context.Context, id eg.Op) error {
	runtime := eggolang.Runtime().
		Environ("GOBIN", "/usr/local/bin")

	return shell.Run(
		ctx,

		runtime.Newf("ls -lha %s", egenv.CacheDirectory()).Lenient(true),
		runtime.Newf("ls -lha %s", egenv.CacheDirectory(".eg")).Lenient(true),
		runtime.Newf("tree -L 2 %s", egenv.CacheDirectory()).Lenient(true),
		runtime.Newf("ls -lha %s", eggolang.CacheDirectory()).Lenient(true),
		runtime.Newf("ls -lha %s", eggolang.CacheBuildDirectory()).Lenient(true),
		runtime.New("ls -lha ."),
		runtime.New("go install -tags genieql.duckdb,no_duckdb_arrow ./..."),
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
			egpostgresql.Auto,
			Setup,
			eggolang.AutoCompile(eggolang.CompileOptionTags("genieql.duckdb", "no_duckdb_arrow")),
			eggolang.AutoTest(eggolang.TestOptionTags("genieql.duckdb", "no_duckdb_arrow"))),
	)

	if err != nil {
		log.Fatalln(err)
	}
}
