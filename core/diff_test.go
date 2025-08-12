package core

import (
	"strings"
	"testing"

	"github.com/konradmalik/efm-langserver/types"
)

func TestComputeEdits(t *testing.T) {
	tests := []struct {
		name     string
		before   string
		after    string
		expected []types.TextEdit
	}{
		{
			name:     "no changes",
			before:   "line1\nline2\nline3\n",
			after:    "line1\nline2\nline3\n",
			expected: []types.TextEdit{},
		},
		{
			name:   "single line insertion at beginning",
			before: "line2\nline3\n",
			after:  "line1\nline2\nline3\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 0, Character: 0},
						End:   types.Position{Line: 0, Character: 0},
					},
					NewText: "line1\n",
				},
			},
		},
		{
			name:   "single line insertion at end",
			before: "line1\nline2\n",
			after:  "line1\nline2\nline3\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "line3\n",
				},
			},
		},
		{
			name:   "single line insertion in middle",
			before: "line1\nline3\n",
			after:  "line1\nline2\nline3\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 0, Character: 0},
						End:   types.Position{Line: 1, Character: 0},
					},
					NewText: "line1\nline2\n",
				},
			},
		},
		{
			name:   "multiple line insertion",
			before: "line1\nline4\n",
			after:  "line1\nline2\nline3\nline4\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 0, Character: 0},
						End:   types.Position{Line: 1, Character: 0},
					},
					NewText: "line1\nline2\nline3\n",
				},
			},
		},
		{
			name:   "single line deletion at beginning",
			before: "line1\nline2\nline3\n",
			after:  "line2\nline3\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 0, Character: 0},
						End:   types.Position{Line: 1, Character: 0},
					},
				},
			},
		},
		{
			name:   "single line deletion at end",
			before: "line1\nline2\nline3\n",
			after:  "line1\nline2\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
						End:   types.Position{Line: 3, Character: 0},
					},
				},
			},
		},
		{
			name:   "single line deletion in middle",
			before: "line1\nline2\nline3\n",
			after:  "line1\nline3\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 0, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "line1\n",
				},
			},
		},
		{
			name:   "multiple line deletion",
			before: "line1\nline2\nline3\nline4\n",
			after:  "line1\nline4\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 0, Character: 0},
						End:   types.Position{Line: 3, Character: 0},
					},
					NewText: "line1\n",
				},
			},
		},
		{
			name:   "line replacement",
			before: "line1\nold_line\nline3\n",
			after:  "line1\nnew_line\nline3\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "new_line\n",
				},
			},
		},
		{
			name:   "multiple changes",
			before: "line1\nline2\nline5\n",
			after:  "line1\nline3\nline4\nline5\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "line3\nline4\n",
				},
			},
		},
		{
			name:     "empty to empty",
			before:   "",
			after:    "",
			expected: []types.TextEdit{},
		},
		{
			name:   "empty to content",
			before: "",
			after:  "line1\nline2\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 0, Character: 0},
						End:   types.Position{Line: 0, Character: 0},
					},
					NewText: "line1\nline2\n",
				},
			},
		},
		{
			name:   "content to empty",
			before: "line1\nline2\n",
			after:  "",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 0, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
				},
			},
		},
		{
			name:   "no trailing newline in before",
			before: "line1\nline2",
			after:  "line1\nline3",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "line3",
				},
			},
		},
		{
			name:   "no trailing newline in after",
			before: "line1\nline2\n",
			after:  "line1\nline3",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "line3",
				},
			},
		},
		{
			name:   "single character line",
			before: "a\nb\nc\n",
			after:  "a\nx\nc\n",
			expected: []types.TextEdit{
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "x\n",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := types.DocumentURI("file:///test.txt")
			actual, err := ComputeEdits(uri, tt.before, tt.after)
			if err != nil {
				t.Fatalf("[%s] unexpected error: %v", tt.name, err)
			}

			// Validate expected exact match if provided
			if tt.expected != nil {
				if len(actual) != len(tt.expected) {
					t.Fatalf("[%s] Expected %d edits, got %d", tt.name, len(tt.expected), len(actual))
				}
				for i, edit := range actual {
					expected := tt.expected[i]
					if edit.Range != expected.Range {
						t.Errorf("[%s] Edit %d: expected range %+v, got %+v", tt.name, i, expected.Range, edit.Range)
					}
					if edit.NewText != expected.NewText {
						t.Errorf("[%s] Edit %d: expected NewText %q, got %q", tt.name, i, expected.NewText, edit.NewText)
					}
				}
			}

			// Validate correctness by applying edits
			afterApplied := applyEdits(tt.before, actual)
			if afterApplied != tt.after {
				t.Errorf("[%s] Applying edits did not yield expected text.\nExpected:\n%q\nGot:\n%q",
					tt.name, tt.after, afterApplied)
			}

			// Validate that edits are sorted and non-overlapping
			for i := 1; i < len(actual); i++ {
				prev := actual[i-1]
				curr := actual[i]
				if curr.Range.Start.Line < prev.Range.End.Line ||
					(curr.Range.Start.Line == prev.Range.End.Line && curr.Range.Start.Character < prev.Range.End.Character) {
					t.Errorf("[%s] Edits are overlapping or out of order: edit %d and %d", tt.name, i-1, i)
				}
			}

			// Validate that ranges are valid
			for i, e := range actual {
				if e.Range.Start.Line < 0 || e.Range.End.Line < e.Range.Start.Line {
					t.Errorf("[%s] Edit %d has invalid line range: %+v", tt.name, i, e.Range)
				}
				if e.Range.Start.Character < 0 || e.Range.End.Character < 0 {
					t.Errorf("[%s] Edit %d has negative character position: %+v", tt.name, i, e.Range)
				}
			}
		})
	}
}

func TestComputeEdits_LargeInput(t *testing.T) {
	before := ""
	after := ""

	// Create 1000 lines
	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			before += "line" + string(rune('0'+i%10)) + "\n"
			after += "line" + string(rune('0'+i%10)) + "\n"
		} else {
			before += "old" + string(rune('0'+i%10)) + "\n"
			after += "new" + string(rune('0'+i%10)) + "\n"
		}
	}

	uri := types.DocumentURI("file:///large.txt")
	edits, err := ComputeEdits(uri, before, after)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Apply edits and compare
	afterApplied := applyEdits(before, edits)
	if afterApplied != after {
		t.Errorf("Large input edits did not yield expected result")
	}
}

func TestComputeEdits_ComplexScenario(t *testing.T) {
	before := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
    x := 42
    fmt.Println(x)
}
`

	after := `package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Hello, Go!")
    y := 100
    fmt.Println(y)
    os.Exit(0)
}
`

	uri := types.DocumentURI("file:///main.go")
	edits, err := ComputeEdits(uri, before, after)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Apply and check correctness
	afterApplied := applyEdits(before, edits)
	if afterApplied != after {
		t.Errorf("Complex scenario edits did not yield expected result.\nExpected:\n%s\nGot:\n%s", after, afterApplied)
	}
}

// applyEdits applies LSP-style text edits to the given text.
func applyEdits(text string, edits []types.TextEdit) string {
	lines := strings.SplitAfter(text, "\n")
	var result strings.Builder
	lastLine := 0

	for _, e := range edits {
		// Write unchanged part
		for i := lastLine; i < e.Range.Start.Line; i++ {
			if i < len(lines) {
				result.WriteString(lines[i])
			}
		}

		// Write replacement text
		result.WriteString(e.NewText)

		lastLine = e.Range.End.Line
	}

	// Append remaining lines
	for i := lastLine; i < len(lines); i++ {
		result.WriteString(lines[i])
	}

	return result.String()
}
