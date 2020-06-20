package debugx

import (
	"fmt"
	"io/ioutil"
	"log"
)

var (
	defaults = log.New(ioutil.Discard, "DEBUG", log.LstdFlags|log.Lshortfile)
)

// Output debug output
func Output(d int, s string) error {
	return defaults.Output(d, s)
}

// Println debug output
func Println(args ...interface{}) {
	Output(2, fmt.Sprintln(args...))
}

// Printf debug output
func Printf(format string, args ...interface{}) {
	Output(2, fmt.Sprintf(format, args...))
}
