package rsyncrun

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/config"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/sshcli"
)

func newSSHCheckContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func defaultExecSSHCheck(args []string) error {
	ctx, cancel := newSSHCheckContext()
	defer cancel()
	return exec.CommandContext(ctx, "ssh", args...).Run()
}

func defaultExecSSH(args []string) error {
	cmd := exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func defaultExecRsync(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ExecSSHCheck runs a non-interactive SSH command with a 30-second timeout.
var ExecSSHCheck = defaultExecSSHCheck

// ExecSSH runs an interactive SSH command (stdout/stderr forwarded).
var ExecSSH = defaultExecSSH

// ExecRsync runs rsync with stdout/stderr forwarded.
var ExecRsync = defaultExecRsync

// RsyncHelpOutput returns `rsync --help` output (injectable for tests).
var RsyncHelpOutput = func() ([]byte, error) {
	return exec.Command("rsync", "--help").CombinedOutput()
}

var (
	// MkpathProbeOnce serializes detection of rsync --mkpath support.
	MkpathProbeOnce      sync.Once
	rsyncMkpathSupported bool
)

func rsyncSupportsMkpath() bool {
	MkpathProbeOnce.Do(func() {
		out, err := RsyncHelpOutput()
		if err != nil {
			rsyncMkpathSupported = false
			return
		}
		rsyncMkpathSupported = strings.Contains(string(out), "--mkpath")
	})
	return rsyncMkpathSupported
}

func rsyncHints(exitCode int) string {
	switch exitCode {
	case 23:
		return "rsync exit 23: partial transfer (often permissions/owner/group/attrs). If this is expected, consider using --no-delete or adjusting permissions."
	case 24:
		return "rsync exit 24: partial transfer due to vanished source files (common on active trees)."
	case 30:
		return "rsync exit 30: timeout in data send/receive."
	default:
		return ""
	}
}

// SetupEnvironment sets LocalRoot from VMRSYNC_PATH or $HOME/Sources.
func SetupEnvironment(cfg *config.AppConfig) {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("failed to retrieve current user: %v", err)
	}

	cfg.LocalRoot = getEnvOrDefault("VMRSYNC_PATH", filepath.Join(currentUser.HomeDir, "Sources"))
}

// CheckRemoteDirExists verifies the remote root is present (non-interactive SSH).
func CheckRemoteDirExists(cfg *config.AppConfig) {
	remoteRoot := cfg.EffectiveRemoteRoot()
	args := append(sshcli.BuildSSHFlags(cfg.SSHPort, cfg.SSHKey),
		cfg.Machine,
		sshcli.TestDirCommand(remoteRoot),
	)
	if err := ExecSSHCheck(args); err != nil {
		if cfg.Staging {
			log.Fatalf("[ERROR] directory %s does not exist on %s. Run: vmrsync setup %s", remoteRoot, cfg.Machine, cfg.Machine)
		}
		log.Fatalf("[ERROR] directory %s does not exist on %s", remoteRoot, cfg.Machine)
	}
}

// RunSync performs the rsync for in/out.
func RunSync(cfg *config.AppConfig) {
	if !cfg.DryRun {
		CheckRemoteDirExists(cfg)
	}

	remoteRoot := cfg.EffectiveRemoteRoot()
	localPath := cfg.LocalRoot
	remotePath := remoteRoot

	if cfg.Folder != "" {
		localPath = filepath.Join(localPath, cfg.Folder)
		remotePath = fmt.Sprintf("%s/%s", remotePath, cfg.Folder)
	}

	remotePathURL := fmt.Sprintf("%s:%s", cfg.Machine, remotePath)

	CheckLocalDirExists(localPath)

	var src, dest, direction string
	if cfg.Command == "in" {
		src = remotePathURL + "/"
		dest = localPath + "/"
		direction = "IN (Remote -> Local)"
	} else {
		src = localPath + "/"
		dest = remotePathURL + "/"
		direction = "OUT (Local -> Remote)"
	}

	sshFlags := sshcli.BuildSSHFlags(cfg.SSHPort, cfg.SSHKey)

	args := []string{"-az", "--info=progress2", "--protect-args"}
	if rsyncSupportsMkpath() {
		args = append(args, "--mkpath")
	} else if cfg.Command == "out" && !cfg.DryRun {
		sshArgs := append(sshcli.BuildSSHFlags(cfg.SSHPort, cfg.SSHKey),
			cfg.Machine,
			sshcli.MkdirPCommand(remotePath),
		)
		if err := ExecSSH(sshArgs); err != nil {
			log.Fatalf("[ERROR] failed to create remote directory %s on %s: %v", remotePath, cfg.Machine, err)
		}
	}
	if cfg.Verbose {
		args = append(args, "-v")
	}
	if !cfg.NoDelete {
		args = append(args, "--delete")
	}
	if len(sshFlags) > 0 {
		args = append(args, "-e", "ssh "+strings.Join(sshFlags, " "))
	}
	for _, excl := range cfg.Excludes {
		args = append(args, fmt.Sprintf("--exclude=%s", excl))
	}
	args = append(args, src, dest)

	if cfg.DryRun {
		fmt.Printf("rsync %s\n", strings.Join(args, " "))
		return
	}

	fmt.Printf("Syncing (%s):\n", direction)
	fmt.Printf("  From: %s\n", src)
	fmt.Printf("  To:   %s\n", dest)
	if len(sshFlags) > 0 {
		fmt.Printf("  SSH Options: %s\n", strings.Join(sshFlags, " "))
	}
	fmt.Println()

	rsyncCtx := context.Background()
	if cfg.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		rsyncCtx, cancel = context.WithTimeout(rsyncCtx, time.Duration(cfg.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	if err := ExecRsync(rsyncCtx, args); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code := ee.ExitCode()
			if hint := rsyncHints(code); hint != "" {
				log.Fatalf("\n[ERROR] sync failed (exit %d): %v\n[HINT] %s", code, err, hint)
			}
			log.Fatalf("\n[ERROR] sync failed (exit %d): %v", code, err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("\n[ERROR] sync failed: timed out after %ds", cfg.TimeoutSeconds)
		}
		log.Fatalf("\n[ERROR] sync failed: %v", err)
	}

	fmt.Println("\n[INFO] Sync completed successfully!")
}

// CheckLocalDirExists exits if path is missing (for tests and sync).
func CheckLocalDirExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("[ERROR] local directory does not exist: %s", path)
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// RestoreExecDefaults resets injectable executables to production behavior.
func RestoreExecDefaults() {
	ExecSSHCheck = defaultExecSSHCheck
	ExecSSH = defaultExecSSH
	ExecRsync = defaultExecRsync
}

// RestoreRsyncProbe resets mkpath capability detection (for tests).
func RestoreRsyncProbe() {
	MkpathProbeOnce = sync.Once{}
	RsyncHelpOutput = func() ([]byte, error) {
		return exec.Command("rsync", "--help").CombinedOutput()
	}
}
