package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/argv"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/cli"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/config"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/rsyncrun"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/setup"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/sshcli"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/usage"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "vmrsync-test")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmp)

	binaryPath = filepath.Join(tmp, "vmrsync")
	out, err := exec.Command("go", "build", "-o", binaryPath, ".").CombinedOutput()
	if err != nil {
		panic("failed to build binary: " + string(out))
	}

	os.Exit(m.Run())
}

func runBinary(args []string, extraEnv ...string) (string, int) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), extraEnv...)
	out, _ := cmd.CombinedOutput()
	return string(out), cmd.ProcessState.ExitCode()
}

func TestVersion(t *testing.T) {
	_, code := runBinary([]string{"version"})
	if code != 0 {
		t.Errorf("exit code %d, want 0", code)
	}
}

func TestHelp(t *testing.T) {
	_, code := runBinary([]string{"--help"})
	if code != 0 {
		t.Errorf("exit code %d, want 0", code)
	}
}

func TestInvalidCommand(t *testing.T) {
	_, code := runBinary([]string{"invalid", "vm21"})
	if code == 0 {
		t.Error("expected non-zero exit code for invalid command")
	}
}

func TestMissingMachine(t *testing.T) {
	_, code := runBinary([]string{"in"})
	if code == 0 {
		t.Error("expected non-zero exit code when machine is missing")
	}
}

func TestDryRun(t *testing.T) {
	tmpSrc := filepath.Join(t.TempDir(), "sources")
	os.MkdirAll(filepath.Join(tmpSrc, "project1"), 0755)
	os.MkdirAll(filepath.Join(tmpSrc, "71"), 0755)
	srcEnv := "VMRSYNC_PATH=" + tmpSrc

	tests := []struct {
		name       string
		args       []string
		env        []string
		wantOut    string
		wantNotOut string
		wantExit   int
	}{
		{
			name:    "dry-run flag before command",
			args:    []string{"--dry-run", "in", "vm21", "project1"},
			env:     []string{srcEnv},
			wantOut: "rsync -az",
		},
		{
			name:    "dry-run shows rsync command",
			args:    []string{"in", "vm21", "project1", "--dry-run"},
			env:     []string{srcEnv},
			wantOut: "rsync -az",
		},
		{
			name:    "ssh port passed to rsync",
			args:    []string{"in", "vm21", "project1", "--dry-run", "--ssh-port", "2222"},
			env:     []string{srcEnv},
			wantOut: "-p 2222",
		},
		{
			name:    "ssh key passed to rsync",
			args:    []string{"out", "vm21", "project1", "--dry-run", "--ssh-key", "/tmp/id_rsa"},
			env:     []string{srcEnv},
			wantOut: "-i /tmp/id_rsa",
		},
		{
			name:    "verbose adds -v flag",
			args:    []string{"in", "vm21", "project1", "--dry-run", "--verbose"},
			env:     []string{srcEnv},
			wantOut: " -v ",
		},
		{
			name:       "no-delete omits --delete",
			args:       []string{"out", "vm21", "project1", "--dry-run", "--no-delete"},
			env:        []string{srcEnv},
			wantNotOut: "--delete",
		},
		{
			name:    "default includes --delete",
			args:    []string{"out", "vm21", "project1", "--dry-run"},
			env:     []string{srcEnv},
			wantOut: "--delete",
		},
		{
			name:    "multiple excludes",
			args:    []string{"out", "vm21", "project1", "--dry-run", "--exclude", "*.log", "--exclude", "node_modules"},
			env:     []string{srcEnv},
			wantOut: "--exclude=node_modules",
		},
		{
			name:    "default mirrors local root (no folder)",
			args:    []string{"in", "vm21", "--dry-run"},
			env:     []string{srcEnv},
			wantOut: "vm21:" + tmpSrc + "/",
		},
		{
			name:    "default mirrors local root (with folder)",
			args:    []string{"in", "vm21", "project1", "--dry-run"},
			env:     []string{srcEnv},
			wantOut: "vm21:" + tmpSrc + "/project1/",
		},
		{
			name:    "staging flag uses /vmrsync",
			args:    []string{"out", "vm21", "71", "--dry-run", "--staging"},
			env:     []string{srcEnv},
			wantOut: "vm21:/vmrsync/71/",
		},
		{
			name:       "default does not use /vmrsync",
			args:       []string{"out", "vm21", "71", "--dry-run"},
			env:        []string{srcEnv},
			wantNotOut: "/vmrsync",
		},
		{
			name:    "setup dry-run",
			args:    []string{"setup", "vm21", "--dry-run"},
			wantOut: "ssh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, code := runBinary(tt.args, tt.env...)
			if code != tt.wantExit {
				t.Errorf("exit code %d, want %d\noutput: %s", code, tt.wantExit, out)
			}
			if tt.wantOut != "" && !strings.Contains(out, tt.wantOut) {
				t.Errorf("output does not contain %q\noutput: %s", tt.wantOut, out)
			}
			if tt.wantNotOut != "" && strings.Contains(out, tt.wantNotOut) {
				t.Errorf("output should not contain %q\noutput: %s", tt.wantNotOut, out)
			}
		})
	}
}

func TestEffectiveRemoteRoot(t *testing.T) {
	t.Run("staging uses /vmrsync", func(t *testing.T) {
		cfg := &config.AppConfig{Staging: true, LocalRoot: "/home/user/Sources"}
		if got := cfg.EffectiveRemoteRoot(); got != config.VmrsyncRoot {
			t.Errorf("got %q, want %q", got, config.VmrsyncRoot)
		}
	})
	t.Run("default mirrors local root", func(t *testing.T) {
		cfg := &config.AppConfig{LocalRoot: "/home/user/Sources"}
		if got := cfg.EffectiveRemoteRoot(); got != "/home/user/Sources" {
			t.Errorf("got %q, want /home/user/Sources", got)
		}
	})
}

func TestRunSyncStaging(t *testing.T) {
	var capturedArgs []string
	rsyncrun.ExecSSHCheck = func(args []string) error { return nil }
	rsyncrun.ExecRsync = func(ctx context.Context, args []string) error {
		capturedArgs = args
		return nil
	}
	defer rsyncrun.RestoreExecDefaults()

	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "71"), 0755)

	cfg := &config.AppConfig{
		Command:   "out",
		Machine:   "vm21",
		Folder:    "71",
		LocalRoot: dir,
		Staging:   true,
	}
	rsyncrun.RunSync(cfg)

	joined := strings.Join(capturedArgs, " ")
	if !strings.Contains(joined, "vm21:/vmrsync/71/") {
		t.Errorf("expected vm21:/vmrsync/71/ in args: %s", joined)
	}
}

func TestBuildSSHFlags(t *testing.T) {
	tests := []struct {
		port string
		key  string
		want []string
	}{
		{"", "", nil},
		{"2222", "", []string{"-p", "2222"}},
		{"", "/tmp/id_rsa", []string{"-i", "/tmp/id_rsa"}},
		{"2222", "/tmp/id_rsa", []string{"-p", "2222", "-i", "/tmp/id_rsa"}},
	}
	for _, tt := range tests {
		got := sshcli.BuildSSHFlags(tt.port, tt.key)
		if len(got) != len(tt.want) {
			t.Errorf("BuildSSHFlags(%q, %q) = %v, want %v", tt.port, tt.key, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("BuildSSHFlags(%q, %q)[%d] = %q, want %q", tt.port, tt.key, i, got[i], tt.want[i])
			}
		}
	}
}

func TestVmrsyncRoot(t *testing.T) {
	if config.VmrsyncRoot != "/vmrsync" {
		t.Errorf("VmrsyncRoot = %q, want /vmrsync", config.VmrsyncRoot)
	}
}

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		input     []string
		wantPos   []string
		wantFlags []string
	}{
		{[]string{"vm21", "project1"}, []string{"vm21", "project1"}, nil},
		{[]string{"vm21", "--dry-run"}, []string{"vm21"}, []string{"--dry-run"}},
		{[]string{"vm21", "project1", "--dry-run"}, []string{"vm21", "project1"}, []string{"--dry-run"}},
		{[]string{}, nil, nil},
	}
	for _, tt := range tests {
		pos, flags := argv.SplitArgs(tt.input)
		if len(pos) != len(tt.wantPos) {
			t.Errorf("SplitArgs(%v) positional = %v, want %v", tt.input, pos, tt.wantPos)
			continue
		}
		if len(flags) != len(tt.wantFlags) {
			t.Errorf("SplitArgs(%v) flags = %v, want %v", tt.input, flags, tt.wantFlags)
		}
	}
}

func TestExcludeFlags(t *testing.T) {
	var ef config.ExcludeFlags

	if ef.String() != "" {
		t.Errorf("empty String() = %q, want empty string", ef.String())
	}

	if err := ef.Set("*.log"); err != nil {
		t.Fatalf("Set(*.log) failed: %v", err)
	}
	if err := ef.Set("node_modules"); err != nil {
		t.Fatalf("Set(node_modules) failed: %v", err)
	}

	if len(ef) != 2 {
		t.Errorf("len = %d, want 2", len(ef))
	}
	if ef.String() != "*.log node_modules" {
		t.Errorf("String() = %q, want %q", ef.String(), "*.log node_modules")
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		osArgs       []string
		command      string
		wantMachine  string
		wantFolder   string
		wantDryRun   bool
		wantPort     string
		wantKey      string
		wantVerbose  bool
		wantNoDelete bool
		wantTimeout  int
	}{
		{
			osArgs:      []string{"vmrsync", "in", "vm21"},
			command:     "in",
			wantMachine: "vm21",
			wantTimeout: 7200,
		},
		{
			osArgs:      []string{"vmrsync", "in", "vm21", "proj"},
			command:     "in",
			wantMachine: "vm21",
			wantFolder:  "proj",
			wantTimeout: 7200,
		},
		{
			osArgs:      []string{"vmrsync", "in", "vm21", "--dry-run"},
			command:     "in",
			wantMachine: "vm21",
			wantDryRun:  true,
			wantTimeout: 7200,
		},
		{
			osArgs:      []string{"vmrsync", "out", "vm21", "proj", "--ssh-port", "2222"},
			command:     "out",
			wantMachine: "vm21",
			wantFolder:  "proj",
			wantPort:    "2222",
			wantTimeout: 7200,
		},
		{
			osArgs:      []string{"vmrsync", "in", "vm21", "--ssh-key", "/tmp/id_rsa"},
			command:     "in",
			wantMachine: "vm21",
			wantKey:     "/tmp/id_rsa",
			wantTimeout: 7200,
		},
		{
			osArgs:       []string{"vmrsync", "in", "vm21", "--verbose", "--no-delete"},
			command:      "in",
			wantMachine:  "vm21",
			wantVerbose:  true,
			wantNoDelete: true,
			wantTimeout:  7200,
		},
		{
			osArgs:      []string{"vmrsync", "in", "vm21", "a/b/c"},
			command:     "in",
			wantMachine: "vm21",
			wantFolder:  "a/b/c",
			wantTimeout: 7200,
		},
		{
			osArgs:      []string{"vmrsync", "out", "vm21", "--timeout-seconds", "0"},
			command:     "out",
			wantMachine: "vm21",
			wantTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.osArgs[2:], " "), func(t *testing.T) {
			cfg := cli.ParseSyncArgs(tt.command, tt.osArgs[2:])
			if cfg.Machine != tt.wantMachine {
				t.Errorf("Machine = %q, want %q", cfg.Machine, tt.wantMachine)
			}
			if cfg.Folder != tt.wantFolder {
				t.Errorf("Folder = %q, want %q", cfg.Folder, tt.wantFolder)
			}
			if cfg.DryRun != tt.wantDryRun {
				t.Errorf("DryRun = %v, want %v", cfg.DryRun, tt.wantDryRun)
			}
			if cfg.SSHPort != tt.wantPort {
				t.Errorf("SSHPort = %q, want %q", cfg.SSHPort, tt.wantPort)
			}
			if cfg.SSHKey != tt.wantKey {
				t.Errorf("SSHKey = %q, want %q", cfg.SSHKey, tt.wantKey)
			}
			if cfg.Verbose != tt.wantVerbose {
				t.Errorf("Verbose = %v, want %v", cfg.Verbose, tt.wantVerbose)
			}
			if cfg.NoDelete != tt.wantNoDelete {
				t.Errorf("NoDelete = %v, want %v", cfg.NoDelete, tt.wantNoDelete)
			}
			if cfg.TimeoutSeconds != tt.wantTimeout {
				t.Errorf("TimeoutSeconds = %v, want %v", cfg.TimeoutSeconds, tt.wantTimeout)
			}
		})
	}
}

func TestSetupEnvironment(t *testing.T) {
	t.Run("default uses home/Sources", func(t *testing.T) {
		os.Unsetenv("VMRSYNC_PATH")
		cfg := &config.AppConfig{}
		rsyncrun.SetupEnvironment(cfg)
		if cfg.LocalRoot == "" {
			t.Error("LocalRoot should not be empty")
		}
		if !strings.HasSuffix(cfg.LocalRoot, "Sources") {
			t.Errorf("LocalRoot %q should end with 'Sources'", cfg.LocalRoot)
		}
	})
	t.Run("override via env", func(t *testing.T) {
		os.Setenv("VMRSYNC_PATH", "/custom/path")
		defer os.Unsetenv("VMRSYNC_PATH")
		cfg := &config.AppConfig{}
		rsyncrun.SetupEnvironment(cfg)
		if cfg.LocalRoot != "/custom/path" {
			t.Errorf("LocalRoot = %q, want /custom/path", cfg.LocalRoot)
		}
	})
}

func TestCheckRemoteDirExists(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var capturedArgs []string
		rsyncrun.ExecSSHCheck = func(args []string) error {
			capturedArgs = args
			return nil
		}
		defer rsyncrun.RestoreExecDefaults()

		cfg := &config.AppConfig{Machine: "vm21", SSHPort: "2222", SSHKey: "/tmp/id_rsa"}
		rsyncrun.CheckRemoteDirExists(cfg)

		joined := strings.Join(capturedArgs, " ")
		for _, want := range []string{"-p 2222", "-i /tmp/id_rsa", "vm21", "test -d"} {
			if !strings.Contains(joined, want) {
				t.Errorf("expected %q in args: %v", want, capturedArgs)
			}
		}
	})
}

func TestRunSetup(t *testing.T) {
	t.Run("executes ssh with remote mkdir", func(t *testing.T) {
		var capturedArgs []string
		rsyncrun.ExecSSH = func(args []string) error {
			capturedArgs = args
			return nil
		}
		defer rsyncrun.RestoreExecDefaults()

		setup.Run([]string{"vm21"})

		joined := strings.Join(capturedArgs, " ")
		uid := fmt.Sprintf("%d", os.Getuid())
		for _, want := range []string{"vm21", "sudo mkdir -p", uid} {
			if !strings.Contains(joined, want) {
				t.Errorf("expected %q in ssh args: %s", want, joined)
			}
		}
	})

	t.Run("dry-run does not call ssh", func(t *testing.T) {
		called := false
		rsyncrun.ExecSSH = func(args []string) error {
			called = true
			return nil
		}
		defer rsyncrun.RestoreExecDefaults()

		setup.Run([]string{"vm21", "--dry-run"})

		if called {
			t.Error("ExecSSH should not be called in dry-run mode")
		}
	})

	t.Run("ssh port and key forwarded", func(t *testing.T) {
		var capturedArgs []string
		rsyncrun.ExecSSH = func(args []string) error {
			capturedArgs = args
			return nil
		}
		defer rsyncrun.RestoreExecDefaults()

		setup.Run([]string{"vm21", "--ssh-port", "2222", "--ssh-key", "/tmp/id_rsa"})

		joined := strings.Join(capturedArgs, " ")
		for _, want := range []string{"-p 2222", "-i /tmp/id_rsa"} {
			if !strings.Contains(joined, want) {
				t.Errorf("expected %q in ssh args: %s", want, joined)
			}
		}
	})
}

func TestRunSyncDryRun(t *testing.T) {
	called := false
	rsyncrun.ExecRsync = func(ctx context.Context, args []string) error {
		called = true
		return nil
	}
	defer rsyncrun.RestoreExecDefaults()

	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "proj"), 0755)

	cfg := &config.AppConfig{
		Command:   "in",
		Machine:   "vm21",
		Folder:    "proj",
		DryRun:    true,
		LocalRoot: dir,
	}
	rsyncrun.RunSync(cfg)

	if called {
		t.Error("ExecRsync should not be called in dry-run mode")
	}
}

func TestRunSyncIn(t *testing.T) {
	var capturedArgs []string
	rsyncrun.ExecSSHCheck = func(args []string) error { return nil }
	rsyncrun.ExecRsync = func(ctx context.Context, args []string) error {
		capturedArgs = args
		return nil
	}
	defer rsyncrun.RestoreExecDefaults()

	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "proj"), 0755)

	cfg := &config.AppConfig{
		Command:   "in",
		Machine:   "vm21",
		Folder:    "proj",
		LocalRoot: dir,
	}
	rsyncrun.RunSync(cfg)

	joined := strings.Join(capturedArgs, " ")
	wantRemote := fmt.Sprintf("vm21:%s/proj/", dir)
	for _, want := range []string{wantRemote, "--delete"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q in rsync args: %s", want, joined)
		}
	}
}

func TestRunSyncOut(t *testing.T) {
	var capturedArgs []string
	rsyncrun.ExecSSHCheck = func(args []string) error { return nil }
	rsyncrun.ExecRsync = func(ctx context.Context, args []string) error {
		capturedArgs = args
		return nil
	}
	defer rsyncrun.RestoreExecDefaults()

	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "proj"), 0755)

	cfg := &config.AppConfig{
		Command:   "out",
		Machine:   "vm21",
		Folder:    "proj",
		LocalRoot: dir,
	}
	rsyncrun.RunSync(cfg)

	wantRemote := fmt.Sprintf("vm21:%s/proj/", dir)
	if !strings.Contains(strings.Join(capturedArgs, " "), wantRemote) {
		t.Errorf("expected %q in rsync args: %v", wantRemote, capturedArgs)
	}
}

func TestRunSyncNoDelete(t *testing.T) {
	var capturedArgs []string
	rsyncrun.ExecSSHCheck = func(args []string) error { return nil }
	rsyncrun.ExecRsync = func(ctx context.Context, args []string) error {
		capturedArgs = args
		return nil
	}
	defer rsyncrun.RestoreExecDefaults()

	dir := t.TempDir()
	cfg := &config.AppConfig{
		Command:   "out",
		Machine:   "vm21",
		LocalRoot: dir,
		NoDelete:  true,
	}
	rsyncrun.RunSync(cfg)

	for _, arg := range capturedArgs {
		if arg == "--delete" {
			t.Errorf("--delete should not be present when NoDelete=true; args: %v", capturedArgs)
		}
	}
}

func TestRunSyncSSHOptions(t *testing.T) {
	var capturedArgs []string
	rsyncrun.ExecSSHCheck = func(args []string) error { return nil }
	rsyncrun.ExecRsync = func(ctx context.Context, args []string) error {
		capturedArgs = args
		return nil
	}
	defer rsyncrun.RestoreExecDefaults()

	dir := t.TempDir()
	cfg := &config.AppConfig{
		Command:   "in",
		Machine:   "vm21",
		LocalRoot: dir,
		SSHPort:   "2222",
		SSHKey:    "/tmp/id_rsa",
	}
	rsyncrun.RunSync(cfg)

	joined := strings.Join(capturedArgs, " ")
	if !strings.Contains(joined, "-e") || !strings.Contains(joined, "-p 2222") {
		t.Errorf("expected SSH options in rsync args: %s", joined)
	}
	if !strings.Contains(joined, "--protect-args") {
		t.Errorf("expected --protect-args in rsync args: %s", joined)
	}
}

func TestRunSyncMkpathFallbackOut(t *testing.T) {
	rsyncrun.MkpathProbeOnce = sync.Once{}
	rsyncrun.RsyncHelpOutput = func() ([]byte, error) { return []byte("rsync help without mkpath"), nil }

	var sshCalled bool
	rsyncrun.ExecSSHCheck = func(args []string) error { return nil }
	rsyncrun.ExecSSH = func(args []string) error {
		sshCalled = true
		return nil
	}
	rsyncrun.ExecRsync = func(ctx context.Context, args []string) error { return nil }
	defer func() {
		rsyncrun.RestoreExecDefaults()
		rsyncrun.RestoreRsyncProbe()
	}()

	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "proj"), 0755)

	cfg := &config.AppConfig{
		Command:   "out",
		Machine:   "vm21",
		Folder:    "proj",
		LocalRoot: dir,
	}
	rsyncrun.RunSync(cfg)
	if !sshCalled {
		t.Fatal("expected ExecSSH to be called to mkdir -p remote path when --mkpath unsupported")
	}
}

func TestCheckLocalDirExists_existing(t *testing.T) {
	dir := t.TempDir()
	rsyncrun.CheckLocalDirExists(dir)
}

func TestShowUsage(t *testing.T) {
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	usage.Print()

	w.Close()
	os.Stdout = orig
	r.Close()
}
