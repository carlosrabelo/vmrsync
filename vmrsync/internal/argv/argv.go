package argv

import (
	"fmt"
	"strings"
)

// ValueFlags lists flags that consume the next argument as their value.
var ValueFlags = map[string]bool{
	"exclude": true, "exclude-from": true, "ssh-port": true, "ssh-key": true, "timeout-seconds": true,
}

// SplitArgs separates positional arguments from flags in fixed order:
// <machine> [<folder>] [options]. Positionals must come first; a non-flag
// token after options have started is an error.
func SplitArgs(args []string) (positional []string, flags []string, err error) {
	i := 0
	for i < len(args) && !strings.HasPrefix(args[i], "-") {
		positional = append(positional, args[i])
		i++
	}
	for i < len(args) {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			return nil, nil, fmt.Errorf("unexpected argument %q after options; expected order: <command> <machine> [<folder>] [options]", arg)
		}
		name := strings.TrimLeft(arg, "-")
		if eq := strings.Index(name, "="); eq >= 0 {
			name = name[:eq]
		}
		flags = append(flags, arg)
		if ValueFlags[name] && !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			i++
			flags = append(flags, args[i])
		}
		i++
	}
	return positional, flags, nil
}
