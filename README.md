NOTE: this is a fork [efm-langserver](https://github.com/mattn/efm-langserver) that I aim to maintain and develop separately.
It will be a cleaned up and simplified version of the original.

It will support the original configuration but only for formatting and linting. No code actions, completions, hover etc.

# efm-langserver

[![Actions Status](https://github.com/konradmalik/efm-langserver/workflows/CI/badge.svg)](https://github.com/konradmalik/efm-langserver/actions)

General purpose Language Server that can use specified error message format
generated from specified command. This is useful for editing code with linter.

![efm](https://raw.githubusercontent.com/konradmalik/efm-langserver/master/screenshot.png)

- [Installation](#installation)
- [Usage](#usage)
    - [Configuration](#configuration)
        - [InitializeParams](#initializeparams)
    - [Example for DidChangeConfiguration notification](#example-for-didchangeconfiguration-notification)
- [Client Setup](#client-setup)
    - [Configuration for vim-lsp](#configuration-for-vim-lsp)
    - [Configuration for coc.nvim](#configuration-for-cocnvim)
    - [Configuration for Eglot (Emacs)](#configuration-for-eglot)
    - [Configuration for neovim builtin LSP with nvim-lspconfig](#configuration-for-neovim-builtin-lsp-with-nvim-lspconfig)
    - [Configuration for Helix](#configuration-for-helix)
    - [Configuration for VSCode](#configuration-for-vscode)
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

`efm-langserver` does not include formatters/linters for any languages, you must install these manually,
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

## Client Setup

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

### Configuration for [Eglot](https://github.com/joaotavora/eglot) (Emacs)

Add to eglot-server-programs with major mode you want.

```lisp
(with-eval-after-load 'eglot
  (add-to-list 'eglot-server-programs
    `(markdown-mode . ("efm-langserver"))))
```

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

If you define your own, make sure to define as table:

```lua
lua = {
    {formatCommand = "lua-format -i", formatStdin = true}
}
-- NOT
lua = {
    formatCommand = "lua-format -i", formatStdin = true
}

-- and for multiple formatters, add to the table
lua = {
    {formatCommand = "lua-format -i", formatStdin = true},
    {formatCommand = "lua-pretty -i"}
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

## License

MIT

## Authors

- Yasuhiro Matsumoto (a.k.a. mattn) before 2025-04-29 (original efm-langserver author)
- Konrad Malik after 2025-04-29 (author and maintainer of this fork)
