package main

import (
	"context"
	"log"

	"github.com/egdaemon/eg/runtime/wasi/eg"
	"github.com/egdaemon/eg/runtime/wasi/egenv"
	"github.com/egdaemon/eg/runtime/wasi/eggit"
	"github.com/egdaemon/eg/runtime/wasi/shell"
	"github.com/egdaemon/eg/runtime/x/wasi/eggolang"
)

func Debug(ctx context.Context, id eg.Op) error {
	return shell.Run(
		ctx,
		shell.New("echo ${PATH}"),
	)
}

func Setup(ctx context.Context, id eg.Op) error {
	runtime := shell.Runtime().
		Environ("GOBIN", "/usr/local/bin").
		Environ("USER", "root")
	return shell.Run(
		ctx,
		runtime.New("pg_isready").Attempts(15), // 15 attempts = ~3seconds
		runtime.New("su postgres -l -c 'psql --no-psqlrc -U postgres -d postgres -c \"CREATE ROLE root WITH SUPERUSER LOGIN\"'"),
		runtime.New("go generate ./... && go fmt ./..."),
	)
}

func main() {
	ctx, done := context.WithTimeout(context.Background(), egenv.TTL())
	defer done()

	c1 := eg.Container("genieql.ubuntu.22.04")

	err := eg.Perform(
		ctx,
		eggit.AutoClone,
		eg.Build(
			c1.BuildFromFile(".eg/Containerfile"),
		),
		eg.Module(ctx, c1, Debug, eggolang.AutoCompile(), Setup, eggolang.AutoTest()),
	)

	if err != nil {
		log.Fatalln(err)
	}
}
