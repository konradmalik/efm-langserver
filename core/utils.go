package core

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/konradmalik/efm-langserver/types"
)

const (
	flagPlaceholder  = "$1"
	windowsShell     = "cmd"
	windowsShellArg  = "/c"
	unixShell        = "sh"
	unixShellArg     = "-c"
	inputPlaceholder = "${INPUT}"
	newlineChar      = "\r"
)

func normalizeFilename(uri types.DocumentURI) (string, error) {
	fname, err := fromURI(uri)
	if err != nil {
		return "", fmt.Errorf("invalid uri: %v: %v", err, uri)
	}
	fname = filepath.ToSlash(fname)
	if runtime.GOOS == "windows" {
		fname = strings.ToLower(fname)
	}
	return fname, nil
}

func itoaPtrIfNotZero(n int) *string {
	if n == 0 {
		return nil
	}
	s := strconv.Itoa(n)
	return &s
}
