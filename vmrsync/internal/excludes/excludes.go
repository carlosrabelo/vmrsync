package excludes

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadFile reads exclude patterns from path (one per line; # comments and blanks skipped).
func LoadFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read exclude file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var patterns []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read exclude file %s: %w", path, err)
	}
	return patterns, nil
}
