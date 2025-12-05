//go:build !windows

package core

import "strings"

const (
	shell     = "sh"
	shellFlag = "-c"
)

func comparePaths(path1, path2 string) bool {
	return path1 == path2
}

// used to check if returned error is related to linting analysis or system error
// system error is e.g. binary not found or not executable
func isSystemError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "127") || strings.Contains(msg, "126")
}
