//go:build !windows

package core

const (
	shell     = "sh"
	shellFlag = "-c"
)

func comparePaths(path1, path2 string) bool {
	return path1 == path2
}
