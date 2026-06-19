# FormaTeX CLI

Compile LaTeX documents to PDF from the command line — no local TeX Live installation required.

```
formatex compile thesis.tex
```

---

## Installation

### macOS / Linux (curl)

```bash
curl -fsSL https://raw.githubusercontent.com/formatexio/cli/main/install.sh | sh
```

### macOS (Homebrew)

```bash
brew install formatexio/tap/formatex
```

### Windows (PowerShell)

```powershell
iwr https://raw.githubusercontent.com/formatexio/cli/main/install.ps1 | iex
```

### Manual download

Download the binary for your platform from the [releases page](https://github.com/formatexio/cli/releases/latest):

| Platform | File |
|---|---|
| macOS (Apple Silicon) | `formatex-darwin-arm64` |
| macOS (Intel) | `formatex-darwin-amd64` |
| Linux (x86_64) | `formatex-linux-amd64` |
| Linux (ARM64) | `formatex-linux-arm64` |
| Windows (x86_64) | `formatex-windows-amd64.exe` |

Make the binary executable and move it to your PATH:

```bash
chmod +x formatex-linux-amd64
mv formatex-linux-amd64 /usr/local/bin/formatex
```

### Build from source

```bash
go install github.com/formatexio/cli@latest
```

---

## Authentication

Get an API key from [formatex.io/dashboard/api-keys](https://formatex.io/dashboard/api-keys), then run:

```bash
formatex login
```

You can also set the `FORMATEX_API_KEY` environment variable, or pass `--api-key` to any command.

**Priority:** `--api-key` flag > `FORMATEX_API_KEY` env > saved config

---

## Commands

### `formatex compile`

Compile a `.tex` file or `.zip` project to PDF.

```bash
formatex compile <file.tex|project.zip> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--engine`, `-e` | `pdflatex` | Engine: `pdflatex`, `xelatex`, `lualatex`, `latexmk` |
| `--output`, `-o` | `<input>.pdf` | Output PDF path |
| `--runs` | `0` (auto) | Number of compiler passes (1–5) |
| `--timeout` | `0` (plan default) | Max compile time in seconds |
| `--main` | auto-detected | Entry-point `.tex` inside a ZIP archive |
| `--open` | `false` | Open the PDF after compilation |

**Examples:**

```bash
# Basic compile
formatex compile thesis.tex

# Use XeLaTeX (required for Unicode, custom fonts)
formatex compile thesis.tex --engine xelatex

# Custom output path
formatex compile thesis.tex --output dist/thesis.pdf

# Compile a multi-file ZIP project
formatex compile project.zip

# Specify the main .tex file inside the ZIP
formatex compile project.zip --main chapters/main.tex

# Open the PDF immediately after compiling
formatex compile thesis.tex --open
```

**On failure**, structured errors are printed with file and line number. If AI explanation is available on your plan, it is shown automatically:

```
Compilation failed.

Errors:
  main.tex:42: Undefined control sequence \foo

AI: The command \foo is not defined. You may have meant \footnote, or you
    need to add \newcommand{\foo}{...} to your preamble.
```

---

### `formatex watch`

Watch a `.tex` file and recompile automatically on every save.

```bash
formatex watch <file.tex> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--engine`, `-e` | `pdflatex` | LaTeX engine |
| `--output`, `-o` | `<input>.pdf` | Output PDF path |
| `--runs` | `0` | Number of compiler passes |
| `--timeout` | `0` | Max compile time in seconds |
| `--open` | `false` | Open the PDF on first successful compile |

**Example:**

```bash
# Watch and recompile; open the PDF once on first success
formatex watch thesis.tex --engine xelatex --open
```

Output:

```
Watching /home/alice/thesis.tex (engine: xelatex)
Press Ctrl+C to stop.
[14:23:01] Compiling... ok (2.1s, 184.3 KB)
[14:24:17] Compiling... ok (2.0s, 184.5 KB)
[14:25:03] Compiling... FAILED
  main.tex:88: Missing $ inserted
```

---

### `formatex format`

Format a LaTeX source file using `latexindent`.

```bash
formatex format <file.tex> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--write`, `-w` | `false` | Write formatted output back to the source file |
| `--output`, `-o` | — | Write formatted output to a different file |

By default, formatted output is printed to stdout so you can preview or pipe it.

**Examples:**

```bash
# Preview formatted output
formatex format main.tex

# Overwrite the file in-place
formatex format main.tex --write

# Write to a new file
formatex format main.tex --output main-formatted.tex

# Integrate with git pre-commit
formatex format main.tex --write && git add main.tex
```

---

### `formatex convert`

Convert a LaTeX document to another format.

```bash
formatex convert <file.tex> [flags]
```

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--format`, `-f` | `docx` | Output format: `docx`, `html`, `odt`, `epub`, `markdown`, `txt` |
| `--output`, `-o` | `<input>.<ext>` | Output file path |

**Examples:**

```bash
# Convert to Word (DOCX) — default
formatex convert thesis.tex

# Convert to HTML
formatex convert thesis.tex --format html

# Convert to ODT (LibreOffice)
formatex convert thesis.tex --format odt --output thesis.odt

# Convert to plain text
formatex convert thesis.tex --format txt
```

---

### `formatex login`

Save your API key to `~/.config/formatex/config.json`. The key is verified against the API before saving.

```bash
formatex login
```

```
Enter your FormaTeX API key: fex_••••••••••••••••••
Verifying API key... ok
Logged in as alice@example.com
```

---

### `formatex version`

Print the installed CLI version.

```bash
formatex version
# formatex version v1.0.0
```

---

## Global flags

Available on every command:

| Flag | Description |
|---|---|
| `--api-key` | API key (overrides config and env var) |
| `--base-url` | API base URL (default: `https://api.formatex.io`) |

---

## Environment variables

| Variable | Description |
|---|---|
| `FORMATEX_API_KEY` | API key (alternative to `formatex login`) |
| `FORMATEX_BASE_URL` | Override the API base URL |

---

## Configuration

The config file is stored at:

| Platform | Path |
|---|---|
| macOS / Linux | `~/.config/formatex/config.json` |
| Windows | `%APPDATA%\formatex\config.json` |

Contents:

```json
{
  "api_key": "fex_your_key_here",
  "base_url": "https://api.formatex.io"
}
```

---

## CI / CD

Use `FORMATEX_API_KEY` as a secret in your pipeline:

**GitHub Actions:**

```yaml
- name: Compile LaTeX
  run: formatex compile thesis.tex --output dist/thesis.pdf
  env:
    FORMATEX_API_KEY: ${{ secrets.FORMATEX_API_KEY }}
```

For GitHub Actions with artifact upload, consider the dedicated [FormaTeX GitHub Action](https://github.com/formatexio/action) instead.

---

## Troubleshooting

**`no API key found`** — Run `formatex login` or set `FORMATEX_API_KEY`.

**`compilation failed`** — Check the printed errors. Common causes: missing packages, undefined commands, missing `\end{document}`.

**`file not found`** — Make sure the path is correct and the file exists.

**Timeout errors** — Large documents may exceed the default timeout. Pass `--timeout 120` to allow more time, or check your plan limits.

---

## License

MIT
