package cli

// FindCommand scans args for the first known command name, ignoring flags.
func FindCommand(args []string) (command string, rest []string) {
	for i, arg := range args {
		switch arg {
		case "in", "out", "setup", "version", "-h", "--help":
			return arg, append(args[:i:i], args[i+1:]...)
		}
	}
	return "", args
}
