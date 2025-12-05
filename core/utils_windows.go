//go:build windows

package core

import (
	"strings"
)

const (
	shell     = "cmd"
	shellFlag = "/c"
)

func comparePaths(path1, path2 string) bool {
	return strings.EqualFold(path1, path2)
}

// used to check if returned error is related to linting analysis or system error
// system error is e.g. binary not found or not executable
func isSystemError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "cannot find the path") || strings.Contains(msg, "is not recognized")
}
