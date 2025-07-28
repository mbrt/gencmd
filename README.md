# gencmd

[![Build](https://github.com/mbrt/gencmd/actions/workflows/build.yml/badge.svg)](https://github.com/mbrt/gencmd/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mbrt/gencmd)](https://goreportcard.com/report/github.com/mbrt/gencmd)

gencmd is an interactive command line utility to generate bash commands from a
natural language description, directly from the console.

[![asciicast](https://asciinema.org/a/QoGh9TXk3GMcyP4FmWyh2iUqH.svg)](https://asciinema.org/a/QoGh9TXk3GMcyP4FmWyh2iUqH)

Ever went to ChatGPT after struggling some time with `man awk`, or with
questions like "was it `curl` or `wget` with `-O`"? Well, save some time and ask
directly from the terminal. Think of this as the
[fzf](https://github.com/junegunn/fzf) for natural language to bash commands.

## Why?

There are many alternative tools for this task, but none did all I wanted:

* Simple to install and configure (i.e. single binary, no dependencies).
* Work interactively in the terminal, but require minimal typing.
* Fast.
* *Do not* run commands for me, but suggest alternatives.
* Paste the result directly in the terminal (without me having to do it).
* Have built-in history for both commands and prompts.
* Open source, no sign-up required, no strings attached.

This project is minimal but provides all of the above.

## Installation

Head over to the
[latest release](https://github.com/mbrt/gencmd/releases/latest), and download a
binary appropriate for your system.

Make it executable and put it somewhere in `$PATH`:

```sh
chmod a+x gencmd
sudo mv gencmd /usr/local/bin
```

If you want to set up key bindings (default is <kbd>Ctrl</kbd> + <kbd>G</kbd>),
add this to your `.bashrc`:

```sh
source ~/.config/gencmd/key-bindings.bash
```

or use `key-bindings.zsh` for `.zshrc`.

## API Keys

Initialize with:

```sh
gencmd init
```

The instructions will guide you through setting up an AI model provider. The
currently supported providers are OpenAI, Gemini, Anthropic, and Ollama.

The easiest to get started is to get a free API key from [Google AI
Studio](https://aistudio.google.com/apikey). Follow the instructions there and
once you have the key, paste it into the interactive prompt.

For local models, you can use [Ollama](https://ollama.ai). First install Ollama
and pull a model (e.g., `ollama pull gemma-3`), then configure gencmd to use it.

Credentials are stored locally, and NEVER sent anywhere else.

> [!NOTE]
> By default, `gencmd` uses "gemini-2.0-flash-lite", which has a generous free
> tier of 200 requests per day. More than enough for typical usage. If you want
> to make sure to block requests over the free tier, use a dedicated GCP project
> without billing enabled.

> [!TIP]
> If you just want to test how `gencmd` looks without configuring it, you can
> try the demo (returning fake history and commands) with `gencmd demo`.

## Usage

Think of this as [fzf](https://github.com/junegunn/fzf) for natural language to
bash commands. Open a new terminal and press <kbd>Ctrl</kbd> + <kbd>G</kbd>.
As you type your query, `gencmd` will filter your recent history, so you can
either select something from there, or submit a new prompt.

In case the prompt is new, your configured LLM will be invoked to generate a few
alternative commands to solve your intended usage.

You can navigate history and completions with keyboard arrows <kbd>↑</kbd>
<kbd>↓</kbd>, or <kbd>Ctrl</kbd> + <kbd>J</kbd> and <kbd>Ctrl</kbd> +
<kbd>K</kbd>.

The result is *not executed*, but pasted into your command line, so that you
can edit it.

Examples for inspiration:

* Find all subdirectories
* Count the lines of a file that don't start with #
* Delete a remote git tag

## Building from source

This project is written in [Go](https://go.dev):

```sh
git clone https://github.com/mbrt/gencmd
cd gencmd
go build .
```
