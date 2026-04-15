# VM RSync

A Go CLI tool for bidirectional file synchronization between local and remote machines via rsync over SSH.

## Overview

`vmrsync` wraps rsync to synchronize a fixed directory structure between a local machine and a remote VM:

- `vmrsync in`: Sync FROM remote TO local
- `vmrsync out`: Sync FROM local TO remote
- `vmrsync setup`: Prepare the remote directory (requires sudo on remote)

The remote directory mirrors the local directory structure by default, or uses `/vmrsync` with `--staging`. The local root defaults to `$HOME/Sources`.

## Requirements

- Go (for building)
- rsync
- SSH access to remote machines

## Installation

```bash
make install
```

Installs `vmrsync` to `$HOME/.local/bin` and bash completion to `$HOME/.local/share/bash-completion/completions/`.

To uninstall:

```bash
make uninstall
```

## First-time Setup

Before syncing, `/vmrsync` must exist on the remote machine with correct ownership (UID 1000). Run:

```bash
vmrsync setup <machine>
```

This SSHes into the machine and runs `sudo mkdir -p /vmrsync && sudo chown 1000:1000 /vmrsync`.

Preview without executing:

```bash
vmrsync setup <machine> --dry-run
```

## Usage

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

# Use staging mode (/vmrsync instead of mirroring local path)
vmrsync out vm21 project1 --staging
```

## Options

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

## Environment Variables

| Variable            | Default              | Description              |
|---------------------|----------------------|--------------------------|
| `VMRSYNC_PATH`      | `$HOME/Sources`      | Sync root directory, local and remote     |

## Path Structure

By default (mirror mode):
```
Local:  $VMRSYNC_PATH/[folder]/   →   Remote: $VMRSYNC_PATH/[folder]/
```

With --staging flag:
```
Local:  $VMRSYNC_PATH/[folder]/   →   Remote: /vmrsync/[folder]/
```

If no folder is specified, the entire root is synced.

## How It Works

1. Checks that the target directory exists on the remote machine (skipped with `--dry-run`)
   - By default: uses the same path as local (mirrors local structure)
   - With --staging: uses `/vmrsync`
2. Builds an rsync command with `-az --info=progress2 --mkpath --delete`
3. Executes rsync over SSH

## Development

```bash
make build   # compile to bin/vmrsync
make test    # run tests
make lint    # run go vet
make fmt     # format source
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/name`
3. Ensure tests pass: `make test`
4. Open a Pull Request

Please keep documentation bilingual (English and Portuguese).

## License

This project is open source. See the LICENSE file for details.
