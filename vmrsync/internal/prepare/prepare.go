package prepare

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/argv"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/config"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/hostcheck"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/rsyncrun"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/sshcli"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/usage"
)

// Run runs the prepare subcommand.
func Run(args []string) {
	fs := flag.NewFlagSet("vmrsync prepare", flag.ExitOnError)
	var sshPort string
	var sshKey string
	var dryRun bool
	fs.StringVar(&sshPort, "ssh-port", "", "SSH port")
	fs.StringVar(&sshKey, "ssh-key", "", "SSH private key path")
	fs.BoolVar(&dryRun, "dry-run", false, "Print the SSH command that would be executed, without running it")
	fs.Usage = usage.Print

	positional, flagTokens := argv.SplitArgs(args)

	if err := fs.Parse(flagTokens); err != nil {
		log.Fatal(err)
	}

	if len(positional) == 0 {
		log.Fatal("machine not specified")
	}
	if len(positional) > 1 {
		log.Fatalf("too many arguments: %v", positional)
	}

	uid := os.Getuid()
	gid := os.Getgid()
	if uid < 1000 && os.Getenv("VMRSYNC_SKIP_UID_CHECK") == "" {
		log.Fatalf("local UID %d is less than 1000; refusing to set remote ownership to avoid overwriting system-owned files", uid)
	}

	machine := positional[0]
	if err := hostcheck.EnsureRemoteSSHHost(machine); err != nil {
		log.Fatal(err)
	}
	remoteRoot := config.VmrsyncRoot
	remoteCmd := fmt.Sprintf("sudo mkdir -p %s && sudo chown %d:%d %s", remoteRoot, uid, gid, remoteRoot)

	sshArgs := append(sshcli.BuildSSHFlags(sshPort, sshKey), machine, remoteCmd)

	if dryRun {
		fmt.Printf("ssh %s\n", strings.Join(sshArgs, " "))
		return
	}

	fmt.Printf("Preparing %s on %s...\n", remoteRoot, machine)
	fmt.Printf("  Running: ssh %s\n\n", strings.Join(sshArgs, " "))

	if err := rsyncrun.ExecSSH(sshArgs); err != nil {
		log.Fatalf("[ERROR] prepare failed: %v", err)
	}

	fmt.Printf("\n[INFO] prepare complete. %s is ready on %s (owner: UID %d GID %d)\n", remoteRoot, machine, uid, gid)
}
