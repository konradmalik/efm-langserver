package core

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/konradmalik/efm-langserver/types"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeFilename(t *testing.T) {
	uri := types.DocumentURI("file:///tmp/testfile.txt")
	fname, err := normalizeFilename(uri)
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(fname, "/tmp/testfile.txt"))

	if runtime.GOOS == "windows" {
		assert.Equal(t, strings.ToLower(fname), fname, "filename should be lowercase on Windows")
	}
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

func TestBuildCommand_RemovesUnfilled(t *testing.T) {
	cfg := types.Language{FormatCommand: "echo ${flag:opt}"}
	opts := types.FormattingOptions{"opt": "value"}
	cmd, err := buildCommand(cfg, "file.txt", opts, nil, "text", "/root")
	assert.NoError(t, err)
	assert.NotContains(t, cmd, "${")
}

func TestApplyFormattingCommand_WithStdin(t *testing.T) {
	tmpDir := t.TempDir()
	script := filepath.Join(tmpDir, "script.sh")
	scriptContent := "#!/bin/sh\ncat -"
	if runtime.GOOS == "windows" {
		script = filepath.Join(tmpDir, "script.bat")
		scriptContent = "@echo off\ntype con"
	}
	err := os.WriteFile(script, []byte(scriptContent), 0755)
	assert.NoError(t, err)

	cmd := script
	out, err := applyFormattingCommand(cmd, "hello", tmpDir, nil, true)
	assert.NoError(t, err)
	assert.Equal(t, "hello", strings.TrimSpace(out))
}

func TestRangeFormatting_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testfile := filepath.Join(tmpDir, "text.txt")
	err := os.WriteFile(testfile, []byte("test"), 0755)
	assert.NoError(t, err)

	h := &LangHandler{
		files: map[types.DocumentURI]*fileRef{
			types.DocumentURI("file://" + testfile): {Text: "hello", LanguageID: "go"},
		},
		configs: map[string][]types.Language{
			"go": {{FormatCommand: "cat", RequireMarker: false}},
		},
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		loglevel: 3,
	}
	edits, err := h.rangeFormatting(types.DocumentURI("file://"+testfile), nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, edits)
}

func TestRangeFormatting_RequireRootMatcher(t *testing.T) {
	base, _ := os.Getwd()
	filepath := filepath.Join(base, "foo")
	uri := toURI(filepath)

	h := &LangHandler{
		logger:   log.New(log.Writer(), "", log.LstdFlags),
		RootPath: base,
		configs: map[string][]types.Language{
			"vim": {
				{
					LintCommand:        `echo ` + filepath + `:2:No it is normal!`,
					LintIgnoreExitCode: true,
					LintAfterOpen:      true,
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

	d, err := h.Formatting(uri, nil, types.FormattingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 0 {
		t.Fatal("text edits should be zero as we have no root marker for the language but require one", d)
	}
}
