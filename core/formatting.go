package core

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
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

	fname, err := normalizedFilenameFromUri(uri)
	if err != nil {
		return nil, err
	}

	configs := getFormatConfigsForDocument(fname, f.LanguageID, h.configs)
	if len(configs) == 0 {
		h.logUnsupportedFormat(f.LanguageID)
		return nil, nil
	}

	originalText := f.Text
	formattedText := originalText
	formatted := false

	for _, config := range configs {
		rootPath := h.findRootPath(fname, config)
		cmdStr, err := buildFormatCommandString(rootPath, fname, f, options, rng, config)
		if err != nil {
			h.logger.Println("command build error:", err)
			continue
		}

		cmd := buildExecCmd(context.Background(), cmdStr, rootPath, f, config, config.FormatStdin)
		out, err := runFormattingCommand(cmd)

		if h.loglevel >= 3 {
			h.logger.Println(cmdStr+":", string(out))
		}

		if err != nil {
			h.logger.Println("formatting error:", err)
			continue
		}

		formatted = true
		formattedText = strings.ReplaceAll(out, carriageReturn, "")
	}

	if !formatted {
		return nil, fmt.Errorf("format for LanguageID not supported: %v", f.LanguageID)
	}

	if h.loglevel >= 3 {
		h.logger.Println("format succeeded")
	}
	return ComputeEdits(uri, originalText, formattedText)
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
	lines := strings.Split(text, "\n")
	charStart := convertRowColToIndex(lines, rng.Start.Line, rng.Start.Character)
	charEnd := convertRowColToIndex(lines, rng.End.Line, rng.End.Character)

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

func buildFormatCommandString(rootPath, fname string, f *fileRef, options types.FormattingOptions, rng *types.Range, config types.Language) (string, error) {
	command := config.FormatCommand
	if !config.FormatStdin && !strings.Contains(command, inputPlaceholder) {
		command += " " + inputPlaceholder
	}
	command = replaceCommandInputFilename(command, fname, rootPath)

	var err error
	command, err = applyOptionsPlaceholders(command, options)
	if err != nil {
		return "", err
	}

	if rng != nil {
		command, err = applyRangePlaceholders(command, rng, f.Text)
		if err != nil {
			return "", err
		}
	}

	return unfilledPlaceholders.ReplaceAllString(command, ""), nil
}

func runFormattingCommand(cmd *exec.Cmd) (string, error) {
	var buf bytes.Buffer
	cmd.Stderr = &buf
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %s", strings.Join(cmd.Args, " "), buf.String())
	}
	return string(b), nil
}

func (h *LangHandler) logUnsupportedFormat(langID string) {
	if h.loglevel >= 2 {
		h.logger.Printf("format for LanguageID not supported: %v", langID)
	}
}

func getFormatConfigsForDocument(fname, langId string, allConfigs map[string][]types.Language) []types.Language {
	var configs []types.Language
	for _, cfg := range getAllConfigsForLang(allConfigs, langId) {
		if cfg.FormatCommand == "" {
			continue
		}
		if dir := matchRootPath(fname, cfg.RootMarkers); dir == "" && cfg.RequireMarker {
			continue
		}
		configs = append(configs, cfg)
	}
	return configs
}

func convertRowColToIndex(lines []string, row, col int) int {
	row = max(row, 0)
	row = min(row, len(lines)-1)

	col = max(col, 0)
	col = min(col, len(lines[row]))

	index := 0
	for i := 0; i < row; i++ {
		// Add the length of each line plus 1 for the newline character
		index += len(lines[i]) + 1
	}
	index += col

	return index
}
