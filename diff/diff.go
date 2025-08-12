package diff

import (
	"github.com/aymanbagabas/go-udiff"
	"github.com/konradmalik/efm-langserver/types"
)

// ComputeEdits computes diff edits from 2 string inputs using go-udiff
func ComputeEdits(name types.DocumentURI, before, after string) ([]types.TextEdit, error) {
	edits := udiff.Strings(before, after)
	d, err := udiff.ToUnifiedDiff(string(name), string(name), before, edits, 0)
	if err != nil {
		return nil, err
	}

	var result []types.TextEdit
	for _, h := range d.Hunks {
		hunkStartLine := h.FromLine - 1
		for il, l := range h.Lines {
			switch l.Kind {
			case udiff.Equal:
				continue
			case udiff.Delete:
				result = append(result, types.TextEdit{Range: types.Range{
					Start: types.Position{Line: hunkStartLine + il, Character: 0},
					End:   types.Position{Line: hunkStartLine + 1 + il, Character: 0},
				}})
			case udiff.Insert:
				result = append(result, types.TextEdit{
					Range: types.Range{
						Start: types.Position{Line: hunkStartLine + il, Character: 0},
						End:   types.Position{Line: hunkStartLine + il, Character: 0},
					},
					NewText: l.Content,
				})
			}
		}
	}

	return result, nil
}
