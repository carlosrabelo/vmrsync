package argv

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		wantPos   []string
		wantFlags []string
		wantErr   string
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
			name:      "folder then options",
			input:     []string{"vm21", "project1", "--dry-run"},
			wantPos:   []string{"vm21", "project1"},
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
			name:    "empty input",
			input:   []string{},
			wantPos: nil,
		},
		{
			name:    "positional after options",
			input:   []string{"vm21", "--dry-run", "project1"},
			wantErr: "unexpected argument",
		},
		{
			name:    "options before machine",
			input:   []string{"--dry-run", "vm21"},
			wantErr: "unexpected argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, flags, err := SplitArgs(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(pos, tt.wantPos) {
				t.Errorf("positional = %v, want %v", pos, tt.wantPos)
			}
			if !reflect.DeepEqual(flags, tt.wantFlags) {
				t.Errorf("flags = %v, want %v", flags, tt.wantFlags)
			}
		})
	}
}
