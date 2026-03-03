package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/cli"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/rsyncrun"
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

	command, rest := cli.FindCommand(osArgs[1:])

	switch command {
	case "version":
		fmt.Printf("VM RSync v%s\n", version.Version)
		os.Exit(0)
	case "help":
		usage.Print()
		os.Exit(0)
	case "prepare", "completion", "check":
		log.Fatalf("%s is not implemented", command)
	case "restore", "backup":
		cfg := cli.ParseSyncArgs(command, rest)
		rsyncrun.SetupEnvironment(cfg)
		rsyncrun.RunSync(cfg)
	default:
		if len(osArgs) > 1 && !strings.HasPrefix(osArgs[1], "-") {
			log.Fatalf("Invalid command '%s'. Must be 'restore', 'backup', 'prepare', 'check', 'completion', 'version', or 'help'.", osArgs[1])
		}
		usage.Print()
		os.Exit(0)
	}
}
