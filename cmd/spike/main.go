package main

import (
	"go/build"
	"log"

	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
)

func main() {
	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})
	i.Use(stdlib.Symbols)
	if _, err := i.Eval(`import "golang.org/x/xerrors"`); err != nil {
		log.Println("failed to eval", err)
	}
	if _, err := i.Eval(`import "golang.org/x/xerrors"`); err != nil {
		log.Println("failed to eval", err)
	}
}
