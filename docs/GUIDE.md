---
layout: default
title: User Guide
---

# VM RSync User Guide

Installation, configuration, usage, and troubleshooting for VM RSync.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Basic Usage](#basic-usage)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

## Installation

### Prerequisites

Before installing VM RSync, ensure you have:
- **Go** (version 1.26 or later)
- **rsync** (usually pre-installed on Linux/macOS)
- **SSH client** (usually pre-installed on Linux/macOS)

### Checking Prerequisites

```bash
# Check Go version
go version

# Check rsync availability
rsync --version

# Check SSH availability
ssh -V
```

### Installing VM RSync

1. **Clone the repository**
```bash
git clone https://github.com/carlosrabelo/vmrsync.git
cd vmrsync
```

2. **Build and install**
```bash
make build
make install         # → ~/.local/bin/vmrsync
make install-system  # → /usr/local/bin/vmrsync (sudo only for copy)
```

3. **Install shell completion** (root only, system-wide under `/usr/local`). Prefer `make install-system` so `sudo vmrsync` is on PATH (`secure_path` often omits `~/.local/bin`):
```bash
make install-system
sudo vmrsync completion bash
sudo vmrsync completion zsh
sudo vmrsync completion fish

# after user-local install:
sudo "$(command -v vmrsync)" completion bash
```

4. **Verify installation**
```bash
vmrsync version
```

### Manual Installation

```bash
make build
install -D bin/vmrsync $HOME/.local/bin/vmrsync
# or, as root:
# install -D bin/vmrsync /usr/local/bin/vmrsync
sudo "$(command -v vmrsync)" completion bash
```

### Uninstallation

```bash
make uninstall         # → ~/.local/bin (sudo only if needed)
make uninstall-system  # → /usr/local/bin (sudo only if needed)
# Completion (if installed):
sudo rm -f /usr/local/share/bash-completion/completions/vmrsync
sudo rm -f /usr/local/share/zsh/site-functions/_vmrsync
sudo rm -f /usr/local/share/fish/vendor_completions.d/vmrsync.fish
# Legacy user-local completion (older installs):
rm -f ~/.local/share/bash-completion/completions/vmrsync
```

## Configuration

### Environment Variables

**VMRSYNC_PATH** - Sync root directory, local and remote

```bash
# Set custom sync root
export VMRSYNC_PATH=$HOME/Projects

# Add to shell profile
echo 'export VMRSYNC_PATH=$HOME/Projects' >> ~/.bashrc
source ~/.bashrc
```

Default: `$HOME/Sources`

### SSH Configuration

#### Using SSH Keys (Recommended)

```bash
# Generate SSH key pair
ssh-keygen -t ed25519 -C "your_email@example.com"

# Copy public key to remote machine
ssh-copy-id user@remote-machine

# Test connection
ssh user@remote-machine
```

#### Custom SSH Port

```bash
# Specify port in command
vmrsync restore vm21 project1 --ssh-port 2222

# Or configure in ~/.ssh/config
Host vm21
    Port 2222
    User your-username
    IdentityFile ~/.ssh/id_rsa
```

#### SSH Config File

Create or edit `~/.ssh/config`:

```
Host vm21
    HostName 192.168.1.100
    User ubuntu
    Port 2222
    IdentityFile ~/.ssh/id_rsa

Host vm22
    HostName 192.168.1.101
    User ubuntu
    Port 22
    IdentityFile ~/.ssh/id_rsa
```

Now use host aliases:
```bash
vmrsync restore vm21 project1
vmrsync backup vm22 project2
```

### Preparing Remote Machines

#### Mirror Mode (Default)

No special setup needed. Just ensure your remote user has write access to the directories you want to sync.

#### Staging Mode

Setup required to use `/vmrsync` as remote root:

```bash
# Set up remote machine (requires sudo on remote)
vmrsync prepare <machine-name>

# Example
vmrsync prepare vm21

# Preview without executing
vmrsync prepare vm21 --dry-run
```

The prepare command:
- Creates `/vmrsync` on the remote machine
- Sets ownership to match your local UID
- Verifies the directory is ready for syncing

## Basic Usage

### Command Syntax

```
vmrsync <command> <machine> [<folder>] [options]
```

### Commands

| Command      | Description                              |
|--------------|------------------------------------------|
| `restore`    | Sync FROM remote TO local                |
| `backup`     | Sync FROM local TO remote                |
| `check`      | Verify SSH and that sync paths exist     |
| `prepare`    | Create and configure `/vmrsync` on remote |
| `completion` | Install system-wide shell completion (root only): bash, zsh, or fish |
| `version`    | Show version                             |
| `help`       | Show help (`-h`, `--help`)               |

### Options

| Option                  | Description                                            |
|-------------------------|--------------------------------------------------------|
| `--dry-run`             | Print the rsync command without executing it           |
| `--exclude <pattern>`   | Exclude files matching pattern (repeatable)            |
| `--exclude-from <file>` | Read exclude patterns from file                        |
| `--ssh-port <port>`     | SSH port                                               |
| `--ssh-key <path>`      | SSH private key path                                   |
| `--verbose`             | Enable verbose rsync output                            |
| `--quiet`               | Suppress progress output (errors still shown)          |
| `--no-delete`           | Do not delete files at destination                     |
| `--staging`             | Use /vmrsync as remote root instead of mirroring local path |
| `--timeout-seconds <n>` | Hard cap on rsync runtime in seconds (default `7200`; `0` disables) |
| `-h`, `--help`, `help`  | Show help (exit 0)                                     |

### Examples

```bash
# Sync entire directory tree
vmrsync restore vm21
vmrsync backup vm21

# Sync a specific folder
vmrsync restore vm21 project1
vmrsync backup vm21 project1

# Preview without syncing
vmrsync backup vm21 project1 --dry-run

# Exclude files
vmrsync backup vm21 project1 --exclude "*.log" --exclude "node_modules"

# Custom SSH options
vmrsync restore vm21 project1 --ssh-port 2222 --ssh-key ~/.ssh/id_rsa

# Use staging mode
vmrsync backup vm21 project1 --staging
```

### Path Structure

**Mirror mode (default):**
```
Local:  $VMRSYNC_PATH/[folder]/   →   Remote: $VMRSYNC_PATH/[folder]/
```

**Staging mode (--staging):**
```
Local:  $VMRSYNC_PATH/[folder]/   →   Remote: /vmrsync/[folder]/
```

If no folder is specified, the entire root is synced.

#### .vmrsyncignore

Place a `.vmrsyncignore` file in the folder being synced (same format as rsync exclude patterns, one per line). Patterns are loaded automatically on every sync, together with `--exclude-from` and `--exclude`.

```bash
# $VMRSYNC_PATH/project1/.vmrsyncignore
node_modules/
*.log
.git/
```

#### Preflight check

Before your first sync, verify SSH and paths:

```bash
vmrsync check vm21 project1
```

### First Sync

Always test with `--dry-run` first:

```bash
# Test sync from remote to local
vmrsync restore vm21 project1 --dry-run

# Test sync from local to remote
vmrsync backup vm21 project1 --dry-run

# Perform actual sync
vmrsync restore vm21 project1
```

## Advanced Usage

### Sync Modes

#### Mirror Mode (Default)

Replicates your local directory structure exactly on the remote machine.

```bash
# Local: /home/user/Sources/project1/
# Remote: /home/user/Sources/project1/

vmrsync backup vm21 project1
```

**Use cases:**
- Development environments that mirror production
- Working with multiple identical VMs
- Maintaining consistent directory structures

#### Staging Mode

Syncs everything to `/vmrsync` regardless of local path.

```bash
# Local: /home/user/Sources/project1/
# Remote: /vmrsync/project1/

vmrsync backup vm21 project1 --staging
```

**Use cases:**
- Centralized testing environment
- Shared workspace across team members
- Temporary staging before deployment

### Advanced Filtering

#### Multiple Exclude Patterns

```bash
vmrsync backup vm21 project1 \
  --exclude "*.log" \
  --exclude "*.tmp" \
  --exclude "node_modules/" \
  --exclude ".git/" \
  --exclude "*.pyc" \
  --exclude "__pycache__/"
```

#### Language-Specific Excludes

**JavaScript/Node.js:**
```bash
vmrsync backup vm21 project1 \
  --exclude "node_modules/" \
  --exclude "*.log" \
  --exclude ".npm/" \
  --exclude "dist/" \
  --exclude "build/"
```

**Python:**
```bash
vmrsync backup vm21 project1 \
  --exclude "__pycache__/" \
  --exclude "*.pyc" \
  --exclude "*.pyo" \
  --exclude ".venv/" \
  --exclude "*.egg-info/"
```

**Go:**
```bash
vmrsync backup vm21 project1 \
  --exclude "bin/" \
  --exclude "*.test" \
  --exclude "*.prof" \
  --exclude "vendor/" \
  --exclude ".go/"
```

### Performance Optimization

#### Large File Handling

```bash
# Use --no-delete to avoid accidental data loss
vmrsync backup vm21 project1 --no-delete

# Preview what will be synced
vmrsync backup vm21 project1 --dry-run --verbose

# Sync multiple smaller directories instead of one large one
vmrsync backup vm21 src --exclude "assets/"
vmrsync backup vm21 assets
```

#### Network Optimization

**Slow connections:**
```bash
vmrsync restore vm21 project1 \
  --exclude "*.log" \
  --exclude "*.tmp" \
  --exclude "node_modules/"
```

**Fast local network:**
```bash
vmrsync backup vm21 project1 --verbose
```

### Automation and Scripts

#### Pre-commit Hook

Create `.git/hooks/pre-commit`:
```bash
#!/bin/bash
BRANCH=$(git rev-parse --abbrev-ref HEAD)

if [ "$BRANCH" = "main" ]; then
    vmrsync backup vm21 . --exclude ".git/" --exclude "*.log"
fi
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

#### Post-commit Hook

Create `.git/hooks/post-commit`:
```bash
#!/bin/bash
vmrsync backup vm21 . --staging --exclude ".git/"
```

#### Cron Jobs

```bash
# Sync every 15 minutes (quiet for cron)
*/15 * * * * $HOME/.local/bin/vmrsync restore vm21 project1 --quiet

# Sync every hour with logging
0 * * * * $HOME/.local/bin/vmrsync restore vm21 project1 >> $HOME/vmrsync.log 2>&1

# Sync every morning at 8 AM
0 8 * * * $HOME/.local/bin/vmrsync restore vm21 project1
```

### Multi-Machine Workflows

#### Sync Across Multiple VMs

```bash
#!/bin/bash
MACHINES="vm21 vm22 vm23"
PROJECT="project1"

for machine in $MACHINES; do
    echo "Syncing to $machine..."
    vmrsync backup "$machine" "$PROJECT" --exclude "*.log"
done
```

#### Round-Robin Testing

```bash
#!/bin/bash
MACHINES="vm21 vm22 vm23"
PROJECT="project1"

for machine in $MACHINES; do
    echo "Running tests on $machine..."
    ssh "$machine" "cd /vmrsync/$PROJECT && make test"
done
```

### Backup and Recovery

#### Backup Strategy

```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="$HOME/backups/vmrsync/$DATE"

mkdir -p "$BACKUP_DIR"
vmrsync restore vm21 . --dry-run | tee "$BACKUP_DIR/rsync-preview.log"

read -p "Proceed with backup? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    vmrsync restore vm21 . | tee "$BACKUP_DIR/rsync.log"
    echo "Backup completed: $BACKUP_DIR"
fi
```

### Security Considerations

#### SSH Key Management

```bash
# Development key
vmrsync backup dev-server project1 --ssh-key ~/.ssh/id_dev

# Production key
vmrsync backup prod-server project1 --ssh-key ~/.ssh/id_prod
```

#### Network Security

`vmrsync` refuses targets that refer to this machine: `localhost` (and common variants), loopback addresses (`127.0.0.1`, `::1`, etc.), the local hostname from `os.Hostname()`, and any address assigned to a non-loopback interface on the host where you run the command. There is no flag to override this. It prevents accidental sync (especially with `--delete`) against your own `~/Sources` or mirror paths.

Use a **real remote hostname or IP** as the machine argument (not `localhost`). If you reach the VM through a bastion, put **`ProxyJump`** or **`ProxyCommand`** in `~/.ssh/config` (or other OpenSSH `Host`/`HostName` settings). `vmrsync` invokes `ssh`/`rsync` like a normal CLI SSH client, so those entries apply automatically when you pass the `Host` alias:

```bash
# ~/.ssh/config example:
# Host my-vm
#   HostName 10.0.0.50
#   User dev
#   ProxyJump user@jump-server

vmrsync backup my-vm project1
```

For a **fixed tree on the next machine** (under `/vmrsync` after `vmrsync prepare`), use **`--staging`** as documented above—same safety rules apply to the machine you pass.

#### rsync robustness notes

- All SSH invocations use `-o BatchMode=yes` and `-o ConnectTimeout=10` so syncs fail fast instead of prompting for a password.
- SSH connection failures are reported separately from missing remote directories.
- `vmrsync` passes `--protect-args` to rsync to avoid remote-side argument interpretation issues.
- `vmrsync` uses `--mkpath` when available; on older rsync versions it falls back to creating the remote path via SSH for `backup`.
- `--timeout-seconds` sets a hard upper bound on the rsync runtime (default: 7200 seconds; set to 0 to disable).

## Troubleshooting

### Installation Issues

#### Binary Not Found

```bash
# Check if binary exists
ls -la $HOME/.local/bin/vmrsync

# Add to PATH if needed
export PATH="$HOME/.local/bin:$PATH"

# Reload shell
source ~/.bashrc
```

#### Build Fails

```bash
# Verify Go installation
go version

# Clean and rebuild
make clean
make build
```

### Connection Problems

#### SSH Connection Refused

```bash
# Test SSH independently
ssh <machine>

# Check if SSH server is running
ssh <machine> "systemctl status ssh"

# Try with custom port
vmrsync restore <machine> project1 --ssh-port 2222
```

#### Authentication Failed

```bash
# Test SSH with verbose output
ssh -v <machine>

# Check SSH key permissions
ls -la ~/.ssh/id_rsa
chmod 600 ~/.ssh/id_rsa

# Copy public key to remote
ssh-copy-id <machine>

# Specify custom key
vmrsync restore <machine> project1 --ssh-key ~/.ssh/custom_key
```

#### Connection Timeout

```bash
# Test network connectivity
ping <machine>

# Verify DNS resolution
nslookup <machine>

# Try using IP address
vmrsync restore 192.168.1.100 project1
```

### Synchronization Errors

#### Directory Does Not Exist

```bash
# For staging mode
vmrsync prepare <machine>

# For mirror mode, create manually
ssh <machine> "mkdir -p $HOME/Sources/project1"

# Verify directory exists
ssh <machine> "ls -la $HOME/Sources/"
```

#### rsync Not Found

```bash
# Install rsync on local
sudo apt-get install rsync  # Ubuntu/Debian
sudo yum install rsync      # CentOS/RHEL

# Install rsync on remote
ssh <machine> "sudo apt-get install rsync"
```

#### Unexpected File Deletions

```bash
# Always test with dry-run first
vmrsync backup <machine> project1 --dry-run --verbose

# Use --no-delete to prevent deletions
vmrsync backup <machine> project1 --no-delete

# Verify direction
vmrsync restore <machine> project1  # Remote -> Local
vmrsync backup <machine> project1  # Local -> Remote
```

### Performance Issues

#### Very Slow Sync

```bash
# Check what's being transferred
vmrsync backup <machine> project1 --dry-run --verbose

# Exclude unnecessary files
vmrsync backup <machine> project1 \
  --exclude "*.log" \
  --exclude "*.tmp" \
  --exclude "node_modules/" \
  --exclude ".git/"

# Sync subdirectories separately
vmrsync backup <machine> project1/src
vmrsync backup <machine> project1/tests
```

#### High CPU Usage

```bash
# Reduce verbosity
vmrsync backup <machine> project1  # without --verbose

# Schedule off-peak
# Add to crontab: 0 2 * * * vmrsync backup <machine> project1

# Exclude more files
vmrsync backup <machine> project1 \
  --exclude "*.log" \
  --exclude "build/" \
  --exclude "dist/"
```

### Permission Problems

#### Permission Denied on Remote

```bash
# Check remote directory permissions
ssh <machine> "ls -la $HOME/Sources/project1"

# Fix permissions
ssh <machine> "chown -R $USER:$USER $HOME/Sources/project1"
ssh <machine> "chmod -R u+rw $HOME/Sources/project1"

# For staging mode, run prepare
vmrsync prepare <machine>
```

#### Sudo Required Error

```bash
# Run prepare manually
ssh <machine> "sudo mkdir -p /vmrsync && sudo chown $UID:$(id -g) /vmrsync"

# Or use mirror mode
vmrsync backup <machine> project1  # without --staging
```

### Getting Help

#### Collect Diagnostic Information

```bash
# System information
uname -a
go version
rsync --version
ssh -V

# VM RSync information
vmrsync version

# Network test
ping -c 4 <machine>
ssh -v <machine> "echo 'SSH connection works'"

# Dry run
vmrsync backup <machine> project1 --dry-run --verbose

# Environment variables
echo "VMRSYNC_PATH=$VMRSYNC_PATH"
```

#### Enable Logging

```bash
# Redirect output to log file
vmrsync backup <machine> project1 --verbose > vmrsync.log 2>&1
```

### Common Error Messages

| Error | Common Cause | Solution |
|-------|--------------|----------|
| `command not found` | VM RSync not in PATH | Add `$HOME/.local/bin` to PATH |
| `connection refused` | SSH server not running | Start SSH server on remote |
| `permission denied` | SSH key not configured | Setup SSH key authentication |
| `directory does not exist` | Remote directory missing | Run `vmrsync prepare <machine>` |
| `rsync: command not found` | rsync not installed | Install rsync on local/remote |
| `timeout` | Network issues | Check network connectivity |

### Best Practices

1. **Always test with --dry-run first**
2. **Use appropriate exclude patterns**
3. **Regular backups before major changes**
4. **Monitor sync operations with --verbose**
5. **Keep SSH keys secure (chmod 600)**
6. **Document your sync workflow**
7. **Test disaster recovery procedures**
8. **Keep software updated**

## Additional Resources

- [Main README](../README.md)
- [GitHub Repository](https://github.com/carlosrabelo/vmrsync)
- [rsync documentation](https://linux.die.net/man/1/rsync)
- [SSH documentation](https://www.openssh.com/manual.html)

---

*Last updated: April 2026*