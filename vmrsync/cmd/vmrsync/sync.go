package main

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
)

// newSSHCheckContext returns a context with a 30-second timeout for SSH connectivity checks.
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

// execSSHCheck runs a non-interactive SSH command with a 30-second timeout.
// Injectable for testing.
var execSSHCheck = defaultExecSSHCheck

// execSSH runs an interactive SSH command (stdout/stderr forwarded).
// Injectable for testing.
var execSSH = defaultExecSSH

// execRsync runs rsync with stdout/stderr forwarded.
// Injectable for testing.
var execRsync = defaultExecRsync

// rsyncHelpOutput returns `rsync --help` output. Injectable for tests.
var rsyncHelpOutput = func() ([]byte, error) {
	return exec.Command("rsync", "--help").CombinedOutput()
}

var (
	rsyncMkpathOnce      sync.Once
	rsyncMkpathSupported bool
)

func rsyncSupportsMkpath() bool {
	rsyncMkpathOnce.Do(func() {
		out, err := rsyncHelpOutput()
		if err != nil {
			// Conservative: if we cannot determine support, assume it's unavailable and
			// use the mkdir fallback for OUT syncs.
			rsyncMkpathSupported = false
			return
		}
		rsyncMkpathSupported = strings.Contains(string(out), "--mkpath")
	})
	return rsyncMkpathSupported
}

func shellQuoteSingle(s string) string {
	// POSIX shell single-quote escaping:  abc'd  ->  'abc'"'"'d'
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func sshTestDirCommand(path string) string {
	return "test -d -- " + shellQuoteSingle(path)
}

func sshMkdirPCommand(path string) string {
	return "mkdir -p -- " + shellQuoteSingle(path)
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

// buildSSHFlags returns SSH flag args for -p and -i options.
func buildSSHFlags(port, key string) []string {
	var args []string
	if port != "" {
		args = append(args, "-p", port)
	}
	if key != "" {
		args = append(args, "-i", key)
	}
	return args
}

func setupEnvironment(config *AppConfig) {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("failed to retrieve current user: %v", err)
	}

	config.LocalRoot = getEnvOrDefault("VMRSYNC_PATH", filepath.Join(currentUser.HomeDir, "Sources"))
}

func checkRemoteDirExists(config *AppConfig) {
	remoteRoot := config.effectiveRemoteRoot()
	args := append(buildSSHFlags(config.SSHPort, config.SSHKey),
		config.Machine,
		sshTestDirCommand(remoteRoot),
	)
	if err := execSSHCheck(args); err != nil {
		if config.Staging {
			log.Fatalf("[ERROR] directory %s does not exist on %s. Run: vmrsync setup %s", remoteRoot, config.Machine, config.Machine)
		}
		log.Fatalf("[ERROR] directory %s does not exist on %s", remoteRoot, config.Machine)
	}
}

func runSync(config *AppConfig) {
	if !config.DryRun {
		checkRemoteDirExists(config)
	}

	remoteRoot := config.effectiveRemoteRoot()
	localPath := config.LocalRoot
	remotePath := remoteRoot

	if config.Folder != "" {
		localPath = filepath.Join(localPath, config.Folder)
		remotePath = fmt.Sprintf("%s/%s", remotePath, config.Folder)
	}

	remotePathURL := fmt.Sprintf("%s:%s", config.Machine, remotePath)

	checkLocalDirExists(localPath)

	var src, dest, direction string
	if config.Command == "in" {
		src = remotePathURL + "/"
		dest = localPath + "/"
		direction = "IN (Remote -> Local)"
	} else {
		src = localPath + "/"
		dest = remotePathURL + "/"
		direction = "OUT (Local -> Remote)"
	}

	sshFlags := buildSSHFlags(config.SSHPort, config.SSHKey)

	args := []string{"-az", "--info=progress2", "--protect-args"}
	if rsyncSupportsMkpath() {
		args = append(args, "--mkpath")
	} else if config.Command == "out" && !config.DryRun {
		// Fallback for older rsync without --mkpath: create remote path via SSH.
		sshArgs := append(buildSSHFlags(config.SSHPort, config.SSHKey),
			config.Machine,
			sshMkdirPCommand(remotePath),
		)
		if err := execSSH(sshArgs); err != nil {
			log.Fatalf("[ERROR] failed to create remote directory %s on %s: %v", remotePath, config.Machine, err)
		}
	}
	if config.Verbose {
		args = append(args, "-v")
	}
	if !config.NoDelete {
		args = append(args, "--delete")
	}
	if len(sshFlags) > 0 {
		args = append(args, "-e", "ssh "+strings.Join(sshFlags, " "))
	}
	for _, excl := range config.Excludes {
		args = append(args, fmt.Sprintf("--exclude=%s", excl))
	}
	args = append(args, src, dest)

	if config.DryRun {
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
	if config.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		rsyncCtx, cancel = context.WithTimeout(rsyncCtx, time.Duration(config.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	if err := execRsync(rsyncCtx, args); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code := ee.ExitCode()
			if hint := rsyncHints(code); hint != "" {
				log.Fatalf("\n[ERROR] sync failed (exit %d): %v\n[HINT] %s", code, err, hint)
			}
			log.Fatalf("\n[ERROR] sync failed (exit %d): %v", code, err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("\n[ERROR] sync failed: timed out after %ds", config.TimeoutSeconds)
		}
		log.Fatalf("\n[ERROR] sync failed: %v", err)
	}

	fmt.Println("\n[INFO] Sync completed successfully!")
}

func checkLocalDirExists(path string) {
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
