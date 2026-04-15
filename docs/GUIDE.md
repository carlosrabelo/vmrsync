# VM RSync - Complete Usage Guide

This comprehensive guide covers installation, configuration, usage, and troubleshooting for VM RSync.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Basic Usage](#basic-usage)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

## Installation

### Prerequisites

Before installing VM RSync, ensure you have:
- **Go** (version 1.16 or later)
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
make install
```

This installs:
- Binary to `$HOME/.local/bin/vmrsync`
- Bash completion to `$HOME/.local/share/bash-completion/completions/vmrsync`

3. **Verify installation**
```bash
vmrsync version
```

### Manual Installation

```bash
make build
cp bin/vmrsync $HOME/.local/bin/
mkdir -p $HOME/.local/share/bash-completion/completions
cp vmrsync.bash-completion $HOME/.local/share/bash-completion/completions/vmrsync
```

### Uninstallation

```bash
make uninstall
# Or manually:
rm $HOME/.local/bin/vmrsync
rm $HOME/.local/share/bash-completion/completions/vmrsync
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
vmrsync in vm21 project1 --ssh-port 2222

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
vmrsync in vm21 project1
vmrsync out vm22 project2
```

### Setting Up Remote Machines

#### Mirror Mode (Default)

No special setup needed. Just ensure your remote user has write access to the directories you want to sync.

#### Staging Mode

Setup required to use `/vmrsync` as remote root:

```bash
# Set up remote machine (requires sudo on remote)
vmrsync setup <machine-name>

# Example
vmrsync setup vm21

# Preview without executing
vmrsync setup vm21 --dry-run
```

The setup command:
- Creates `/vmrsync` on the remote machine
- Sets ownership to match your local UID
- Verifies the directory is ready for syncing

## Basic Usage

### Command Syntax

```
vmrsync <command> <machine> [<folder>] [options]
```

### Commands

| Command   | Description                              |
|-----------|------------------------------------------|
| `in`      | Sync FROM remote TO local                |
| `out`     | Sync FROM local TO remote                |
| `setup`   | Create and configure `/vmrsync` on remote |
| `version` | Show version                             |

### Options

| Option               | Description                                            |
|----------------------|--------------------------------------------------------|
| `--dry-run`          | Print the rsync command without executing it           |
| `--exclude <pattern>`| Exclude files matching pattern (repeatable)            |
| `--ssh-port <port>`  | SSH port                                               |
| `--ssh-key <path>`   | SSH private key path                                   |
| `--verbose`          | Enable verbose rsync output                            |
| `--no-delete`        | Do not delete files at destination                     |
| `--staging`          | Use /vmrsync as remote root instead of mirroring local path |
| `-h, --help`         | Show help                                              |

### Examples

```bash
# Sync entire directory tree
vmrsync in vm21
vmrsync out vm21

# Sync a specific folder
vmrsync in vm21 project1
vmrsync out vm21 project1

# Preview without syncing
vmrsync out vm21 project1 --dry-run

# Exclude files
vmrsync out vm21 project1 --exclude "*.log" --exclude "node_modules"

# Custom SSH options
vmrsync in vm21 project1 --ssh-port 2222 --ssh-key ~/.ssh/id_rsa

# Use staging mode
vmrsync out vm21 project1 --staging
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

### First Sync

Always test with `--dry-run` first:

```bash
# Test sync from remote to local
vmrsync in vm21 project1 --dry-run

# Test sync from local to remote
vmrsync out vm21 project1 --dry-run

# Perform actual sync
vmrsync in vm21 project1
```

## Advanced Usage

### Sync Modes

#### Mirror Mode (Default)

Replicates your local directory structure exactly on the remote machine.

```bash
# Local: /home/user/Sources/project1/
# Remote: /home/user/Sources/project1/

vmrsync out vm21 project1
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

vmrsync out vm21 project1 --staging
```

**Use cases:**
- Centralized testing environment
- Shared workspace across team members
- Temporary staging before deployment

### Advanced Filtering

#### Multiple Exclude Patterns

```bash
vmrsync out vm21 project1 \
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
vmrsync out vm21 project1 \
  --exclude "node_modules/" \
  --exclude "*.log" \
  --exclude ".npm/" \
  --exclude "dist/" \
  --exclude "build/"
```

**Python:**
```bash
vmrsync out vm21 project1 \
  --exclude "__pycache__/" \
  --exclude "*.pyc" \
  --exclude "*.pyo" \
  --exclude ".venv/" \
  --exclude "*.egg-info/"
```

**Go:**
```bash
vmrsync out vm21 project1 \
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
vmrsync out vm21 project1 --no-delete

# Preview what will be synced
vmrsync out vm21 project1 --dry-run --verbose

# Sync multiple smaller directories instead of one large one
vmrsync out vm21 src --exclude "assets/"
vmrsync out vm21 assets
```

#### Network Optimization

**Slow connections:**
```bash
vmrsync in vm21 project1 \
  --exclude "*.log" \
  --exclude "*.tmp" \
  --exclude "node_modules/"
```

**Fast local network:**
```bash
vmrsync out vm21 project1 --verbose
```

### Automation and Scripts

#### Pre-commit Hook

Create `.git/hooks/pre-commit`:
```bash
#!/bin/bash
BRANCH=$(git rev-parse --abbrev-ref HEAD)

if [ "$BRANCH" = "main" ]; then
    vmrsync out vm21 . --exclude ".git/" --exclude "*.log"
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
vmrsync out vm21 . --staging --exclude ".git/"
```

#### Cron Jobs

```bash
# Sync every 15 minutes
*/15 * * * * $HOME/.local/bin/vmrsync in vm21 project1

# Sync every hour with logging
0 * * * * $HOME/.local/bin/vmrsync in vm21 project1 >> $HOME/vmrsync.log 2>&1

# Sync every morning at 8 AM
0 8 * * * $HOME/.local/bin/vmrsync in vm21 project1
```

### Multi-Machine Workflows

#### Sync Across Multiple VMs

```bash
#!/bin/bash
MACHINES="vm21 vm22 vm23"
PROJECT="project1"

for machine in $MACHINES; do
    echo "Syncing to $machine..."
    vmrsync out "$machine" "$PROJECT" --exclude "*.log"
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
vmrsync in vm21 . --dry-run | tee "$BACKUP_DIR/rsync-preview.log"

read -p "Proceed with backup? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    vmrsync in vm21 . | tee "$BACKUP_DIR/rsync.log"
    echo "Backup completed: $BACKUP_DIR"
fi
```

### Security Considerations

#### SSH Key Management

```bash
# Development key
vmrsync out dev-server project1 --ssh-key ~/.ssh/id_dev

# Production key
vmrsync out prod-server project1 --ssh-key ~/.ssh/id_prod
```

#### Network Security

```bash
# Create SSH tunnel
ssh -L 2222:remote-server:22 jump-server

# Use tunnel for sync
vmrsync out localhost:2222 project1 --ssh-port 2222
```

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
vmrsync in <machine> project1 --ssh-port 2222
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
vmrsync in <machine> project1 --ssh-key ~/.ssh/custom_key
```

#### Connection Timeout

```bash
# Test network connectivity
ping <machine>

# Verify DNS resolution
nslookup <machine>

# Try using IP address
vmrsync in 192.168.1.100 project1
```

### Synchronization Errors

#### Directory Does Not Exist

```bash
# For staging mode
vmrsync setup <machine>

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
vmrsync out <machine> project1 --dry-run --verbose

# Use --no-delete to prevent deletions
vmrsync out <machine> project1 --no-delete

# Verify direction
vmrsync in <machine> project1  # Remote -> Local
vmrsync out <machine> project1  # Local -> Remote
```

### Performance Issues

#### Very Slow Sync

```bash
# Check what's being transferred
vmrsync out <machine> project1 --dry-run --verbose

# Exclude unnecessary files
vmrsync out <machine> project1 \
  --exclude "*.log" \
  --exclude "*.tmp" \
  --exclude "node_modules/" \
  --exclude ".git/"

# Sync subdirectories separately
vmrsync out <machine> project1/src
vmrsync out <machine> project1/tests
```

#### High CPU Usage

```bash
# Reduce verbosity
vmrsync out <machine> project1  # without --verbose

# Schedule off-peak
# Add to crontab: 0 2 * * * vmrsync out <machine> project1

# Exclude more files
vmrsync out <machine> project1 \
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

# For staging mode, run setup
vmrsync setup <machine>
```

#### Sudo Required Error

```bash
# Run setup manually
ssh <machine> "sudo mkdir -p /vmrsync && sudo chown $UID:$UID /vmrsync"

# Or use mirror mode
vmrsync out <machine> project1  # without --staging
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
vmrsync out <machine> project1 --dry-run --verbose

# Environment variables
echo "VMRSYNC_PATH=$VMRSYNC_PATH"
```

#### Enable Logging

```bash
# Redirect output to log file
vmrsync out <machine> project1 --verbose > vmrsync.log 2>&1
```

### Common Error Messages

| Error | Common Cause | Solution |
|-------|--------------|----------|
| `command not found` | VM RSync not in PATH | Add `$HOME/.local/bin` to PATH |
| `connection refused` | SSH server not running | Start SSH server on remote |
| `permission denied` | SSH key not configured | Setup SSH key authentication |
| `directory does not exist` | Remote directory missing | Run `vmrsync setup <machine>` |
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