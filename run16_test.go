// +build !go1.7

package plist

import (
	"flag"
	"testing"
	"runtime"
	"strings"
	"fmt"
)

func init() {
	flag.Parse()
}

var pattern = flag.String("pattern", "", "specify which test(s) should be executed")
var verbose = flag.Bool("verbose", false, "write whether test was done")

// This is a hack, that a bit simulate t.Run available from go1.7
func runTest(name string, fn func(t *testing.T), t *testing.T) {
	// obtain name of caller
	var pc[10]uintptr
	runtime.Callers(2, pc[:])
	var fnName = ""

	f := runtime.FuncForPC(pc[0])
	if f != nil {
		fnName = f.Name()
	}
	names := strings.Split(fnName, ".")
	fnName = names[len(names)-1] + "/" + name
	if strings.Contains(fnName, *pattern) {
		if *verbose {
			fmt.Printf("%s is executed\n", fnName)
		}
		fn(t)
	} else {
		if *verbose {
			fmt.Printf("%s is skipped\n", fnName)
		}
	}
}
