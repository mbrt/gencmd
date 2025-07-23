package ui

import (
	"time"

	"github.com/mbrt/gencmd/ctrl"
)

func NewFakeController() *FakeController {
	return &FakeController{
		history: []ctrl.HistoryEntry{
			{
				Prompt:  "list files",
				Command: "ls -l",
			},
			{
				Prompt:  "print the third element of a comma separated string",
				Command: "awk -F, '{print $3}'",
			},
			{
				Prompt:  "print the last element of a json array",
				Command: `head -1 <<< "$(jq -c '.[-1]' file.json)"`,
			},
			{
				Prompt:  "find all subdirectories",
				Command: "find . -type d",
			},
			{
				Prompt:  "find the first 3 files in a directory",
				Command: "ls -f | head -n 3",
			},
			{
				Prompt:  "return the second column of a csv",
				Command: `awk -F, '{print $2}'`,
			},
			{
				Prompt:  "delete all .bak files in subdirectories",
				Command: `find . -name "*.bak" -delete`,
			},
			{
				Prompt:  "rename all .jpg files into .jpeg",
				Command: `find . -name "*.jpg" -print0 | while IFS= read -r -d $'\n' file; do mv "$file" "${file%.jpg}.jpeg"; done`,
			},
			{
				Prompt:  "kill all processes of a user",
				Command: "pkill -u <username>",
			},
			{
				Prompt:  "find go files in subdirectories",
				Command: `find . -type f -name "*.go"`,
			},
			{
				Prompt:  "list images in current directory",
				Command: `find . -type f -print0 | grep -zE '\\.(gif|png|jpg|jpeg)$'`,
			},
		},
		commands: []string{
			`find . -name *.jpg`,
			`find . -type f -name *.jpg`,
			`find ./ -name "*.jpg"`,
			`find . -iname *.jpg`,
			`find ./ -type f -iname "*.jpg"`,
		},
		generateDelay: 2 * time.Second,
	}
}

type FakeController struct {
	history       []ctrl.HistoryEntry
	commands      []string
	generateDelay time.Duration
	generateErr   error
}

func (f *FakeController) LoadHistory() []ctrl.HistoryEntry {
	return f.history
}

func (f *FakeController) UpdateHistory(prompt, command string) error {
	f.history = append(f.history, ctrl.HistoryEntry{
		Prompt:  prompt,
		Command: command,
	})
	return nil
}

func (f *FakeController) GenerateCommands(string) ([]string, error) {
	time.Sleep(f.generateDelay) // Simulate a delay
	return f.commands, f.generateErr
}
