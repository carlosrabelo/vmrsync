package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// findCommand scans args for the first known command name, ignoring flags.
// Returns the command and the remaining args (without the command element).
func findCommand(args []string) (command string, rest []string) {
	for i, arg := range args {
		switch arg {
		case "in", "out", "setup", "version", "-h", "--help":
			return arg, append(args[:i:i], args[i+1:]...)
		}
	}
	return "", args
}

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	command, rest := findCommand(os.Args[1:])

	switch command {
	case "version":
		fmt.Printf("VM RSync v%s\n", VERSION)
		os.Exit(0)
	case "-h", "--help":
		showUsage()
		os.Exit(0)
	case "setup":
		runSetup(rest)
		os.Exit(0)
	case "in", "out":
		config := parseArgs(command, rest)
		setupEnvironment(config)
		runSync(config)
	default:
		// Show the unrecognized token for better diagnostics.
		if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
			log.Fatalf("Invalid command '%s'. Must be 'in', 'out', or 'setup'.", os.Args[1])
		}
		showUsage()
		os.Exit(1)
	}
}
