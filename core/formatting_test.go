package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konradmalik/efm-langserver/types"
	"github.com/stretchr/testify/assert"
)

func TestNormalizedFilenameFromURI(t *testing.T) {
	uri := types.DocumentURI("file:///tmp/TestFile.txt")
	fname, err := normalizedFilenameFromUri(uri)
	assert.NoError(t, err)
	assert.Equal(t, "/tmp/TestFile.txt", fname)
}

func TestApplyOptionsPlaceholders_DefaultTypes(t *testing.T) {
	cmd := "echo ${--flag:opt} ${--flag2=opt}"
	opts := types.FormattingOptions{
		"opt": "value",
	}
	out, err := applyOptionsPlaceholders(cmd, opts)
	assert.NoError(t, err)
	assert.Contains(t, out, "--flag value")
	assert.Contains(t, out, "--flag2=value")
}

func TestApplyOptionsPlaceholders_BoolTrue(t *testing.T) {
	cmd := "echo ${--flag:opt} ${--flag2=opt}"
	opts := types.FormattingOptions{
		"opt": true,
	}
	out, err := applyOptionsPlaceholders(cmd, opts)
	assert.NoError(t, err)
	assert.Equal(t, "echo --flag --flag2", out)
}

func TestApplyOptionsPlaceholders_BoolFalse(t *testing.T) {
	cmd := "echo ${--flag:!opt} ${--flag2=!opt}"
	opts := types.FormattingOptions{
		"opt": false,
	}
	out, err := applyOptionsPlaceholders(cmd, opts)
	assert.NoError(t, err)
	assert.Equal(t, "echo --flag --flag2", out)
}

func TestApplyRangePlaceholders(t *testing.T) {
	cmd := "echo ${--flag:charStart} ${--flag=charEnd}"
	rng := &types.Range{
		Start: types.Position{Line: 0, Character: 2},
		End:   types.Position{Line: 0, Character: 4},
	}
	text := "abcdef"
	out, err := applyRangePlaceholders(cmd, rng, text)
	assert.NoError(t, err)
	assert.Contains(t, out, "--flag 2")
	assert.Contains(t, out, "--flag=4")
}

func TestBuildCommand_HandlesPlaceholders(t *testing.T) {
	cfg := types.Language{FormatCommand: "echo ${flag:opt} ${anotherflag:tpo}"}
	opts := types.FormattingOptions{"opt": "value"}
	f := fileRef{Text: "text", LanguageID: "go", NormalizedFilename: "file.txt"}

	cmdStr, err := buildFormatCommandString("/root", &f, opts, nil, cfg)

	assert.NoError(t, err)

	assert.Contains(t, cmdStr, "flag value")
	assert.NotContains(t, cmdStr, "anotherflag")
	assert.Contains(t, cmdStr, "file.txt")
}

func TestFormatDocument_WithStdin(t *testing.T) {
	cfg := types.Language{FormatCommand: "cat -", FormatStdin: true}
	f := fileRef{Text: "hello text", LanguageID: "go", NormalizedFilename: "file.txt"}
	tmpDir := t.TempDir()

	out, err := formatDocument(t.Context(), tmpDir, f, nil, nil, cfg)

	assert.NoError(t, err)
	assert.Equal(t, "hello text", strings.TrimSpace(out))
}

func TestRunFormatters_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testfile := filepath.Join(tmpDir, "text.txt")
	err := os.WriteFile(testfile, []byte("test"), 0755)
	assert.NoError(t, err)

	h := &LangHandler{
		files: map[types.DocumentURI]*fileRef{
			types.DocumentURI("file://" + testfile): {Text: "hello", LanguageID: "go", NormalizedFilename: testfile},
		},
		configs: map[string][]types.Language{
			"go": {{FormatCommand: "cat", RequireMarker: false}},
		},
	}
	edits, err := h.RunAllFormatters(t.Context(), types.DocumentURI("file://"+testfile), nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, edits)
}

func TestRunFormatters_RequireRootMatcher(t *testing.T) {
	base, _ := os.Getwd()
	filepath := filepath.Join(base, "foo")
	uri := ParseLocalFileToURI(filepath)

	h := &LangHandler{
		RootPath: base,
		configs: map[string][]types.Language{
			"vim": {
				{
					LintCommand:        `echo ` + filepath + `:2:No it is normal!`,
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

	d, err := h.RunAllFormatters(t.Context(), uri, nil, types.FormattingOptions{})
	assert.NoError(t, err)
	assert.Empty(t, d)
}
