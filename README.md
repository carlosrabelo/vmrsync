# VM RSync

Bidirectional file synchronization between a local workspace tree and remote machines, driven by rsync over SSH with `restore`, `backup`, and `prepare` commands.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.26%2B-blue.svg)](https://go.dev/)
[![Release](https://img.shields.io/github/release/carlosrabelo/vmrsync.svg)](https://github.com/carlosrabelo/vmrsync/releases)

## Highlights

- Pull from or push to a remote host with `vmrsync restore` and `vmrsync backup`
- Initialize `/vmrsync` on the remote with `vmrsync prepare` (uses `sudo` on the remote)
- Mirror paths under `VMRSYNC_PATH` or use `--staging` to target `/vmrsync` on the remote
- Reject targets that resolve to this machine (localhost, loopback, local hostname, local IPs)
- Preview with `--dry-run`, cap runtime with `--timeout-seconds`, tune SSH via `--ssh-port` / `--ssh-key`
- `--exclude`, `--exclude-from`, auto-loaded `.vmrsyncignore`, `--no-delete`, `--verbose`, `--quiet`
- `vmrsync check` verifies SSH and that local/remote sync paths exist before syncing
- System-wide shell completion via `sudo vmrsync completion bash|zsh|fish`

## Prerequisites

- **Go 1.26+** — build from source; [download](https://go.dev/dl/)
- **rsync** and **OpenSSH client** (`ssh`) on the machine where you run `vmrsync`
- **SSH access** to the remote host (and `sudo` on the remote for `prepare`)
- **Root** — required for system-wide `completion` under `/usr/local`

## Installation

### Build from Source

```bash
git clone https://github.com/carlosrabelo/vmrsync.git
cd vmrsync
make build
```

```bash
make install         # → ~/.local/bin/vmrsync
make install-system  # → /usr/local/bin/vmrsync (sudo only for copy)
```

Prefer `make install-system` before completion so `sudo vmrsync` is on PATH (`secure_path` often omits `~/.local/bin`):

```bash
make install-system
sudo vmrsync completion bash   # or zsh / fish

# after user-local install:
sudo "$(command -v vmrsync)" completion bash
```

### Uninstall

```bash
make uninstall         # → ~/.local/bin (sudo only if needed)
make uninstall-system  # → /usr/local/bin (sudo only if needed)
```

See the [User Guide](docs/GUIDE.md) for completion cleanup paths (system-wide and legacy user-local).

### Using `go install`

```bash
go install github.com/carlosrabelo/vmrsync/vmrsync/cmd/vmrsync@latest
```

## Usage

Full guides (also on [GitHub Pages](https://carlosrabelo.github.io/vmrsync/)): [User Guide](docs/GUIDE.md) · [Guia de Uso](docs/GUIDE-PT.md)

```bash
vmrsync prepare my-vm                 # create /vmrsync on remote (staging)
vmrsync check my-vm project1
vmrsync backup my-vm project1
vmrsync restore my-vm project1
vmrsync backup my-vm project1 --dry-run
vmrsync backup my-vm project1 --staging
sudo vmrsync completion bash
```

Default sync root is `$HOME/Sources` (`VMRSYNC_PATH`). Mirror mode uses the same path on the remote; `--staging` uses `/vmrsync` on the remote.

## Configuration

| Variable / flag | Default | Description |
|-----------------|---------|-------------|
| `VMRSYNC_PATH` | `$HOME/Sources` | Sync root, local and remote (mirror mode) |
| `--staging` | off | Use `/vmrsync` as remote root |
| `--timeout-seconds` | `7200` | Hard rsync timeout (`0` disables) |
| `.vmrsyncignore` | — | Auto-loaded exclude patterns in the synced folder |

## Project Layout

```
vmrsync/cmd/vmrsync/   # CLI entry point
vmrsync/internal/      # cli, config, hostcheck, rsyncrun, completion, …
bin/                   # Compiled binaries (git-ignored)
.make/                 # Build and install scripts
docs/                  # User guides + GitHub Pages
```

## Development

```bash
make build           # Compile binary to bin/vmrsync
make test            # Run all tests
make quality         # Format, vet, and lint
make install         # Install binary to ~/.local/bin
make install-system  # Install binary to /usr/local/bin
make uninstall       # Remove from ~/.local/bin
make uninstall-system  # Remove from /usr/local/bin
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/description`
3. Commit with Conventional Commits: `git commit -m "feat: add X"`
4. Push and open a pull request

Please keep documentation bilingual (English and Portuguese).

## License

This project is licensed under the GNU General Public License v3.0 — see [LICENSE](LICENSE) for details.
