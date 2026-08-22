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
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/excludes"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/sshcli"
)

func defaultExecSSHCheck(args []string) error {
	return sshcli.RunSSH(args)
}

func defaultExecSSH(args []string) error {
	cmd := exec.Command("ssh", args...)
	if !quietMode {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = nil
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func defaultExecRsync(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "rsync", args...)
	if !quietMode {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = nil
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// ExecSSHCheck runs a non-interactive SSH command with a 30-second timeout.
var ExecSSHCheck = defaultExecSSHCheck

// ExecSSH runs an SSH command (stdout/stderr forwarded unless quiet).
var ExecSSH = defaultExecSSH

// ExecRsync runs rsync with stdout/stderr forwarded unless quiet.
var ExecRsync = defaultExecRsync

// quietMode suppresses progress output for the current sync operation.
var quietMode bool

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

// FatalSSHError exits with a message that distinguishes SSH failures from missing paths.
func FatalSSHError(machine string, err error) {
	if err == nil {
		return
	}
	if strings.Contains(err.Error(), "timed out") {
		log.Fatalf("[ERROR] SSH connection to %s timed out. Verify network reachability and ~/.ssh/config.", machine)
	}
	code, ok := sshcli.ExitCode(err)
	if ok && code == 1 {
		log.Fatalf("[ERROR] SSH command failed on %s: %v", machine, err)
	}
	log.Fatalf("[ERROR] SSH connection to %s failed: %v\n[HINT] Verify key auth and run: ssh -o BatchMode=yes %s",
		machine, err, machine)
}

func fatalRemoteDirMissing(cfg *config.AppConfig, remoteRoot string, err error) {
	code, ok := sshcli.ExitCode(err)
	if !ok || code != 1 {
		FatalSSHError(cfg.Machine, err)
	}
	if cfg.Staging {
		log.Fatalf("[ERROR] directory %s does not exist on %s. Run: vmrsync prepare %s",
			remoteRoot, cfg.Machine, cfg.Machine)
	}
	log.Fatalf("[ERROR] directory %s does not exist on %s", remoteRoot, cfg.Machine)
}

// CheckRemoteDirExists verifies the remote root is present (non-interactive SSH).
func CheckRemoteDirExists(cfg *config.AppConfig) {
	remoteRoot := cfg.EffectiveRemoteRoot()
	args := append(sshcli.BuildSSHFlags(cfg.SSHPort, cfg.SSHKey),
		cfg.Machine,
		sshcli.TestDirCommand(remoteRoot),
	)
	if err := ExecSSHCheck(args); err != nil {
		fatalRemoteDirMissing(cfg, remoteRoot, err)
	}
}

func mergeExcludePatterns(localPath string, cfg *config.AppConfig) []string {
	var merged []string

	if patterns, err := excludes.LoadFile(filepath.Join(localPath, ".vmrsyncignore")); err != nil {
		log.Fatalf("[ERROR] failed to read .vmrsyncignore: %v", err)
	} else {
		merged = append(merged, patterns...)
	}

	merged = append(merged, cfg.Excludes...)
	return merged
}

// RunSync performs the rsync for in/out.
func RunSync(cfg *config.AppConfig) {
	quietMode = cfg.Quiet

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
	if cfg.Command == "restore" {
		src = remotePathURL + "/"
		dest = localPath + "/"
		direction = "RESTORE (Remote -> Local)"
	} else {
		src = localPath + "/"
		dest = remotePathURL + "/"
		direction = "BACKUP (Local -> Remote)"
	}

	sshFlags := sshcli.BuildSSHFlags(cfg.SSHPort, cfg.SSHKey)
	allExcludes := mergeExcludePatterns(localPath, cfg)

	args := []string{"-az", "--protect-args"}
	if cfg.Quiet {
		args = append(args, "--info=none")
	} else {
		args = append(args, "--info=progress2")
	}
	if rsyncSupportsMkpath() {
		args = append(args, "--mkpath")
	} else if cfg.Command == "backup" && !cfg.DryRun {
		sshArgs := append(sshcli.BuildSSHFlags(cfg.SSHPort, cfg.SSHKey),
			cfg.Machine,
			sshcli.MkdirPCommand(remotePath),
		)
		if err := ExecSSH(sshArgs); err != nil {
			FatalSSHError(cfg.Machine, err)
		}
	}
	if cfg.Verbose {
		args = append(args, "-v")
	}
	if !cfg.NoDelete {
		args = append(args, "--delete")
	}
	args = append(args, "-e", sshcli.FormatRsyncSSHShell(sshFlags))
	for _, excl := range allExcludes {
		args = append(args, fmt.Sprintf("--exclude=%s", excl))
	}
	args = append(args, src, dest)

	if cfg.DryRun {
		fmt.Printf("rsync %s\n", strings.Join(args, " "))
		return
	}

	if !cfg.Quiet {
		fmt.Printf("Syncing (%s):\n", direction)
		fmt.Printf("  From: %s\n", src)
		fmt.Printf("  To:   %s\n", dest)
		if len(sshFlags) > 0 {
			fmt.Printf("  SSH Options: %s\n", strings.Join(sshFlags, " "))
		}
		fmt.Println()
	}

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

	if !cfg.Quiet {
		fmt.Println("\n[INFO] Sync completed successfully!")
	}
}

// CheckLocalDirExists exits if path is missing (for tests and sync).
func CheckLocalDirExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("[ERROR] local directory does not exist: %s", path)
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return fallback
}

// RestoreExecDefaults resets injectable executables to production behavior.
func RestoreExecDefaults() {
	ExecSSHCheck = defaultExecSSHCheck
	ExecSSH = defaultExecSSH
	ExecRsync = defaultExecRsync
	quietMode = false
}

// RestoreRsyncProbe resets mkpath capability detection (for tests).
func RestoreRsyncProbe() {
	MkpathProbeOnce = sync.Once{}
	RsyncHelpOutput = func() ([]byte, error) {
		return exec.Command("rsync", "--help").CombinedOutput()
	}
}
