package excludes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile_patternsAndComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "patterns")
	content := "# comment\n\n*.log\nnode_modules/\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"*.log", "node_modules/"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLoadFile_missing(t *testing.T) {
	got, err := LoadFile(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}
