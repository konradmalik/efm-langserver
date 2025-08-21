package core

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
	assert.Error(t, err)
}

func TestLintFileMatched(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(file)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)

	d := uriToDiag[uri]
	assert.Len(t, d, 1)
	assert.Equal(t, d[0].Range.Start.Line, 1)
	assert.Equal(t, d[0].Range.Start.Character, 0)
	assert.Equal(t, d[0].Severity, types.DiagnosticSeverity(1))
	assert.Equal(t, d[0].Message, "No it is normal!")
}

func TestLintFileMatchedWildcard(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(file)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)

	d := uriToDiag[uri]
	assert.Len(t, d, 1)
	assert.Equal(t, d[0].Range.Start.Line, 1)
	assert.Equal(t, d[0].Range.Start.Character, 0)
	assert.Equal(t, d[0].Severity, types.DiagnosticSeverity(1))
	assert.Equal(t, d[0].Message, "No it is normal!")
}

// column 0 remains unchanged, regardless of the configured offset
// column 0 indicates a whole line (although for 0-based column linters we can not distinguish between word starting at 0 and the whole line)
func TestLintOffsetColumnsZero(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(file)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)

	d := uriToDiag[uri]
	assert.Len(t, d, 1)
	assert.Equal(t, d[0].Range.Start.Character, 0)
}

// without column offset, 1-based columns are assumed, which means that we should get 0 for column 1
// as LSP assumes 0-based columns
func TestLintOffsetColumnsNoOffset(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(file)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)

	d := uriToDiag[uri]
	assert.Len(t, d, 1)
	assert.Equal(t, d[0].Range.Start.Character, 0)
}

// for column 1 with offset we should get column 1 back
// without the offset efm would subtract 1 as it expects 1 based columns
func TestLintOffsetColumnsNonZero(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(file)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)

	d := uriToDiag[uri]
	assert.Len(t, d, 1)
	assert.Equal(t, d[0].Range.Start.Character, 1)
}

func TestLintCategoryMap(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(file)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)

	d := uriToDiag[uri]
	assert.Len(t, d, 1)
	assert.Equal(t, d[0].Severity, types.DiagnosticSeverity(3))
}

// Test if lint is executed if required root markers for the language are missing
func TestLintRequireRootMarker(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(file)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)

	assert.Empty(t, uriToDiag)
}

func TestLintSingleEntry(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	file2 := filepath.Join(base, "bar")
	uri := ParseLocalFileToURI(file)
	uri2 := ParseLocalFileToURI(file2)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
			uri2: {
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file2,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)
	assert.Len(t, uriToDiag, 1)

	d := uriToDiag[uri]
	assert.Len(t, d, 1)
	assert.Equal(t, d[0].Range.Start.Line, 1)
	assert.Equal(t, d[0].Range.Start.Character, 0)
}

func TestLintMultipleEntries(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	file2 := filepath.Join(base, "bar")
	uri := ParseLocalFileToURI(file)
	uri2 := ParseLocalFileToURI(file2)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
			uri2: {
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file2,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri2, types.EventTypeChange)
	assert.NoError(t, err)
	assert.Len(t, uriToDiag, 1)

	d := uriToDiag[uri2]
	assert.Len(t, d, 2)
	assert.Equal(t, d[0].Range.Start.Line, 1)
	assert.Equal(t, d[0].Range.Start.Character, 2)
	assert.Equal(t, d[1].Range.Start.Line, 0)
	assert.Equal(t, d[1].Range.Start.Character, 0)
}

func TestLintNoDiagnostics(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(file)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
			},
		},
	}

	uriToDiag, err := h.lintDocument(context.Background(), nil, uri, types.EventTypeChange)
	assert.NoError(t, err)
	assert.Len(t, uriToDiag, 1)

	d := uriToDiag[uri]
	assert.Empty(t, d)
}

func TestLintEventTypes(t *testing.T) {
	base, _ := os.Getwd()
	file := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(file)

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
				LanguageID:         "vim",
				Text:               "scriptencoding utf-8\nabnormal!\n",
				NormalizedFilename: file,
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
			assert.NoError(t, err)

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
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsEntryForRequestedURI(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		uri      string
		entry    *errorformat.Entry
		expected bool
	}{
		{
			name: "main dir",
			root: "/home/torvalds/linux/",
			uri:  "file:///home/torvalds/linux/main.go",
			entry: &errorformat.Entry{
				Filename: "main.go",
			},
			expected: true,
		},
		{
			name: "subdir without slash",
			root: "/home/torvalds/linux/",
			uri:  "file:///home/torvalds/linux/gpu/nvidia/driver.go",
			entry: &errorformat.Entry{
				Filename: "gpu/nvidia/driver.go",
			},
			expected: true,
		},
		{
			name: "subdir with slash is absolute",
			root: "/home/torvalds/linux/",
			uri:  "file:///home/torvalds/linux/gpu/nvidia/driver.go",
			entry: &errorformat.Entry{
				Filename: "/gpu/nvidia/driver.go",
			},
			expected: runtime.GOOS == "windows",
		},
		{
			name: "empty filename is accepted",
			root: "/home/torvalds/linux/",
			uri:  "file:///home/torvalds/linux/gpu/nvidia/driver.go",
			entry: &errorformat.Entry{
				Filename: "",
			},
			expected: true,
		},
		{
			name: "comparison is case sensitive",
			root: "/home/torvalds/linux/",
			uri:  "file:///home/torvalds/linux/main.go",
			entry: &errorformat.Entry{
				Filename: "Main.go",
			},
			expected: runtime.GOOS == "windows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok := isEntryForRequestedURI(tt.root, types.DocumentURI(tt.uri), tt.entry)
			assert.Equal(t, tt.expected, ok)
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
		{
			name: "multiline is handled",
			entry: &errorformat.Entry{
				Lnum:    1,
				EndLnum: 3,
				Col:     0,
				Text:    "bad",
				Type:    'E',
			},
			cfg: &types.Language{
				LintOffset:        -2,
				LintOffsetColumns: 0,
			},
			expected: types.Diagnostic{
				Message:  "bad",
				Severity: types.Error,
				Range: types.Range{
					Start: types.Position{Line: 2, Character: 0},
					End:   types.Position{Line: 4, Character: 0},
				},
			},
		},
		{
			name: "multicol is handled",
			entry: &errorformat.Entry{
				Lnum:    2,
				EndLnum: 2,
				Col:     3,
				EndCol:  7,
				Text:    "bad",
				Type:    'E',
			},
			cfg: &types.Language{
				LintOffset:        0,
				LintOffsetColumns: 2,
			},
			expected: types.Diagnostic{
				Message:  "bad",
				Severity: types.Error,
				Range: types.Range{
					Start: types.Position{Line: 1, Character: 4},
					End:   types.Position{Line: 1, Character: 8},
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
