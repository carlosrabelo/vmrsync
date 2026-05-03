package setup

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

// Run runs the setup subcommand.
func Run(args []string) {
	fs := flag.NewFlagSet("vmrsync setup", flag.ExitOnError)
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
	if uid < 1000 {
		log.Fatalf("local UID %d is less than 1000; refusing to set remote ownership to avoid overwriting system-owned files", uid)
	}

	machine := positional[0]
	if err := hostcheck.EnsureRemoteSSHHost(machine); err != nil {
		log.Fatal(err)
	}
	remoteRoot := config.VmrsyncRoot
	remoteCmd := fmt.Sprintf("sudo mkdir -p %s && sudo chown %d:%d %s", remoteRoot, uid, uid, remoteRoot)

	sshArgs := append(sshcli.BuildSSHFlags(sshPort, sshKey), machine, remoteCmd)

	if dryRun {
		fmt.Printf("ssh %s\n", strings.Join(sshArgs, " "))
		return
	}

	fmt.Printf("Setting up %s on %s...\n", remoteRoot, machine)
	fmt.Printf("  Running: ssh %s\n\n", strings.Join(sshArgs, " "))

	if err := rsyncrun.ExecSSH(sshArgs); err != nil {
		log.Fatalf("[ERROR] setup failed: %v", err)
	}

	fmt.Printf("\n[INFO] setup complete. %s is ready on %s (owner: UID %d)\n", remoteRoot, machine, uid)
}
