package state

import (
	"path"
	"runtime"
)

func forceRelativeToRoot(s string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unreachable")
	}
	return path.Join(path.Dir(filename), s)
}

var (
    // this is exceptionally shitty code, but it allows our tests to actually get the database from root
    DatabasePath = forceRelativeToRoot("../app.sqlite")
)
