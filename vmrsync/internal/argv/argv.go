package argv

import "strings"

// ValueFlags lists flags that consume the next argument as their value.
var ValueFlags = map[string]bool{
	"exclude": true, "ssh-port": true, "ssh-key": true, "timeout-seconds": true,
}

// SplitArgs separates positional arguments from flags, supporting any order.
// Flags that take a value consume the next token.
func SplitArgs(args []string) (positional []string, flags []string) {
	i := 0
	for i < len(args) {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			i++
			continue
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
	return
}
