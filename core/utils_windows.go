//go:build windows

package core

import (
	"strings"
)

const (
	shell               = "cmd"
	shellFlag           = "/c"
	commandNotFoundCode = 9009
)

func comparePaths(path1, path2 string) bool {
	return strings.EqualFold(path1, path2)
}
