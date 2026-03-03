package main

import (
	"fmt"
	"os"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/usage"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/version"
)

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

func run(osArgs []string) {
	if len(osArgs) < 2 {
		usage.Print()
		os.Exit(0)
	}

	if isHelpArg(osArgs[1]) {
		usage.Print()
		os.Exit(0)
	}

	if osArgs[1] == "version" {
		fmt.Printf("VM RSync v%s\n", version.Version)
		os.Exit(0)
	}

	os.Exit(0)
}
