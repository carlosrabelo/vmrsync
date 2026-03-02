package argv

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		wantPos   []string
		wantFlags []string
	}{
		{
			name:    "positional only",
			input:   []string{"vm21", "project1"},
			wantPos: []string{"vm21", "project1"},
		},
		{
			name:      "flag after positional",
			input:     []string{"vm21", "--dry-run"},
			wantPos:   []string{"vm21"},
			wantFlags: []string{"--dry-run"},
		},
		{
			name:      "exclude-from with value",
			input:     []string{"vm21", "--exclude-from", "patterns.txt", "--dry-run"},
			wantPos:   []string{"vm21"},
			wantFlags: []string{"--exclude-from", "patterns.txt", "--dry-run"},
		},
		{
			name:      "timeout-seconds equals form",
			input:     []string{"vm21", "--timeout-seconds=0"},
			wantPos:   []string{"vm21"},
			wantFlags: []string{"--timeout-seconds=0"},
		},
		{
			name:      "ssh-key equals form",
			input:     []string{"vm21", "--ssh-key=/tmp/id_rsa"},
			wantPos:   []string{"vm21"},
			wantFlags: []string{"--ssh-key=/tmp/id_rsa"},
		},
		{
			name:      "empty input",
			input:     []string{},
			wantPos:   nil,
			wantFlags: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, flags := SplitArgs(tt.input)
			if !reflect.DeepEqual(pos, tt.wantPos) {
				t.Errorf("positional = %v, want %v", pos, tt.wantPos)
			}
			if !reflect.DeepEqual(flags, tt.wantFlags) {
				t.Errorf("flags = %v, want %v", flags, tt.wantFlags)
			}
		})
	}
}
