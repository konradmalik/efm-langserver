package core

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konradmalik/efm-langserver/types"
	"github.com/reviewdog/errorformat"
	"github.com/stretchr/testify/assert"
)

func TestLintNoLinter(t *testing.T) {
	h := &LangHandler{
		logger:  log.New(log.Writer(), "", log.LstdFlags),
		configs: map[string][]types.Language{},
		files: map[types.DocumentURI]*fileRef{
			types.DocumentURI("file:///foo"): {},
		},
	}

	_, err := h.lintDocument(context.Background(), nil, "file:///foo", types.EventTypeChange)
	assert.NoError(t, err)
}

func TestLintNoFileMatched(t *testing.T) {
	h := &LangHandler{
		logger:  log.New(log.Writer(), "", log.LstdFlags),
		configs: map[string][]types.Language{},
		files: map[types.DocumentURI]*fileRef{
			types.DocumentURI("file:///foo"): {},
		},
	}

	_, err := h.lintDocument(context.Background(), nil, "file:///bar", types.EventTypeChange)
	assert.NoError(t, err)
}

func TestLintFileMatched(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := toURI(file)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			"vim": {
				{
					LintCommand:        `echo ` + file + `:2:No it is normal!`,
					LintIgnoreExitCode: true,
					LintStdin:          true,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)

	d := uriToDiag[uri]
	if len(d) != 1 {
		t.Fatal("diagnostics should be only one", d)
	}
	if d[0].Range.Start.Line != 1 {
		t.Fatalf("range.start.line should be %v but got: %v", 1, d[0].Range.Start.Line)
	}
	if d[0].Range.Start.Character != 0 {
		t.Fatalf("range.start.character should be %v but got: %v", 0, d[0].Range.Start.Character)
	}
	if d[0].Severity != 1 {
		t.Fatalf("severity should be %v but got: %v", 0, d[0].Severity)
	}
	if strings.TrimSpace(d[0].Message) != "No it is normal!" {
		t.Fatalf("message should be %q but got: %q", "No it is normal!", strings.TrimSpace(d[0].Message))
	}
}

func TestLintFileMatchedWildcard(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := toURI(file)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			types.Wildcard: {
				{
					LintCommand:        `echo ` + file + `:2:No it is normal!`,
					LintIgnoreExitCode: true,
					LintStdin:          true,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)

	d := uriToDiag[uri]
	if len(d) != 1 {
		t.Fatal("diagnostics should be only one")
	}
	if d[0].Range.Start.Line != 1 {
		t.Fatalf("range.start.line should be %v but got: %v", 1, d[0].Range.Start.Line)
	}
	if d[0].Range.Start.Character != 0 {
		t.Fatalf("range.start.character should be %v but got: %v", 0, d[0].Range.Start.Character)
	}
	if d[0].Severity != 1 {
		t.Fatalf("severity should be %v but got: %v", 0, d[0].Severity)
	}
	if strings.TrimSpace(d[0].Message) != "No it is normal!" {
		t.Fatalf("message should be %q but got: %q", "No it is normal!", strings.TrimSpace(d[0].Message))
	}
}

// column 0 remains unchanged, regardless of the configured offset
// column 0 indicates a whole line (although for 0-based column linters we can not distinguish between word starting at 0 and the whole line)
func TestLintOffsetColumnsZero(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := toURI(file)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			types.Wildcard: {
				{
					LintCommand:        `echo ` + file + `:2:0:msg`,
					LintFormats:        []string{"%f:%l:%c:%m"},
					LintIgnoreExitCode: true,
					LintStdin:          true,
					LintOffsetColumns:  1,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	d := uriToDiag[uri]
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 1 {
		t.Fatal("diagnostics should be only one")
	}
	if d[0].Range.Start.Character != 0 {
		t.Fatalf("range.start.character should be %v but got: %v", 0, d[0].Range.Start.Character)
	}
}

// without column offset, 1-based columns are assumed, which means that we should get 0 for column 1
// as LSP assumes 0-based columns
func TestLintOffsetColumnsNoOffset(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := toURI(file)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			types.Wildcard: {
				{
					LintCommand:        `echo ` + file + `:2:1:msg`,
					LintFormats:        []string{"%f:%l:%c:%m"},
					LintIgnoreExitCode: true,
					LintStdin:          true,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	d := uriToDiag[uri]
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 1 {
		t.Fatal("diagnostics should be only one")
	}
	if d[0].Range.Start.Character != 0 {
		t.Fatalf("range.start.character should be %v but got: %v", 0, d[0].Range.Start.Character)
	}
}

// for column 1 with offset we should get column 1 back
// without the offset efm would subtract 1 as it expects 1 based columns
func TestLintOffsetColumnsNonZero(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := toURI(file)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			types.Wildcard: {
				{
					LintCommand:        `echo ` + file + `:2:1:msg`,
					LintFormats:        []string{"%f:%l:%c:%m"},
					LintIgnoreExitCode: true,
					LintStdin:          true,
					LintOffsetColumns:  1,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	d := uriToDiag[uri]
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 1 {
		t.Fatal("diagnostics should be only one")
	}
	if d[0].Range.Start.Character != 1 {
		t.Fatalf("range.start.character should be %v but got: %v", 1, d[0].Range.Start.Character)
	}
}

func TestLintCategoryMap(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := toURI(file)

	mapping := make(map[string]string)
	mapping["R"] = "I" // pylint refactoring to info

	formats := []string{"%f:%l:%c:%t:%m"}

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			types.Wildcard: {
				{
					LintCommand:        `echo ` + file + `:2:1:R:No it is normal!`,
					LintIgnoreExitCode: true,
					LintStdin:          true,
					LintFormats:        formats,
					LintCategoryMap:    mapping,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	d := uriToDiag[uri]
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 1 {
		t.Fatal("diagnostics should be only one")
	}
	if d[0].Severity != 3 {
		t.Fatalf("Severity should be %v but is: %v", 3, d[0].Severity)
	}
}

// Test if lint is executed if required root markers for the language are missing
func TestLintRequireRootMarker(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := toURI(file)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			"vim": {
				{
					LintCommand:        `echo ` + file + `:2:No it is normal!`,
					LintIgnoreExitCode: true,
					LintStdin:          true,
					RequireMarker:      true,
					RootMarkers:        []string{".vimlintrc"},
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	d, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 0 {
		t.Fatal("diagnostics should be zero as we have no root marker for the language but require one", d)
	}
}

func TestLintSingleEntry(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	file2 := filepath.Join(base, "bar")
	uri := toURI(file)
	uri2 := toURI(file2)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			"vim": {
				{
					LintCommand:        `echo ` + file + `:2:1:First file! && echo ` + file2 + `:1:2:Second file!`,
					LintFormats:        []string{"%f:%l:%c:%m"},
					LintIgnoreExitCode: true,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
			uri2: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	d, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 1 {
		t.Fatalf("diagnostics should be one, but got %#v", d)
	}
	if d[uri][0].Range.Start.Character != 0 {
		t.Fatalf("first range.start.character should be %v but got: %v", 0, d[uri][0].Range.Start.Character)
	}
	if d[uri][0].Range.Start.Line != 1 {
		t.Fatalf("first range.start.line should be %v but got: %v", 1, d[uri][0].Range.Start.Line)
	}
}

func TestLintMultipleEntries(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	file2 := filepath.Join(base, "bar")
	uri := toURI(file)
	uri2 := toURI(file2)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			"vim": {
				{
					LintCommand:        `echo ` + file + `:2:1:First file! && echo ` + file2 + `:2:3:Second file! && echo ` + file2 + `:Empty l and c!`,
					LintFormats:        []string{"%f:%l:%c:%m", "%f:%m"},
					LintIgnoreExitCode: true,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
			uri2: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	d, err := h.lintDocument(context.Background(), nil, uri2, types.EventTypeChange)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 1 {
		t.Fatalf("diagnostics should be for one file, but got %#v", d)
	}
	if len(d[uri2]) != 2 {
		t.Fatalf("should have two diagnostics, but got %#v", d[uri2])
	}
	assert.Equal(t, 2, d[uri2][0].Range.Start.Character)
	assert.Equal(t, 1, d[uri2][0].Range.Start.Line)
	assert.Equal(t, 0, d[uri2][1].Range.Start.Character)
	assert.Equal(t, 0, d[uri2][1].Range.Start.Line)
}

func TestLintNoDiagnostics(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := toURI(file)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			"vim": {
				{
					LintCommand:        "echo ",
					LintIgnoreExitCode: true,
					LintStdin:          true,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	if err != nil {
		t.Fatal(err)
	}
	d, ok := uriToDiag[uri]
	if !ok {
		t.Fatal("didn't get any diagnostics")
	}
	if len(d) != 0 {
		t.Fatal("diagnostics should be an empty list", d)
	}
}

func TestLintEventTypes(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := toURI(file)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			"vim": {
				{
					LintCommand:        `echo ` + file + `:2:No it is normal!`,
					LintIgnoreExitCode: true,
					LintStdin:          true,
				},
			},
		},
		files: map[types.DocumentURI]*fileRef{
			uri: {
				LanguageID: "vim",
				Text:       "scriptencoding utf-8\nabnormal!\n",
			},
		},
	}

	tests := []struct {
		name           string
		event          types.EventType
		lintAfterOpen  bool
		lintOnSave     bool
		lintOnChange   bool
		expectMessages int
	}{
		{
			name:           "LintOnOpen true",
			event:          types.EventTypeOpen,
			lintAfterOpen:  true,
			expectMessages: 1,
		},
		{
			name:           "LintOnOpen false",
			event:          types.EventTypeOpen,
			lintAfterOpen:  false,
			expectMessages: 0,
		},
		{
			name:           "LintOnChange true",
			event:          types.EventTypeChange,
			lintOnChange:   true,
			expectMessages: 1,
		},
		{
			name:           "LintOnChange false",
			event:          types.EventTypeChange,
			lintOnChange:   false,
			expectMessages: 0,
		},
		{
			name:           "LintOnSave true",
			event:          types.EventTypeSave,
			lintOnSave:     true,
			expectMessages: 1,
		},
		{
			name:           "LintOnSave false",
			event:          types.EventTypeSave,
			lintOnSave:     false,
			expectMessages: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.configs["vim"][0].LintAfterOpen = boolPtr(tt.lintAfterOpen)
			h.configs["vim"][0].LintOnChange = boolPtr(tt.lintOnChange)
			h.configs["vim"][0].LintOnSave = boolPtr(tt.lintOnSave)
			uriToDiag, err := h.lintDocument(context.Background(), nil, uri, tt.event)
			if err != nil {
				t.Fatal(err)
			}

			d := uriToDiag[uri]
			assert.Equal(t, tt.expectMessages, len(d))
		})
	}
}

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name            string
		typ             rune
		categoryMap     map[string]string
		defaultSeverity types.DiagnosticSeverity
		want            types.DiagnosticSeverity
	}{
		{"Error type", 'E', nil, 0, types.Error},
		{"Warning type", 'W', nil, 0, types.Warning},
		{"Info type", 'I', nil, 0, types.Information},
		{"Hint type", 'N', nil, 0, types.Hint},
		{"Default severity overrides", 'X', nil, types.Warning, types.Warning},
		{"Category map remap", 'X', map[string]string{"X": "W"}, 0, types.Warning},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSeverity(tt.typ, tt.categoryMap, tt.defaultSeverity)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEfmEntryToDiagnostic(t *testing.T) {
	file := &fileRef{Text: "hello world\ngolang rulezz", LanguageID: "txt"}
	tests := []struct {
		name     string
		entry    *errorformat.Entry
		cfg      *types.Language
		expected types.Diagnostic
	}{
		{
			name: "first line as 1, word",
			entry: &errorformat.Entry{
				Lnum: 1,
				Col:  7,
				Text: "world bad",
				Type: 'E',
			},
			cfg: &types.Language{
				LintOffset:        0,
				LintOffsetColumns: 0,
			},
			expected: types.Diagnostic{
				Message:  "world bad",
				Severity: types.Error,
				Range: types.Range{
					Start: types.Position{Line: 0, Character: 6},
					End:   types.Position{Line: 0, Character: 11},
				},
			},
		},
		{
			name: "first line as 0, word",
			entry: &errorformat.Entry{
				Lnum: 0,
				Col:  7,
				Text: "world bad",
				Type: 'E',
			},
			cfg: &types.Language{
				LintOffset:        0,
				LintOffsetColumns: 0,
			},
			expected: types.Diagnostic{
				Message:  "world bad",
				Severity: types.Error,
				Range: types.Range{
					Start: types.Position{Line: 0, Character: 6},
					End:   types.Position{Line: 0, Character: 11},
				},
			},
		},
		{
			name: "second line, word",
			entry: &errorformat.Entry{
				Lnum: 2,
				Col:  1,
				Text: "golang bad",
				Type: 'E',
			},
			cfg: &types.Language{
				LintOffset:        0,
				LintOffsetColumns: 0,
			},
			expected: types.Diagnostic{
				Message:  "golang bad",
				Severity: types.Error,
				Range: types.Range{
					Start: types.Position{Line: 1, Character: 0},
					End:   types.Position{Line: 1, Character: 6},
				},
			},
		},
		{
			name: "second line, whole",
			entry: &errorformat.Entry{
				Lnum: 2,
				Col:  0,
				Text: "golang not rulezz",
				Type: 'E',
			},
			cfg: &types.Language{
				LintOffset:        0,
				LintOffsetColumns: 0,
			},
			expected: types.Diagnostic{
				Message:  "golang not rulezz",
				Severity: types.Error,
				Range: types.Range{
					Start: types.Position{Line: 1, Character: 0},
					End:   types.Position{Line: 1, Character: 0},
				},
			},
		},
		{
			name: "line offset is subtracted",
			entry: &errorformat.Entry{
				Lnum: 1,
				Col:  7,
				Text: "world bad",
				Type: 'E',
			},
			cfg: &types.Language{
				LintOffset:        -1,
				LintOffsetColumns: 0,
			},
			expected: types.Diagnostic{
				Message:  "world bad",
				Severity: types.Error,
				Range: types.Range{
					Start: types.Position{Line: 1, Character: 6},
					End:   types.Position{Line: 1, Character: 7},
				},
			},
		},
		{
			name: "col offset is added",
			entry: &errorformat.Entry{
				Lnum: 1,
				Col:  7,
				Text: "world bad",
				Type: 'E',
			},
			cfg: &types.Language{
				LintOffset:        0,
				LintOffsetColumns: 1,
			},
			expected: types.Diagnostic{
				Message:  "world bad",
				Severity: types.Error,
				Range: types.Range{
					Start: types.Position{Line: 0, Character: 7},
					End:   types.Position{Line: 0, Character: 12},
				},
			},
		},
		{
			name: "col offset is not added if whole line",
			entry: &errorformat.Entry{
				Lnum: 1,
				Col:  0,
				Text: "world bad",
				Type: 'E',
			},
			cfg: &types.Language{
				LintOffset:        0,
				LintOffsetColumns: 11,
			},
			expected: types.Diagnostic{
				Message:  "world bad",
				Severity: types.Error,
				Range: types.Range{
					Start: types.Position{Line: 0, Character: 0},
					End:   types.Position{Line: 0, Character: 0},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := parseEfmEntryToDiagnostic(tt.entry, tt.cfg, file)
			assert.Equal(t, tt.expected.Message, diag.Message)
			assert.Equal(t, tt.expected.Severity, diag.Severity)
			assert.Equal(t, tt.expected.Range.Start.Line, diag.Range.Start.Line)
			assert.Equal(t, tt.expected.Range.Start.Character, diag.Range.Start.Character)
			assert.Equal(t, tt.expected.Range.End.Line, diag.Range.End.Line)
			assert.Equal(t, tt.expected.Range.End.Character, diag.Range.End.Character)
		})
	}
}
