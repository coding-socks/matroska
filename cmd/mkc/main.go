package main

import (
	"flag"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/coding-socks/matroska/cmd/mkc/internal/cli"
	"github.com/coding-socks/matroska/cmd/mkc/internal/extract"
	"github.com/coding-socks/matroska/cmd/mkc/internal/list"
	"log"
	"os"
)

var commands = []*cli.Command{
	extract.Cmd,
	list.Cmd,
}

func main() {
	flag.Parse()

	mode := flag.Arg(0)

	if mode == "" {
		options := make([]string, len(commands))
		for i := range commands {
			options[i] = commands[i].Flags.Name()
		}
		modeSelector := &survey.Select{
			Message: "Choose a mode:",
			Options: options,
		}
		if err := survey.AskOne(modeSelector, &mode); err != nil {
			if err == terminal.InterruptErr {
				return
			}
			log.Fatal(err)
		}
	}

	for _, cmd := range commands {
		if cmd.Flags.Name() != mode {
			continue
		}

		args := flag.Args()
		if len(args) > 0 {
			if err := cmd.Flags.Parse(args[1:]); err != nil {
				os.Exit(1)
				return
			}
		}
		cmd.Run(cmd.Flags)
		return
	}

	fmt.Printf("Unknown mode: %s\n", mode)
	os.Exit(1)
}
