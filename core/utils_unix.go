//go:build !windows

package core

const (
	shell               = "sh"
	shellFlag           = "-c"
	commandNotFoundCode = 127
)

func comparePaths(path1, path2 string) bool {
	return path1 == path2
}
