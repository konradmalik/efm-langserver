package core

import (
	"testing"

	"github.com/konradmalik/flint-ls/types"
	"github.com/stretchr/testify/assert"
)

func TestParseLocalFileToURI(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"simple unix", "/home/is/here/uri.go", "file:///home/is/here/uri.go"},
		{"simple windows", "C:/home/is/not/here/uri.go", "file:///C:/home/is/not/here/uri.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLocalFileToURI(tt.path)
			assert.Equal(t, tt.expected, string(got))
		})
	}
}

func TestParsePathToURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      types.DocumentURI
		expected string
		error    bool
	}{
		{"simple unix", "file:///home/is/here/uri.go", "/home/is/here/uri.go", false},
		{"simple windows", "file:///C:/home/is/not/here/uri.go", "C:/home/is/not/here/uri.go", false},
		{"no file", "http://example.com", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PathFromURI(tt.uri)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, got)
		})
	}
}
