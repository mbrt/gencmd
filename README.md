# gencmd

gencmd is a simple interactive command line utility to generate bash commands
with generative AI, directly from the console.

![demo](docs/demo.gif)

Ever went to ChatGPT after struggling some time with `man awk`, or with
questions like "was it `curl` or `wget` with `-O`? Well, save some time and ask
it directly from the terminal.

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

Initialize with:

```sh
gencmd init
```

The instructions will point you to a `.env` file to edit, in order for `gencmd`
to have API access to an AI model (such as Gemini).

The easiest to get started is to get a free API key from [Google AI
Studio](https://aistudio.google.com/apikey). Follow the instructions there and
paste the key into the `.env` file initialized in the step above.

## Usage

Open a new terminal and press <kbd>Ctrl</kbd> + <kbd>G</kbd>. `gencmd` should pop
up and ask you for a prompt. This is forwarded to the LLM which will generate a
few alternative commands that should solve your intended usage.

You can navigate history and completions with keyboard arrows <kbd>↑</kbd>
<kbd>↓</kbd>, or <kbd>Ctrl</kbd> + <kbd>J</kbd> and <kbd>Ctrl</kbd> +
<kbd>K</kbd>.

Examples for inspiration:

* Find all subdirectories
* Count the lines of a file that don't start with #

## Building from source

This project is written in [Go](https://go.dev):

```sh
git clone https://github.com/mbrt/gencmd
cd gencmd
go build .
```
