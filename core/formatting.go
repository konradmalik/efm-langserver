package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/konradmalik/efm-langserver/types"
)

func (h *LangHandler) Formatting(uri types.DocumentURI, rng *types.Range, opt types.FormattingOptions) ([]types.TextEdit, error) {
	if h.formatTimer != nil {
		if h.loglevel >= 4 {
			h.logger.Printf("format debounced: %v", h.formatDebounce)
		}
		return []types.TextEdit{}, nil
	}

	h.formatMu.Lock()
	h.formatTimer = time.AfterFunc(h.formatDebounce, func() {
		h.formatMu.Lock()
		h.formatTimer = nil
		h.formatMu.Unlock()
	})
	h.formatMu.Unlock()
	return h.rangeFormatting(uri, rng, opt)
}

var (
	unfilledPlaceholders = regexp.MustCompile(`\${[^}]*}`)
)

// rangeFormatting formats a document or a selected range using configured formatters.
func (h *LangHandler) rangeFormatting(uri types.DocumentURI, rng *types.Range, options types.FormattingOptions) ([]types.TextEdit, error) {
	f, ok := h.files[uri]
	if !ok {
		return nil, fmt.Errorf("document not found: %v", uri)
	}

	fname, err := normalizeFilename(uri)
	if err != nil {
		return nil, err
	}

	configs := formatConfigsForDocument(fname, f.LanguageID, h.configs)
	if len(configs) == 0 {
		h.logUnsupportedFormat(f.LanguageID)
		return nil, nil
	}

	originalText := f.Text
	formattedText := originalText
	formatted := false

	for _, config := range configs {
		cmdStr, err := buildFormatCommand(config, fname, options, rng, formattedText, h.RootPath)
		if err != nil {
			h.logger.Println("command build error:", err)
			continue
		}

		out, err := applyFormattingCommand(
			cmdStr,
			formattedText,
			h.findRootPath(fname, config),
			config.Env,
			config.FormatStdin,
		)
		if err != nil {
			h.logger.Println("formatting error:", err)
			continue
		}

		formatted = true
		if h.loglevel >= 3 {
			h.logger.Println(cmdStr+":", string(out))
		}
		formattedText = strings.ReplaceAll(out, newlineChar, "")
	}

	if !formatted {
		return nil, fmt.Errorf("format for LanguageID not supported: %v", f.LanguageID)
	}

	if h.loglevel >= 3 {
		h.logger.Println("format succeeded")
	}
	return ComputeEdits(uri, originalText, formattedText)
}

func buildFormatCommand(config types.Language, fname string, options types.FormattingOptions, rng *types.Range, text, rootPath string) (string, error) {
	if config.FormatCommand == "" {
		return "", errors.New("empty format command")
	}

	cmd := config.FormatCommand
	if !config.FormatStdin && !strings.Contains(cmd, inputPlaceholder) {
		cmd += " " + inputPlaceholder
	}
	cmd = replaceCommandInputFilename(cmd, fname, rootPath)

	var err error
	cmd, err = applyOptionsPlaceholders(cmd, options)
	if err != nil {
		return "", err
	}

	if rng != nil {
		cmd, err = applyRangePlaceholders(cmd, rng, text)
		if err != nil {
			return "", err
		}
	}

	cmd = unfilledPlaceholders.ReplaceAllString(cmd, "")
	return cmd, nil
}

func applyOptionsPlaceholders(command string, options types.FormattingOptions) (string, error) {
	for placeholder, value := range options {
		re, err := regexp.Compile(fmt.Sprintf(`\${([^:|^}]+):%s}`, placeholder))
		re2, err2 := regexp.Compile(fmt.Sprintf(`\${([^=|^}]+)=%s}`, placeholder))
		nre, nerr := regexp.Compile(fmt.Sprintf(`\${([^:|^}]+):!%s}`, placeholder))
		nre2, nerr2 := regexp.Compile(fmt.Sprintf(`\${([^=|^}]+)=!%s}`, placeholder))
		if err != nil || err2 != nil || nerr != nil || nerr2 != nil {
			return command, fmt.Errorf("invalid option placeholder regex for %s", placeholder)
		}

		switch v := value.(type) {
		default:
			command = re.ReplaceAllString(command, fmt.Sprintf("%s %v", flagPlaceholder, v))
			command = re2.ReplaceAllString(command, fmt.Sprintf("%s=%v", flagPlaceholder, v))
		case bool:
			if v {
				command = re.ReplaceAllString(command, flagPlaceholder)
				command = re2.ReplaceAllString(command, flagPlaceholder)
			} else {
				command = nre.ReplaceAllString(command, flagPlaceholder)
				command = nre2.ReplaceAllString(command, flagPlaceholder)
			}
		}
	}
	return command, nil
}

func applyRangePlaceholders(command string, rng *types.Range, text string) (string, error) {
	charStart := convertRowColToIndex(text, rng.Start.Line, rng.Start.Character)
	charEnd := convertRowColToIndex(text, rng.End.Line, rng.End.Character)

	rangeOptions := map[string]int{
		"charStart": charStart,
		"charEnd":   charEnd,
		"rowStart":  rng.Start.Line,
		"colStart":  rng.Start.Character,
		"rowEnd":    rng.End.Line,
		"colEnd":    rng.End.Character,
	}

	for placeholder, value := range rangeOptions {
		re, err := regexp.Compile(fmt.Sprintf(`\${([^:|^}]+):%s}`, placeholder))
		re2, err2 := regexp.Compile(fmt.Sprintf(`\${([^=|^}]+)=%s}`, placeholder))
		if err != nil || err2 != nil {
			return command, fmt.Errorf("invalid range placeholder regex for %s", placeholder)
		}
		command = re.ReplaceAllString(command, fmt.Sprintf("%s %d", flagPlaceholder, value))
		command = re2.ReplaceAllString(command, fmt.Sprintf("%s=%d", flagPlaceholder, value))
	}

	return command, nil
}

func applyFormattingCommand(command, text, workDir string, env []string, formatStdin bool) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(windowsShell, windowsShellArg, command)
	} else {
		cmd = exec.Command(unixShell, unixShellArg, command)
	}
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), env...)
	if formatStdin {
		cmd.Stdin = strings.NewReader(text)
	}
	var buf bytes.Buffer
	cmd.Stderr = &buf
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %s", command, buf.String())
	}
	return string(b), nil
}

func (h *LangHandler) logUnsupportedFormat(langID string) {
	if h.loglevel >= 2 {
		h.logger.Printf("format for LanguageID not supported: %v", langID)
	}
}

func formatConfigsForDocument(fname, langId string, allConfigs map[string][]types.Language) []types.Language {
	var configs []types.Language
	if cfgs, ok := allConfigs[langId]; ok {
		for _, cfg := range cfgs {
			if cfg.FormatCommand != "" {
				if dir := matchRootPath(fname, cfg.RootMarkers); dir == "" && cfg.RequireMarker {
					continue
				}
				configs = append(configs, cfg)
			}
		}
	}
	if cfgs, ok := allConfigs[types.Wildcard]; ok {
		for _, cfg := range cfgs {
			if cfg.FormatCommand != "" {
				configs = append(configs, cfg)
			}
		}
	}
	return configs
}

func convertRowColToIndex(s string, row, col int) int {
	lines := strings.Split(s, "\n")

	if row < 0 {
		row = 0
	} else if row >= len(lines) {
		row = len(lines) - 1
	}

	if col < 0 {
		col = 0
	} else if col > len(lines[row]) {
		col = len(lines[row])
	}

	index := 0
	for i := 0; i < row; i++ {
		// Add the length of each line plus 1 for the newline character
		index += len(lines[i]) + 1
	}
	index += col

	return index
}
