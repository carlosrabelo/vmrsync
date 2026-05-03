# VM RSync

Bidirectional file synchronization between a local workspace tree and remote machines, driven by rsync over SSH with `in`, `out`, and `setup` commands.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.26%2B-blue.svg)](https://go.dev/)

## Highlights

- Pull from or push to a remote host with `vmrsync in` and `vmrsync out`
- Initialize `/vmrsync` on the remote with `vmrsync setup` (uses `sudo` on the remote)
- Mirror paths under `VMRSYNC_PATH` or use `--staging` to target `/vmrsync` on the remote
- Reject targets that resolve to this machine (localhost, loopback, local hostname, and local interface addresses) to reduce accidental destructive syncs
- Preview commands with `--dry-run`, cap runtime with `--timeout-seconds`, and tune SSH via `--ssh-port` and `--ssh-key`
- Repeatable `--exclude` patterns, `--no-delete`, and `--verbose` for rsync output
- Install copies the binary to `~/.local/bin` and installs bash completion when you `make install`
- Remote directory preflight uses POSIX-style `test -d` with a quoted path (no GNU-only `--`) so BusyBox and minimal shells on the VM behave correctly

## Overview

`vmrsync` wraps `rsync` so you keep the same relative layout locally and on the VM. The default local and remote root is `$HOME/Sources` unless you set `VMRSYNC_PATH`. For staging-style trees fixed on the remote under `/vmrsync`, pass `--staging`. See [docs/GUIDE.md](docs/GUIDE.md) for network safety notes (for example bastions and `ProxyJump`).

## Prerequisites

- **Go 1.26+** — build from source; see [go.dev/dl](https://go.dev/dl/)
- **rsync** and **OpenSSH client** (`ssh`) on the machine where you run `vmrsync`
- **SSH access** to the remote host as a user that can read and write the synced paths (and `sudo` for `setup` when creating `/vmrsync`)

## Installation

### Build from source

```bash
git clone https://github.com/carlosrabelo/vmrsync.git
cd vmrsync
make build
```

### Install to `~/.local/bin`

```bash
make install
```

This installs `vmrsync` to `$HOME/.local/bin` and bash completion to `$HOME/.local/share/bash-completion/completions/` when the completion file is present.

### Uninstall

```bash
make uninstall
```

### Using `go install`

```bash
go install github.com/carlosrabelo/vmrsync/vmrsync/cmd/vmrsync@latest
```

## Quick Start

Ensure `/vmrsync` exists on the remote when you plan to use `--staging`:

```bash
vmrsync setup my-vm
```

Then sync a folder:

```bash
vmrsync out my-vm project1
vmrsync in my-vm project1
```

Preview the rsync command without running it:

```bash
vmrsync out my-vm project1 --dry-run
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

### Sync paths

By default (mirror mode):

```
Local:  $VMRSYNC_PATH/[folder]/   →   Remote: $VMRSYNC_PATH/[folder]/
```

With `--staging`:

```
Local:  $VMRSYNC_PATH/[folder]/   →   Remote: /vmrsync/[folder]/
```

If `folder` is omitted, the whole root under `VMRSYNC_PATH` is synced.

### Behavior

1. Verifies the destination path exists on the remote over non-interactive SSH (skipped with `--dry-run`): runs `test -d` with a single-quoted path; mirror mode expects the **same absolute path** on the remote as under `VMRSYNC_PATH` locally; `--staging` checks `/vmrsync`. If this step fails, confirm `ssh -o BatchMode=yes <machine> "test -d <path>"` from the same environment as `vmrsync` — SSH auth or shell errors are currently reported with the same message as a missing directory
2. Builds an `rsync` invocation with `-az --info=progress2 --mkpath` and delete semantics unless `--no-delete` is set
3. Runs `rsync` over SSH

## Configuration

### Flags

| Option                  | Description                                                      |
|-------------------------|------------------------------------------------------------------|
| `--dry-run`             | Print the rsync command without executing it                     |
| `--exclude <pattern>`   | Exclude files matching pattern (repeatable)                      |
| `--ssh-port <port>`     | SSH port                                                         |
| `--ssh-key <path>`      | SSH private key path                                             |
| `--verbose`             | Enable verbose rsync output                                      |
| `--no-delete`           | Do not delete files at destination                               |
| `--staging`             | Use `/vmrsync` as remote root instead of mirroring the local path |
| `--timeout-seconds <n>` | Hard cap on rsync runtime in seconds (default `7200`; `0` disables) |
| `-h`, `--help`          | Show help                                                        |

### Environment variables

| Variable       | Default         | Description                               |
|----------------|-----------------|-------------------------------------------|
| `VMRSYNC_PATH` | `$HOME/Sources` | Root directory for sync, local and remote |

## Project Layout

```
vmrsync/cmd/vmrsync/   # Go entrypoint (`main` package)
vmrsync/internal/      # Private packages (cli, config, hostcheck, rsyncrun, …)
.make/                 # Build, test, install, and uninstall shell helpers
docs/                  # Long-form guides (English and Portuguese)
bin/                   # Compiled binary (git-ignored; created by `make build`)
vmrsync.bash-completion
Makefile
go.mod
LICENSE
```

## Development

```bash
make build      # Compile binary to bin/vmrsync
make test       # Run Go unit tests
make lint       # Run go vet
make fmt        # Format code with gofmt
make clean      # Remove build artifacts under bin/
make install    # Build and install to ~/.local/bin
make uninstall  # Remove binary and completion from ~/.local
make help       # List Makefile targets
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/description`
3. Ensure tests pass: `make test`
4. Open a Pull Request

Please keep documentation bilingual (English and Portuguese).

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.
