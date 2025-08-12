package diff

import (
	"github.com/aymanbagabas/go-udiff"
	"github.com/konradmalik/efm-langserver/types"
)

func ComputeEdits(name types.DocumentURI, before, after string) ([]types.TextEdit, error) {
	edits := udiff.Strings(before, after)
	d, err := udiff.ToUnifiedDiff(string(name), string(name), before, edits, 0)
	if err != nil {
		return nil, err
	}

	var result []types.TextEdit
	for _, h := range d.Hunks {
		startLine := h.FromLine - 1
		endLine := startLine
		var newText string

		for _, l := range h.Lines {
			switch l.Kind {
			case udiff.Equal:
				endLine++
			case udiff.Delete:
				endLine++
			case udiff.Insert:
				newText += l.Content
			}
		}

		result = append(result, types.TextEdit{
			Range: types.Range{
				Start: types.Position{Line: startLine, Character: 0},
				End:   types.Position{Line: endLine, Character: 0},
			},
			NewText: newText,
		})
	}

	return result, nil
}
