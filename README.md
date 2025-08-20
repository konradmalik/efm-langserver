# efm-langserver

[![Actions Status](https://github.com/konradmalik/efm-langserver/workflows/CI/badge.svg)](https://github.com/konradmalik/efm-langserver/actions)

General purpose Language Server that can spawn formatters and linters.

This is a fork [efm-langserver](https://github.com/mattn/efm-langserver) that will maintain and develop separately.
It is a cleaned up and simplified version of the original.
It supports the original configuration but only for formatting and linting. No code actions, completions, hover etc.

Notable changes from the original:

- no config.yaml, settings needs to be passed via DidChangeConfiguration
- only linting and formatting (for now)
- better diffs handling for formatting (no more "format twice to remove an extra newline")
    - external, maintained diff library used for that
- support for errorformat's end line and end column
- added tests (always in progress)
- refactored, cleaned and more maintainable code (always in progress)
- fixed and applied sane defaults for options like `LintAfterOpen`, `LintOnSave` etc.

![linting](./screenshot.png)

- [Installation](#installation)
- [Usage](#usage)
    - [Configuration](#configuration)
        - [InitializeParams](#initializeparams)
    - [Example for DidChangeConfiguration notification](#example-for-didchangeconfiguration-notification)
- [Client Setup](#client-setup)
    - [Configuration for neovim builtin LSP with nvim-lspconfig](#configuration-for-neovim-builtin-lsp-with-nvim-lspconfig)
    - [Configuration for coc.nvim](#configuration-for-cocnvim)
    - [Configuration for VSCode](#configuration-for-vscode)
    - [Configuration for Helix](#configuration-for-helix)
    - [Configuration for SublimeText LSP](#configuration-for-sublimetext-lsp)
    - [Configuration for vim-lsp](#configuration-for-vim-lsp)
    - [Configuration for Eglot (Emacs)](#configuration-for-eglot)
- [License](#license)
- [Author](#author)

## Installation

```console
go install github.com/konradmalik/efm-langserver@latest
```

or via `nix`. Flake provided in this repo.

```console
nix build
```

## Usage

```text
Usage of efm-langserver:
  -h    Show help
  -logfile string
        File to save logs into. If provided stderr won't be used anymore.
  -loglevel int
        Set the log level. Max is 5, min is 0. (default 1)
  -q    Run quiet
  -v    Print the version
```

### Configuration

Configuration can be done through a [DidChangeConfiguration](https://microsoft.github.io/language-server-protocol/specification.html#workspace_didChangeConfiguration)
notification from the client.
`DidChangeConfiguration` can be called any time and will overwrite only provided
properties (note though that per language configuration will be overwritten as a whole array).

`DidChangeConfiguration` cannot set `LogFile`.

`efm-langserver` does not include formatters/linters for any language. You must install these manually,
e.g.

- lua: [LuaFormatter](https://github.com/Koihik/LuaFormatter)
- python: [yapf](https://github.com/google/yapf) [isort](https://github.com/PyCQA/isort)
- [vint](https://github.com/Kuniwak/vint) for Vim script
- [markdownlint-cli](https://github.com/igorshubovych/markdownlint-cli) for Markdown
- etc...

#### InitializeParams

Because the configuration can be updated on the fly, capabilities might change
throughout the lifetime of the server. To enable support for capabilities that will
be available later, set them in the [InitializeParams](https://microsoft.github.io/language-server-protocol/specification.html#initialize)

Example

```json
{
    "initializationOptions": {
        "documentFormatting": true,
        "documentRangeFormatting": true
    }
}
```

### Example for DidChangeConfiguration notification

```json
{
    "settings": {
        "rootMarkers": [".git/"],
        "languages": {
            "lua": {
                "formatCommand": "lua-format -i",
                "formatStdin": true
            }
        }
    }
}
```

### Full config

```go
type Config struct {
	Version        int                    `json:"version,omitempty"`
	LogLevel       int                    `json:"logLevel,omitempty"`
	Languages      *map[string][]Language `json:"languages,omitempty"`
	RootMarkers    *[]string              `json:"rootMarkers,omitempty"`
	LintDebounce   time.Duration          `json:"lintDebounce,omitempty"`
	FormatDebounce time.Duration          `json:"formatDebounce,omitempty"`
}

type Language struct {
	Prefix             string             `json:"prefix,omitempty"`
	LintFormats        []string           `json:"lintFormats,omitempty"`
	LintStdin          bool               `json:"lintStdin,omitempty"`
	// warning: this will be subtracted from the line reported by the linter
	LintOffset         int                `json:"lintOffset,omitempty"`
 	// warning: this will be added to the column reported by the linter
	LintOffsetColumns  int                `json:"lintOffsetColumns,omitempty"`
	LintCommand        string             `json:"lintCommand,omitempty"`
	LintIgnoreExitCode bool               `json:"lintIgnoreExitCode,omitempty"`
	LintCategoryMap    map[string]string  `json:"lintCategoryMap,omitempty"`
	LintSource         string             `json:"lintSource,omitempty"`
	LintSeverity       DiagnosticSeverity `json:"lintSeverity,omitempty"`
	// defaults to true if not provided as a sanity default
	LintAfterOpen *bool `json:"lintAfterOpen,omitempty"`
	// defaults to true if not provided as a sanity default
	LintOnChange *bool `json:"lintOnChange,omitempty"`
	// defaults to true if not provided as a sanity default
	LintOnSave     *bool    `json:"lintOnSave,omitempty"`
	FormatCommand  string   `json:"formatCommand,omitempty"`
	FormatCanRange bool     `json:"formatCanRange,omitempty"`
	FormatStdin    bool     `json:"formatStdin,omitempty"`
	Env            []string `json:"env,omitempty"`
	RootMarkers    []string `json:"rootMarkers,omitempty"`
	RequireMarker  bool     `json:"requireMarker,omitempty"`
}
```

Also note that there's a wildcard for language name `=`. So if you want to define some config entry for all languages,
you can use `=` as a key.

## Client Setup

### Configuration for [neovim builtin LSP](https://neovim.io/doc/user/lsp.html) with [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig)

Neovim's built-in LSP client sends `DidChangeConfiguration`.

`init.lua` example (`settings` follows [`schema.md`](schema.md)):

```lua
require "lspconfig".efm.setup {
    init_options = {documentFormatting = true},
    settings = {
        rootMarkers = {".git/"},
        languages = {
            lua = {
                {formatCommand = "lua-format -i", formatStdin = true}
            }
        }
    }
}
```

You can get premade tool definitions from [`creativenull/efmls-configs-nvim`](https://github.com/creativenull/efmls-configs-nvim):

```lua
lua = {
  require('efmls-configs.linters.luacheck'),
  require('efmls-configs.formatters.stylua'),
}
```

If you define your own, make sure to define as a table of tables:

```lua
lua = {
    {formatCommand = "lua-format -i", formatStdin = true}
}

-- and for multiple formatters, add to the table
lua = {
    {formatCommand = "lua-format -i", formatStdin = true},
    {formatCommand = "lua-pretty -i"}
}
```

### Configuration for [coc.nvim](https://github.com/neoclide/coc.nvim)

coc-settings.json

```jsonc
  // languageserver
  "languageserver": {
    "efm": {
      "command": "efm-langserver",
      "args": [],
      "filetypes": ["vim", "eruby", "markdown", "yaml"]
    }
  },
```

### Configuration for [VSCode](https://github.com/microsoft/vscode)

[Generic LSP Client for VSCode](https://github.com/llllvvuu/vscode-glspc)

Example `settings.json` (change to fit your local installs):

```json
{
    "glspc.languageId": "lua",
    "glspc.serverCommand": "/Users/me/.local/share/nvim/mason/bin/efm-langserver",
    "glspc.pathPrepend": "/Users/me/.local/share/rtx/installs/python/3.11.4/bin:/Users/me/.local/share/rtx/installs/node/20.3.1/bin"
}
```

### Configuration for [Helix](https://github.com/helix-editor/helix)

`~/.config/helix/languages.toml`

```toml
[language-server.efm]
command = "efm-langserver"

[[language]]
name = "typescript"
language-servers = [
  { name = "efm", only-features = [ "diagnostics", "format" ] },
  { name = "typescript-language-server", except-features = [ "format" ] }
]
```

### Configuration for [SublimeText LSP](https://lsp.sublimetext.io)

Open `Preferences: LSP Settings` command from the Command Palette (Ctrl+Shift+P)

```
{
	"clients": {
	    "efm-langserver": {
	      "enabled": true,
	      "command": ["efm-langserver"],
	      "selector": "source.c | source.php | source.python" // see https://www.sublimetext.com/docs/3/selectors.html
	    }
  	}
}
```

### Configuration for [vim-lsp](https://github.com/prabirshrestha/vim-lsp/)

```vim
augroup LspEFM
  au!
  autocmd User lsp_setup call lsp#register_server({
      \ 'name': 'efm-langserver',
      \ 'cmd': {server_info->['efm-langserver']},
      \ 'allowlist': ['vim', 'eruby', 'markdown', 'yaml'],
      \ })
augroup END
```

[vim-lsp-settings](https://github.com/mattn/vim-lsp-settings) provide installer for efm-langserver.

### Configuration for [Eglot](https://github.com/joaotavora/eglot) (Emacs)

Add to eglot-server-programs with major mode you want.

```lisp
(with-eval-after-load 'eglot
  (add-to-list 'eglot-server-programs
    `(markdown-mode . ("efm-langserver"))))
```

## License

MIT

## Authors

- Yasuhiro Matsumoto (a.k.a. mattn) before 2025-04-29 (original efm-langserver author)
- Konrad Malik after 2025-04-29 (author and maintainer of this fork)
