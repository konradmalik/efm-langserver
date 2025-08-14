package core

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/konradmalik/efm-langserver/types"
	"github.com/reviewdog/errorformat"
)

var defaultLintFormats = []string{"%f:%l:%m", "%f:%l:%c:%m"}
var running = make(map[types.DocumentURI]context.CancelFunc)

type notifier interface {
	PublishDiagnostics(ctx context.Context, params types.PublishDiagnosticsParams)
	LogMessage(ctx context.Context, typ types.MessageType, message string)
}

func (h *LangHandler) ScheduleLinting(notifier notifier, uri types.DocumentURI, eventType types.EventType) {
	if h.lintTimer != nil {
		h.lintTimer.Reset(h.lintDebounce)
		if h.loglevel >= 4 {
			h.logger.Printf("lint debounced: %v", h.lintDebounce)
		}
		return
	}
	h.lintMu.Lock()
	h.lintTimer = time.AfterFunc(h.lintDebounce, func() {
		h.lintTimer = nil

		h.lintMu.Lock()
		cancel, ok := running[uri]
		if ok {
			cancel()
		}

		ctx, cancel := context.WithCancel(context.Background())
		running[uri] = cancel
		h.lintMu.Unlock()
		go h.runLintersPublishDiagnostics(ctx, notifier, uri, eventType)
	})
	h.lintMu.Unlock()
}

func (h *LangHandler) runLintersPublishDiagnostics(ctx context.Context, notifier notifier, uri types.DocumentURI, eventType types.EventType) {
	uriToDiagnostics, err := h.lintDocument(ctx, notifier, uri, eventType)
	if err != nil {
		h.logger.Println(err)
		return
	}

	for diagURI, diagnostics := range uriToDiagnostics {
		if diagURI == "file:" {
			diagURI = uri
		}
		version := 0
		if _, ok := h.files[uri]; ok {
			version = h.files[uri].Version
		}
		notifier.PublishDiagnostics(
			ctx,
			types.PublishDiagnosticsParams{
				URI:         diagURI,
				Diagnostics: diagnostics,
				Version:     version,
			})
	}
}

func (h *LangHandler) lintDocument(ctx context.Context, notifier notifier, uri types.DocumentURI, eventType types.EventType) (map[types.DocumentURI][]types.Diagnostic, error) {
	f, ok := h.files[uri]
	if !ok {
		return nil, fmt.Errorf("document not found: %v", uri)
	}

	fname, err := normalizedFilenameFromUri(uri)
	if err != nil {
		return nil, err
	}

	configs := getLintConfigsForDocument(fname, f.LanguageID, h.configs, eventType)
	if len(configs) == 0 {
		h.logUnsupportedLint(f.LanguageID)
		return nil, nil
	}

	uriToDiagnostics := map[types.DocumentURI][]types.Diagnostic{
		uri: {},
	}

	for _, config := range configs {
		rootPath := h.findRootPath(fname, config)
		command := buildLintCommand(ctx, rootPath, f, fname, &config)

		lintOutput, err := runLintCommand(command, &config)
		if h.loglevel >= 3 {
			h.logger.Println(config.LintCommand+":", string(lintOutput))
		}
		if err != nil {
			notifier.LogMessage(ctx, types.LogError, err.Error())
			h.logger.Println(err)
			continue
		}

		efms, err := buildErrorformats(config.LintFormats)
		if err != nil {
			return nil, err
		}

		efmsScanner := efms.NewScanner(bytes.NewReader(lintOutput))
		for efmsScanner.Scan() {
			entry := efmsScanner.Entry()
			if !entry.Valid {
				continue
			}

			entry.Filename = replaceStdinInEntryFilename(entry.Filename, &config, fname)
			if !isEntryForRequestedURI(rootPath, uri, entry) {
				// entry for a different file, skip
				continue
			}

			diagnostic := parseEfmEntryToDiagnostic(entry, &config, f)
			uriToDiagnostics[uri] = append(uriToDiagnostics[uri], diagnostic)
		}
	}

	return uriToDiagnostics, nil
}

func getSeverity(typ rune, categoryMap map[string]string, defaultSeverity types.DiagnosticSeverity) types.DiagnosticSeverity {
	// we allow the config to provide a mapping between LSP types E,W,I,N and whatever categories the linter has
	if len(categoryMap) > 0 {
		typ = []rune(categoryMap[string(typ)])[0]
	}

	severity := types.Error
	if defaultSeverity != 0 {
		severity = defaultSeverity
	}

	switch typ {
	case 'E', 'e':
		severity = types.Error
	case 'W', 'w':
		severity = types.Warning
	case 'I', 'i':
		severity = types.Information
	case 'N', 'n':
		severity = types.Hint
	}
	return severity
}

func getLintConfigsForDocument(fname, langId string, allConfigs map[string][]types.Language, eventType types.EventType) []types.Language {
	var configs []types.Language
	for _, cfg := range getAllConfigsForLang(allConfigs, langId) {
		if cfg.LintCommand == "" {
			continue
		}
		// if we require markers and find that they dont exist we do not add the configuration
		if dir := matchRootPath(fname, cfg.RootMarkers); dir == "" && cfg.RequireMarker {
			continue
		}
		switch eventType {
		case types.EventTypeOpen:
			if !boolOrDefault(cfg.LintAfterOpen, true) {
				continue
			}
		case types.EventTypeChange:
			if !boolOrDefault(cfg.LintOnChange, true) {
				continue
			}
		case types.EventTypeSave:
			if !boolOrDefault(cfg.LintOnSave, true) {
				continue
			}
		default:
		}
		configs = append(configs, cfg)
	}
	return configs
}

func buildErrorformats(configFormats []string) (*errorformat.Errorformat, error) {
	if len(configFormats) == 0 {
		configFormats = defaultLintFormats
	}

	efms, err := errorformat.NewErrorformat(configFormats)
	if err != nil {
		return nil, fmt.Errorf("invalid error-format: %v", configFormats)
	}
	return efms, nil
}

func buildLintCommand(ctx context.Context, rootPath string, f *fileRef, fname string, config *types.Language) *exec.Cmd {
	command := config.LintCommand
	if !config.LintStdin && !strings.Contains(command, inputPlaceholder) {
		command = command + " " + inputPlaceholder
	}
	command = replaceCommandInputFilename(command, fname, rootPath)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, windowsShell, windowsShellArg, command)
	} else {
		cmd = exec.CommandContext(ctx, unixShell, unixShellArg, command)
	}
	cmd.Dir = rootPath
	cmd.Env = append(os.Environ(), config.Env...)
	if config.LintStdin {
		cmd.Stdin = strings.NewReader(f.Text)
	}

	return cmd
}

func runLintCommand(cmd *exec.Cmd, config *types.Language) ([]byte, error) {
	lintOutput, lintCmdError := cmd.CombinedOutput()
	// Most of lint tools exit with non-zero value. But some commands
	// return with zero value. We can not handle the output is real result
	// or output of usage. So efm-langserver ignore that command exiting
	// with zero-value. So if you want to handle the command which exit
	// with zero value, please specify lint-ignore-exit-code.
	if !config.LintIgnoreExitCode && lintCmdError == nil {
		return lintOutput, fmt.Errorf("command `%s` exit with zero. Probably you forgot to specify `lint-ignore-exit-code: true`", config.LintCommand)
	}
	return lintOutput, nil
}

func replaceStdinInEntryFilename(entryFilename string, config *types.Language, fname string) string {
	if config.LintStdin && isStdinPlaceholder(entryFilename) {
		entryFilename = fname
	}
	return filepath.ToSlash(entryFilename)
}

func isEntryForRequestedURI(rootPath string, uri types.DocumentURI, entry *errorformat.Entry) bool {
	// if entry.Filename is empty, we simply assume it's for this file
	if entry.Filename == "" {
		return true
	}
	// if entry.Filename is not empty, we need to check if this entry is indeed for this uri
	var diagURI types.DocumentURI
	if filepath.IsAbs(entry.Filename) {
		diagURI = toURI(entry.Filename)
	} else {
		diagURI = toURI(filepath.Join(rootPath, entry.Filename))
	}
	// windows FS is case insensitive
	if runtime.GOOS == "windows" {
		return strings.EqualFold(string(diagURI), string(uri))
	}
	return diagURI == uri
}

func parseEfmEntryToDiagnostic(entry *errorformat.Entry, config *types.Language, f *fileRef) types.Diagnostic {
	linePos := max(entry.Lnum-1-config.LintOffset, 0)
	colPos := max(entry.Col-1, 0)

	// entry.Col is expected to be one based
	// if the linter reports 0 it means the whole line
	word := ""
	if entry.Col != 0 {
		// have the ability to add an offset here.
		// We only add the offset if the linter reports entry.Col > 0 because 0 means the whole line
		colPos = colPos + config.LintOffsetColumns
		word = f.wordAt(types.Position{Line: linePos, Character: colPos})
	}

	return types.Diagnostic{
		Range: types.Range{
			Start: types.Position{Line: linePos, Character: colPos},
			// len(runes) counts unicode code points, not bytes, which is what we want here
			End: types.Position{Line: linePos, Character: colPos + len([]rune(word))},
		},
		Code:     itoaPtrIfNotZero(entry.Nr),
		Message:  getLintMessagePrefix(config) + entry.Text,
		Severity: getSeverity(entry.Type, config.LintCategoryMap, config.LintSeverity),
		Source:   getLintSource(config),
	}
}

func (h *LangHandler) logUnsupportedLint(langID string) {
	if h.loglevel >= 2 {
		h.logger.Printf("lint for LanguageID not supported: %v", langID)
	}
}

func getLintSource(config *types.Language) *string {
	if config.LintSource != "" {
		return &config.LintSource
	}
	return nil
}

func getLintMessagePrefix(config *types.Language) string {
	var prefix string
	if config.Prefix != "" {
		prefix = fmt.Sprintf("[%s] ", config.Prefix)
	}
	return prefix
}
