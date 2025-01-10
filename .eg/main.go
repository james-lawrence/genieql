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
	runtime := eggolang.Runtime()

	return shell.Run(
		ctx,
		runtime.New("id"),
		runtime.New("env"),
		runtime.New("ls -lha ."),
		runtime.New("ls -lha examples"),
		runtime.New("tree -L 3 examples"),
		runtime.New("ls -lha examples/postgresql/structure.sql"),
		runtime.New("go install -tags genieql.duckdb,no_duckdb_arrow ./...").Environ("GOBIN", egenv.EphemeralDirectory()),
		runtime.Newf("cp %s/genieql /usr/local/bin", egenv.EphemeralDirectory()).Privileged(),
		// runtime.New("dropdb --if-exists -U postgres genieql_examples"),
		// runtime.New("createdb -U postgres genieql_examples \"genieql example database\""),
		// runtime.New("psql -X -U postgres -d genieql_examples --file=examples/postgresql/structure.sql"),
		// runtime.New("genieql bootstrap --queryer=sqlx.Queryer --driver=github.com/jackc/pgx postgres://egd@localhost:5432/genieql_examples?sslmode=disable"),
		runtime.New("go generate ./..."),
		runtime.New("go fmt ./..."),
	)
}

func Debug(ctx context.Context, id eg.Op) error {
	// runtime := shell.Runtime()

	return shell.Run(
		ctx,
		// runtime.New("pg_isready").Attempts(15).Privileged(),
		// runtime.New("cat $(psql --no-psqlrc -U postgres -d postgres -q -At -c 'SHOW hba_file;')").Privileged(),
	)
}

func Generate(ctx context.Context, id eg.Op) error {
	runtime := eggolang.Runtime()

	return shell.Run(
		ctx,
		runtime.New("go generate ./...").Privileged(),
		runtime.New("go fmt ./...").Privileged(),
	)
}

func DebugPSQL(ctx context.Context, _ eg.Op) (err error) {
	runtime := shell.Runtime()
	return shell.Run(
		ctx,
		runtime.New("psql --no-psqlrc -U postgres -d postgres -q -At -c 'SHOW hba_file;'").Privileged(),
		runtime.New("cat $(psql --no-psqlrc -U postgres -d postgres -q -At -c 'SHOW hba_file;')").Privileged(),
		runtime.New("psql --no-psqlrc -U egd -d postgres -q -At -c 'SHOW hba_file;'"),
		runtime.New("psql --version"),
	)
}

func main() {
	ctx, done := context.WithTimeout(context.Background(), egenv.TTL())
	defer done()

	c1 := eg.Container("genieql.ubuntu.24.04")

	err := eg.Perform(
		ctx,
		Debug,
		eggit.AutoClone,
		eg.Build(
			c1.BuildFromFile(".eg/Containerfile"),
		),
		eg.Module(
			ctx,
			c1,
			egpostgresql.Auto,
			Setup,
			DebugPSQL,
			Generate,
			eggolang.AutoCompile(
				eggolang.CompileOption.Tags("genieql.duckdb", "no_duckdb_arrow"),
				eggolang.CompileOption.Debug(true),
			),
			eggolang.AutoTest(
				eggolang.TestOption.Tags("genieql.duckdb", "no_duckdb_arrow"),
			),
		),
	)

	if err != nil {
		log.Fatalln(err)
	}
}
