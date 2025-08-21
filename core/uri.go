package core

import (
	"fmt"
	"net/url"
	"path/filepath"
	"unicode"

	"github.com/konradmalik/efm-langserver/types"
)

const fileScheme = "file"

func PathFromURI(uri types.DocumentURI) (string, error) {
	if uri == "" {
		return "", nil
	}

	u, err := url.ParseRequestURI(string(uri))
	if err != nil {
		return "", err
	}
	if u.Scheme != fileScheme {
		return "", fmt.Errorf("only %s URIs are supported, got %v", fileScheme, u.Scheme)
	}
	if isWindowsDriveURIPath(u.Path) {
		// skip initial slash
		u.Path = u.Path[1:]
	}
	return u.Path, nil
}

func ParseLocalFileToURI(path string) types.DocumentURI {
	if path == "" {
		return ""
	}

	if isWindowsDrivePath(path) {
		// add initial slash
		path = "/" + path
	}
	return types.DocumentURI((&url.URL{
		Scheme: fileScheme,
		Path:   filepath.ToSlash(path),
	}).String())
}

func isWindowsDrivePath(path string) bool {
	if len(path) < 3 {
		return false
	}
	return unicode.IsLetter(rune(path[0])) && path[1] == ':'
}

func isWindowsDriveURIPath(uri string) bool {
	if len(uri) < 3 {
		return false
	}
	return uri[0] == '/' && unicode.IsLetter(rune(uri[1])) && uri[2] == ':'
}
