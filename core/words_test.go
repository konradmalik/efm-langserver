package core

import (
	"testing"
	"unicode/utf16"

	"github.com/konradmalik/efm-langserver/types"
	"github.com/stretchr/testify/assert"
)

func TestWordAt(t *testing.T) {
	tests := []struct {
		name string
		text string
		// positions here are LSP, so lines and chars are 0 indexed
		pos      types.Position
		expected string
	}{
		{"middle of ascii word", "hello world", types.Position{Line: 0, Character: 1}, "hello"},
		{"middle of second word", "hello world", types.Position{Line: 0, Character: 8}, "world"},
		{"start of word", "hello world", types.Position{Line: 0, Character: 0}, "hello"},
		{"end of word", "hello world", types.Position{Line: 0, Character: 4}, "hello"},
		{"between words", "hello world", types.Position{Line: 0, Character: 5}, " "},
		{"punctuation separated", "foo,bar", types.Position{Line: 0, Character: 0}, "foo"},
		{"punctuation separated second", "foo,bar", types.Position{Line: 0, Character: 4}, "bar"},
		{"underscores kept", "foo_bar baz", types.Position{Line: 0, Character: 5}, "foo_bar"},
		{"hyphen separated", "foo-bar baz", types.Position{Line: 0, Character: 5}, "bar"},
		{"unicode accents", "mañana café", types.Position{Line: 0, Character: 1}, "mañana"},
		{"emoji as word", "hello 😊 world", types.Position{Line: 0, Character: 6}, "😊"},
		{"CJK characters", "你好 世界", types.Position{Line: 0, Character: 0}, "你好"},
		{"position past end", "hello", types.Position{Line: 0, Character: 10}, ""},
		{"empty string", "", types.Position{Line: 0, Character: 0}, ""},
		{"only punctuation", "!!!", types.Position{Line: 0, Character: 1}, "!!!"},
		{"unicode combining marks", "e\u0301clair", types.Position{Line: 0, Character: 0}, "e\u0301clair"},
		{"method call", "someobject.somemethod(arg1,arg2)", types.Position{Line: 0, Character: 12}, "somemethod"},
		{"function def", "func(arg1 string, arg2 int)", types.Position{Line: 0, Character: 1}, "func"},
		{"function def", "func(arg1 string, arg2 int)", types.Position{Line: 0, Character: 7}, "arg1"},

		// multi-line tests
		{"start of second line", "hello\nworld", types.Position{Line: 1, Character: 0}, "world"},
		{"middle of second line", "hello\nworld", types.Position{Line: 1, Character: 3}, "world"},
		{"space on second line", "hello\nworld test", types.Position{Line: 1, Character: 6}, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(utf16.Decode(WordAtUtf16(tt.text, tt.pos)))
			assert.Equal(t, tt.expected, got)
		})
	}
}
