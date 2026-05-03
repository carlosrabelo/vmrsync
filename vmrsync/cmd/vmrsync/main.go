package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/cli"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/rsyncrun"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/setup"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/usage"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage.Print()
		os.Exit(1)
	}

	command, rest := cli.FindCommand(os.Args[1:])

	switch command {
	case "version":
		fmt.Printf("VM RSync v%s\n", version.Version)
		os.Exit(0)
	case "-h", "--help":
		usage.Print()
		os.Exit(0)
	case "setup":
		setup.Run(rest)
		os.Exit(0)
	case "in", "out":
		cfg := cli.ParseSyncArgs(command, rest)
		rsyncrun.SetupEnvironment(cfg)
		rsyncrun.RunSync(cfg)
	default:
		if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
			log.Fatalf("Invalid command '%s'. Must be 'in', 'out', or 'setup'.", os.Args[1])
		}
		usage.Print()
		os.Exit(1)
	}
}
