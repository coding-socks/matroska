package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/coding-socks/matroska/cmd/mkc/internal/cli"
	"github.com/coding-socks/matroska/cmd/mkc/internal/extract"
	"github.com/coding-socks/matroska/cmd/mkc/internal/list"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var commands = []*cli.Command{
	extract.Cmd,
	list.Cmd,
}

func main() {
	flag.Parse()

	mode := flag.Arg(0)
	signal.Ignore(syscall.SIGPIPE)

	if mode == "" {
		options := make([]huh.Option[string], len(commands))
		for i := range commands {
			options[i] = huh.NewOption(commands[i].Flags.Name(), commands[i].Flags.Name())
		}
		err := huh.NewSelect[string]().
			Title("Choose a mode:").
			Options(options...).
			Value(&mode).
			Run()
		if errors.Is(err, huh.ErrUserAborted) {
			return
		}
		if err != nil {
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
