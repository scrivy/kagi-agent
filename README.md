# kagi-agent

A terminal UI (TUI) that helps upload and download files while interacting with the [Kagi Assistant](https://assistant.kagi.com).

`kagi-agent` brings the Kagi Assistant into your terminal: optionally upload files from your current directory as context, chat with a model, and watch as it downloads the files from the response into kagi-downloads.

## Features

- **Model selection** — choose from every model available to your account.
- **File context uploads** — optionally attach files from the current directory to your question. Hidden files and the `kagi-downloads/` directory are skipped, and `.gitignore` / `.agentignore` rules (including `*` globs) are honored.

![Demo](demo.gif)

## Requirements

- A Kagi account with Assistant access

## Installation

### Compile from source
- Requires the golang compiler
```sh
git clone https://github.com/scrivy/kagi-agent.git
cd kagi-agent
make build
mv kagi-agent ~/.local/bin/kagi-agent
```

## Configuration

Settings live in the standard XDG directories:

| File | Path |
| ---- | ---- |
| Config | `$XDG_CONFIG_HOME/kagi-agent/config.yaml` (usually `~/.config/kagi-agent/config.yaml`) |

The config file is created automatically on first run and looks like this:

```yaml
loglevel: error
sessiontoken: "<your Kagi session token>"
alwaysaskformodel: true
alwaysaskforreasoningeffort: true
```

## Disclaimer

An LLM from kagi-agent produced about 5% of this code and most of this README.md.

This project uses Kagi's private, undocumented Assistant API. It is not affiliated with or endorsed by Kagi, and the API may change at any time without notice. Use at your own risk.
