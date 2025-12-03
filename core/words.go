package core

import (
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/konradmalik/flint-ls/types"
)

const (
	invalid = iota - 1
	blank
	punctuation
	word
)

// lsp can now select encoding from the list that clients send that they support,
// but utf16 is selected if the server does not send it and is required for backwards compatibility,
// so we just support utf16
func WordAtUtf16(text string, pos types.Position) []uint16 {
	lines := strings.Split(text, "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return nil
	}
	chars := utf16.Encode([]rune(lines[pos.Line]))
	if pos.Character < 0 || pos.Character > len(chars) {
		return nil
	}

	prevPos := 0
	currPos := -1
	prevCls := invalid
	for i, char := range chars {
		currCls := getRuneClass(rune(char))
		if currCls != prevCls {
			if i <= pos.Character {
				prevPos = i
			} else {
				currPos = i
				break
			}
		}
		prevCls = currCls
	}
	if currPos == -1 {
		currPos = len(chars)
	}
	return chars[prevPos:currPos]
}

func getRuneClass(r rune) int {
	if r >= 0x100 {
		return word
	}
	if unicode.IsSpace(r) || unicode.IsOneOf([]*unicode.RangeTable{unicode.Z, unicode.Pattern_White_Space}, r) {
		return blank
	}
	if r == '_' || !unicode.IsPunct(r) {
		return word
	}
	return punctuation
}
