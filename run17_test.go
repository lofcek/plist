// +build go1.7

package plist

import "testing"

func runTest(name string, fn func(t *testing.T), t *testing.T) {
	t.Run(name, fn)
}
