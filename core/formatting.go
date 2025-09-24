package core

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/konradmalik/efm-langserver/logs"
	"github.com/konradmalik/efm-langserver/types"
)

var unfilledPlaceholders = regexp.MustCompile(`\${[^}]*}`)

func (h *LangHandler) RunAllFormatters(ctx context.Context, uri types.DocumentURI, rng *types.Range, options types.FormattingOptions) ([]types.TextEdit, error) {
	f, ok := h.files[uri]
	if !ok {
		return nil, fmt.Errorf("document not found: %v", uri)
	}

	configs, err := getFormatConfigsForDocument(f.NormalizedFilename, f.LanguageID, h.configs)
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		logs.Log.Logf(logs.Warn, "no matching format configs for LanguageID: %v", f.LanguageID)
		return nil, nil
	}

	originalText := f.Text
	formattedText := originalText
	formatted := false

	errors := make([]string, 0)
	for _, config := range configs {
		rootPath := h.findRootPath(f.NormalizedFilename, config)
		newText, err := formatDocument(ctx, rootPath, f.NormalizedFilename, formattedText, rng, options, config)

		if err != nil {
			errors = append(errors, err.Error())
			logs.Log.Logln(logs.Error, err.Error())
			continue
		}

		formatted = true
		formattedText = newText
	}

	if !formatted {
		return nil, fmt.Errorf("could not format for LanguageID: %s. All errors: %v", f.LanguageID, errors)
	}

	logs.Log.Logln(logs.Info, "format succeeded")
	return ComputeEdits(uri, originalText, formattedText)
}

// this needs to accept textToFormat because in case we have multiple formatters, we can pass previous formatted text.
// otherwise, we'd format the original file over and over.
func formatDocument(ctx context.Context, rootPath string, filename string, textToFormat string, rng *types.Range, options types.FormattingOptions, config types.Language) (string, error) {
	cmdStr, err := buildFormatCommandString(rootPath, filename, textToFormat, options, rng, config)
	if err != nil {
		return "", fmt.Errorf("command build error: %s", err)
	}

	cmd := buildExecCmd(ctx, cmdStr, rootPath, textToFormat, config, config.FormatStdin)
	out, err := runFormattingCommand(cmd)

	logs.Log.Logln(logs.Info, cmdStr)
	logs.Log.Logln(logs.Debug, out)

	if err != nil {
		return "", fmt.Errorf("formatting error: %s", err)
	}

	return strings.ReplaceAll(out, carriageReturn, ""), nil
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

func buildFormatCommandString(rootPath string, filename string, textToFormat string, options types.FormattingOptions, rng *types.Range, config types.Language) (string, error) {
	command := config.FormatCommand
	if !config.FormatStdin && !strings.Contains(command, inputPlaceholder) {
		command += " " + inputPlaceholder
	}
	command = replaceCommandInputFilename(command, filename, rootPath)

	var err error
	command, err = applyOptionsPlaceholders(command, options)
	if err != nil {
		return "", err
	}

	if rng != nil {
		command, err = applyRangePlaceholders(command, rng, textToFormat)
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

func getFormatConfigsForDocument(fname, langId string, allConfigs map[string][]types.Language) ([]types.Language, error) {
	addedCounter := 0
	var configs []types.Language
	for _, cfg := range getAllConfigsForLang(allConfigs, langId) {
		if cfg.FormatCommand == "" {
			continue
		}
		if dir := matchRootPath(fname, cfg.RootMarkers); dir == "" && cfg.RequireMarker {
			continue
		}

		if addedCounter > 0 && !cfg.FormatStdin {
			return nil, fmt.Errorf("format cfg for %s is invalid -> for multiple formatters, only the first one can be non-stdin", langId)
		}

		configs = append(configs, cfg)
		addedCounter++
	}

	return configs, nil
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
