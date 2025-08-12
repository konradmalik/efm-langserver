package diff

import (
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
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 1, Character: 0},
					},
					NewText: "line1\n",
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "line2\n",
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
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 1, Character: 0},
					},
					NewText: "line1\n",
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "line2\n",
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 3, Character: 0},
						End:   types.Position{Line: 3, Character: 0},
					},
					NewText: "line3\n",
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
						End:   types.Position{Line: 1, Character: 0},
					},
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
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
						End:   types.Position{Line: 1, Character: 0},
					},
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
						End:   types.Position{Line: 3, Character: 0},
					},
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 3, Character: 0},
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
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
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
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
						End:   types.Position{Line: 2, Character: 0},
					},
					NewText: "line3\n",
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 3, Character: 0},
						End:   types.Position{Line: 3, Character: 0},
					},
					NewText: "line4\n",
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
					NewText: "line1\n",
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
						End:   types.Position{Line: 1, Character: 0},
					},
					NewText: "line2\n",
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
						End:   types.Position{Line: 1, Character: 0},
					},
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 1, Character: 0},
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
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
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
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
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
				},
				{
					Range: types.Range{
						Start: types.Position{Line: 2, Character: 0},
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
				panic(err)
			}

			if len(actual) != len(tt.expected) {
				t.Fatalf("[%s] Expected %d edits, got %d", tt.name, len(tt.expected), len(actual))
			}

			for i, edit := range actual {
				expected := tt.expected[i]
				if edit.Range.Start.Line != expected.Range.Start.Line ||
					edit.Range.Start.Character != expected.Range.Start.Character ||
					edit.Range.End.Line != expected.Range.End.Line ||
					edit.Range.End.Character != expected.Range.End.Character {
					t.Errorf("[%s] Edit %d: expected range %+v, got %+v",
						tt.name, i, expected.Range, edit.Range)
				}
				if edit.NewText != expected.NewText {
					t.Errorf("[%s] Edit %d: expected NewText %q, got %q",
						tt.name, i, expected.NewText, edit.NewText)
				}
			}
		})
	}
}

func TestComputeEdits_LargeInput(t *testing.T) {
	// Test with larger input to ensure performance is reasonable
	before := ""
	after := ""

	// Create 1000 lines
	for i := range 1000 {
		if i%2 == 0 {
			before += "line" + string(rune('0'+i%10)) + "\n"
			after += "line" + string(rune('0'+i%10)) + "\n"
		} else {
			// Every odd line is different
			before += "old" + string(rune('0'+i%10)) + "\n"
			after += "new" + string(rune('0'+i%10)) + "\n"
		}
	}

	uri := types.DocumentURI("file:///large.txt")
	edits, err := ComputeEdits(uri, before, after)
	if err != nil {
		panic(err)
	}

	// Should have edits for the changed lines
	if len(edits) == 0 {
		t.Error("Expected some edits for large input with differences")
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
		panic(err)
	}

	// Should have some edits
	if len(edits) == 0 {
		t.Error("Expected edits for complex code change scenario")
	}

	// Verify all edits have valid ranges
	for i, edit := range edits {
		if edit.Range.Start.Line < 0 || edit.Range.End.Line < edit.Range.Start.Line {
			t.Errorf("Edit %d has invalid range: %+v", i, edit.Range)
		}
		if edit.Range.Start.Character != 0 || edit.Range.End.Character != 0 {
			t.Errorf("Edit %d: expected Character to be 0, got Start: %d, End: %d",
				i, edit.Range.Start.Character, edit.Range.End.Character)
		}
	}
}
