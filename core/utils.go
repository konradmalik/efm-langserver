package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/konradmalik/efm-langserver/types"
)

const (
	flagPlaceholder  = "$1"
	inputPlaceholder = "${INPUT}"
	carriageReturn   = "\r"
)

func normalizedFilenameFromUri(uri types.DocumentURI) (string, error) {
	fname, err := fromURI(uri)
	if err != nil {
		return "", fmt.Errorf("invalid uri: %v: %v", err, uri)
	}
	fname = filepath.ToSlash(fname)
	if runtime.GOOS == "windows" {
		fname = strings.ToLower(fname)
	}
	return fname, nil
}

func getAllConfigsForLang(allConfigs map[string][]types.Language, langId string) []types.Language {
	configsForLang := make([]types.Language, 0)
	if cfgs, ok := allConfigs[langId]; ok {
		configsForLang = append(configsForLang, cfgs...)
	}
	if cfgs, ok := allConfigs[types.Wildcard]; ok {
		configsForLang = append(configsForLang, cfgs...)
	}
	return configsForLang
}

func buildExecCmd(ctx context.Context, command, rootPath string, f *fileRef, config types.Language, stdin bool) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Dir = rootPath
	cmd.Env = append(os.Environ(), config.Env...)
	if stdin {
		cmd.Stdin = strings.NewReader(f.Text)
	}

	return cmd
}

func itoaPtrIfNotZero(n int) *string {
	if n == 0 {
		return nil
	}
	s := strconv.Itoa(n)
	return &s
}

func boolOrDefault(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

func boolPtr(v bool) *bool { return &v }
