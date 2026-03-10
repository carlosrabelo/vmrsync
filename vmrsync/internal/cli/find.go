package cli

// knownCommands are accepted as the first argv token after the program name.
var knownCommands = map[string]bool{
	"restore": true, "backup": true, "prepare": true, "check": true,
	"version": true, "help": true, "completion": true,
}

// FindCommand requires the command as the first argument.
// Returns ("", args) when args is empty or the first token is not a command.
func FindCommand(args []string) (command string, rest []string) {
	if len(args) == 0 {
		return "", args
	}
	if !knownCommands[args[0]] {
		return "", args
	}
	return args[0], args[1:]
}
