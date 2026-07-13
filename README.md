# Jurnal

[![Go Report Card](https://goreportcard.com/badge/github.com/luckywirasakti/jurnal)](https://goreportcard.com/report/github.com/luckywirasakti/jurnal)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Releases](https://img.shields.io/github/v/release/luckywirasakti/jurnal)](https://github.com/luckywirasakti/jurnal/releases)

Jurnal is an AI-powered Git workflow assistant designed to plan, stage, and execute modular commits using OpenAI-compatible APIs. It analyzes your unstaged changes and suggests clean, semantic commits grouped by logical context.

Written entirely in native Go, Jurnal compiles into a single static binary with **zero external dependencies** and starts up instantly (<10ms).

---

## Features

- **Semantic Commit Grouping:** Automatically batches related modifications (e.g., separating migrations, logic, and tests) into clean commits.
- **Interactive Staging & Commit Flow:** Switch to proposed branch names and apply staging batches sequentially with terminal prompts.
- **Zero Runtime Overhead:** Pre-compiled native binaries for macOS, Linux, and Windows. No Python interpreter, PyInstaller unpack delays, or node runtime required.
- **Flexible Configuration:** Supports local JSON configuration store (`~/.jurnal.json`) as well as environment variables for non-interactive pipelines (CI/CD).

---

## Installation

### Quick Install (macOS & Linux)

Run the script below in your terminal to fetch and set up the latest compiled binary:

```bash
# macOS (Apple Silicon & Intel)
sudo curl -L -o /usr/local/bin/jurnal https://github.com/luckywirasakti/jurnal/releases/latest/download/jurnal-macos && sudo chmod +x /usr/local/bin/jurnal

# Linux (amd64)
sudo curl -L -o /usr/local/bin/jurnal https://github.com/luckywirasakti/jurnal/releases/latest/download/jurnal-linux && sudo chmod +x /usr/local/bin/jurnal
```

*Note for macOS users:* If macOS blocks execution with a security prompt, strip the quarantine attribute:
```bash
xattr -d com.apple.quarantine /usr/local/bin/jurnal 2>/dev/null || true
```

### Manual Installation

Download target binaries directly from the [GitHub Releases](https://github.com/luckywirasakti/jurnal/releases) page:
- **Linux**: `jurnal-linux` (amd64)
- **macOS**: `jurnal-macos` (amd64/arm64)
- **Windows**: `jurnal-windows.exe` (amd64)

Move the downloaded binary to a folder in your system `$PATH` and rename it to `jurnal`.

---

## Setup & Configuration

Configure Jurnal by running the interactive initialization command:

```bash
jurnal setup
```

This will prompt you for your API credentials and Git settings, saving them to `~/.jurnal.json`:
```json
{
    "openai_base_url": "https://api.openai.com/v1",
    "openai_api_key": "your-api-key",
    "openai_model": "gpt-4o-mini"
}
```

### Environment Variables
Environment variables take precedence over the local JSON configuration file if defined:
```bash
export OPENAI_BASE_URL="https://api.your-provider.com/v1"
export OPENAI_API_KEY="your-api-key"
export OPENAI_MODEL="gpt-4o-mini"
```

---

## Usage

### 1. `jurnal stage`
Analyze modifications in your current repository workspace, generate proposed branch naming structures, and interactively staging and committing in modular batches.

```bash
jurnal stage
```

### 2. `jurnal setup`
Set up or override local JSON credentials and global git settings.

```bash
jurnal setup
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
