package scanner

import (
	"io/ioutil"
	"strings"
)

func ReadString(path string) string {
	raw, err := ioutil.ReadFile(path)
	maybePanic(err)
	return strings.TrimSpace(string(raw))
}

func maybePanic(err error) {
	if err != nil {
		panic(err)
	}
}
