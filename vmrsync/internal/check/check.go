package check

import (
	"fmt"

	"github.com/carlosrabelo/vmrsync/vmrsync/internal/cli"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/rsyncrun"
	"github.com/carlosrabelo/vmrsync/vmrsync/internal/sshcli"
)

// Run verifies SSH connectivity and that local/remote sync paths exist.
func Run(args []string) {
	cfg := cli.ParseSyncArgs("check", args)
	rsyncrun.SetupEnvironment(cfg)

	sshArgs := append(sshcli.BuildSSHFlags(cfg.SSHPort, cfg.SSHKey), cfg.Machine, "echo ok")
	if err := rsyncrun.ExecSSHCheck(sshArgs); err != nil {
		rsyncrun.FatalSSHError(cfg.Machine, err)
	}

	if !cfg.DryRun {
		rsyncrun.CheckRemoteDirExists(cfg)
	}

	localPath := cfg.LocalRoot
	if cfg.Folder != "" {
		localPath = fmt.Sprintf("%s/%s", cfg.LocalRoot, cfg.Folder)
	}
	rsyncrun.CheckLocalDirExists(localPath)

	if cfg.Quiet {
		return
	}

	remoteRoot := cfg.EffectiveRemoteRoot()
	if cfg.Folder != "" {
		fmt.Printf("[OK] %s: SSH reachable; local %s and remote %s/%s exist\n",
			cfg.Machine, localPath, remoteRoot, cfg.Folder)
		return
	}
	fmt.Printf("[OK] %s: SSH reachable; local %s and remote %s exist\n",
		cfg.Machine, localPath, remoteRoot)
}
