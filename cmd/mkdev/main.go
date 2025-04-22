package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: mkdev <mode> <file>...
modes:
ogg
`)
	}

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	switch args[0] {
	case "ogg":
		files := args[1:]
		for _, f := range files {
			log.Println("=== " + filepath.Base(f))
			err := analyzeOggPackages(f)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
