//go:build spike
// +build spike

package main

import (
	"go/build"
	"log"

	"bitbucket.org/jatone/genieql/compiler/stdlib"
	"github.com/traefik/yaegi/interp"
	// "github.com/traefik/yaegi/stdlib"
)

func main() {
	// 2023/01/28 11:35:50 failed to eval 1:21: import "bitbucket.org/jatone/genieql/generators"
	// error: /home/james.lawrence/development/genieql/generators/column_constants.go:13:2: import "bitbucket.org/jatone/genieql"
	//  error: /home/james.lawrence/development/genieql/genieql.go:12:2: import "golang.org/x/tools/imports"
	//  error: /home/james.lawrence/development/genieql/vendor/golang.org/x/tools/imports/forward.go:13:2: import "golang.org/x/tools/internal/gocommand"
	//  error: /home/james.lawrence/development/genieql/vendor/golang.org/x/tools/internal/gocommand/invoke.go:20:2: import "golang.org/x/sys/execabs"
	//  error: /home/james.lawrence/development/genieql/vendor/golang.org/x/sys/execabs/execabs.go:19:2: import "os/exec"
	//  error: /usr/lib/go/src/os/exec/exec.go:97:2: import "internal/syscall/execenv"
	//  error: /usr/lib/go/src/internal/syscall/execenv/execenv_default.go:9:8: import "syscall"
	//  error: /usr/lib/go/src/syscall/asan.go:11:2: import "unsafe"
	//  error: /usr/lib/go/src/unsafe/unsafe.go:196:1: function declaration without body is unsupported (linkname or assembly can not be interpreted).
	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})
	for name := range stdlib.Symbols {
		log.Println("DERP", name)
	}
	// stdlib.Symbols["os/exec"] = map[string]reflect.Value{}
	i.Use(stdlib.Symbols)
	// if _, err := i.Eval(`import "bitbucket.org/jatone/genieql/astutil"`); err != nil {
	// 	log.Println("failed to eval", err)
	// }
	if _, err := i.Eval(`import "bitbucket.org/jatone/genieql/generators"`); err != nil {
		log.Println("failed to eval", err)
	}

}
