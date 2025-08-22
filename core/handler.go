package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/konradmalik/efm-langserver/types"
)

type LangHandler struct {
	Loglevel       int
	Logger         *log.Logger
	configs        map[string][]types.Language
	files          map[types.DocumentURI]*fileRef
	LintDebounce   time.Duration
	FormatDebounce time.Duration
	RootPath       string
	rootMarkers    []string
}

type fileRef struct {
	Version            int
	NormalizedFilename string
	LanguageID         string
	Text               string
}

func NewConfig() *types.Config {
	languages := make(map[string][]types.Language)
	rootMarkers := make([]string, 0)
	return &types.Config{
		Languages:   &languages,
		RootMarkers: &rootMarkers,
	}
}

func NewHandler(logger *log.Logger, config *types.Config) *LangHandler {
	handler := &LangHandler{
		Loglevel:     config.LogLevel,
		Logger:       logger,
		configs:      *config.Languages,
		files:        make(map[types.DocumentURI]*fileRef),
		LintDebounce: config.LintDebounce,

		FormatDebounce: config.FormatDebounce,
		rootMarkers:    *config.RootMarkers,
	}
	return handler
}

func (h *LangHandler) Initialize(params types.InitializeParams) (types.InitializeResult, error) {
	if params.RootURI != "" {
		rootPath, err := PathFromURI(params.RootURI)
		if err != nil {
			return types.InitializeResult{}, err
		}
		h.RootPath = filepath.Clean(rootPath)
	}

	var hasFormatCommand bool
	var hasRangeFormatCommand bool

	if params.InitializationOptions != nil {
		hasFormatCommand = params.InitializationOptions.DocumentFormatting
		hasRangeFormatCommand = params.InitializationOptions.RangeFormatting
	}

	for _, config := range h.configs {
		for _, lang := range config {
			if lang.FormatCommand != "" {
				hasFormatCommand = true
				if lang.FormatCanRange {
					hasRangeFormatCommand = true
					break
				}
			}
		}
	}

	return types.InitializeResult{
		Capabilities: types.ServerCapabilities{
			TextDocumentSync:           types.TDSKFull,
			DocumentFormattingProvider: hasFormatCommand,
			RangeFormattingProvider:    hasRangeFormatCommand,
		},
	}, nil
}

func (h *LangHandler) UpdateConfiguration(config *types.Config) (any, error) {
	if config.Languages != nil {
		h.configs = *config.Languages
	}
	if config.RootMarkers != nil {
		h.rootMarkers = *config.RootMarkers
	}
	if config.LogLevel > 0 {
		h.Loglevel = config.LogLevel
	}
	if config.LintDebounce > 0 {
		h.LintDebounce = config.LintDebounce
	}
	if config.FormatDebounce > 0 {
		h.FormatDebounce = config.FormatDebounce
	}
	if config.LogLevel > 0 {
		h.Loglevel = config.LogLevel
	}

	return nil, nil
}

func (h *LangHandler) CloseFile(uri types.DocumentURI) error {
	delete(h.files, uri)
	return nil
}

func (h *LangHandler) OpenFile(uri types.DocumentURI, languageID string, version int, text string) error {
	fname, err := normalizedFilenameFromUri(uri)
	if err != nil {
		return err
	}

	f := &fileRef{
		Text:               text,
		LanguageID:         languageID,
		Version:            version,
		NormalizedFilename: fname,
	}
	h.files[uri] = f

	return nil
}

func (h *LangHandler) UpdateFile(uri types.DocumentURI, text string, version *int) error {
	f, ok := h.files[uri]
	if !ok {
		return fmt.Errorf("document not found: %v", uri)
	}
	f.Text = text
	if version != nil {
		f.Version = *version
	}

	return nil
}

func (h *LangHandler) findRootPath(fname string, lang types.Language) string {
	if dir := matchRootPath(fname, lang.RootMarkers); dir != "" {
		return dir
	}
	if dir := matchRootPath(fname, h.rootMarkers); dir != "" {
		return dir
	}

	return h.RootPath
}

func matchRootPath(fname string, markers []string) string {
	dir := filepath.Dir(fname)
	var prev string
	for dir != prev {
		files, _ := os.ReadDir(dir)
		for _, file := range files {
			name := file.Name()
			isDir := file.IsDir()
			for _, marker := range markers {
				if strings.HasSuffix(marker, "/") {
					if !isDir {
						continue
					}
					marker = strings.TrimRight(marker, "/")
					if ok, _ := filepath.Match(marker, name); ok {
						return dir
					}
				} else {
					if isDir {
						continue
					}
					if ok, _ := filepath.Match(marker, name); ok {
						return dir
					}
				}
			}
		}
		prev = dir
		dir = filepath.Dir(dir)
	}

	return ""
}

func isStdinPlaceholder(s string) bool {
	switch s {
	case "stdin", "-", "<text>", "<stdin>":
		return true
	default:
		return false
	}
}

func replaceCommandInputFilename(command, fname, rootPath string) string {
	ext := filepath.Ext(fname)
	ext = strings.TrimPrefix(ext, ".")

	command = strings.ReplaceAll(command, inputPlaceholder, escapeBrackets(fname))
	command = strings.ReplaceAll(command, "${FILEEXT}", ext)
	command = strings.ReplaceAll(command, "${FILENAME}", escapeBrackets(filepath.FromSlash(fname)))
	command = strings.ReplaceAll(command, "${ROOT}", escapeBrackets(rootPath))

	return command
}

func escapeBrackets(path string) string {
	path = strings.ReplaceAll(path, "(", `\(`)
	path = strings.ReplaceAll(path, ")", `\)`)

	return path
}
