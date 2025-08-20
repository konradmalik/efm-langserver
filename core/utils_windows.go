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
