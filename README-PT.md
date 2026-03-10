# VM RSync

Sincronização bidirecional de arquivos entre uma árvore de trabalho local e máquinas remotas, com rsync sobre SSH e os comandos `restore`, `backup` e `prepare`.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.26%2B-blue.svg)](https://go.dev/)
[![Release](https://img.shields.io/github/release/carlosrabelo/vmrsync.svg)](https://github.com/carlosrabelo/vmrsync/releases)

## Destaques

- Puxe ou envie para o host remoto com `vmrsync restore` e `vmrsync backup`
- Inicialize `/vmrsync` no remoto com `vmrsync prepare` (usa `sudo` no remoto)
- Sincronize para `/vmrsync` no remoto por padrão, ou `--mirror` para usar o mesmo caminho do local
- Recusa alvos que se resolvem para esta máquina (localhost, loopback, hostname local, IPs locais)
- Visualize com `--dry-run`, limite o tempo com `--timeout-seconds`, ajuste SSH via `--ssh-port` / `--ssh-key`
- `--exclude`, `--exclude-from`, `.vmrsyncignore` automático, `--no-delete`, `--verbose`, `--quiet`
- `vmrsync check` verifica SSH e se os caminhos de sync existem antes de sincronizar
- Completion system-wide via `sudo vmrsync completion bash|zsh|fish`

## Pré-requisitos

- **Go 1.26+** — build a partir do código; [download](https://go.dev/dl/)
- **rsync** e **cliente OpenSSH** (`ssh`) na máquina onde você executa o `vmrsync`
- **Acesso SSH** ao host remoto (e `sudo` no remoto para `prepare`)
- **Root** — necessário para `completion` system-wide em `/usr/local`

## Instalação

### Build a partir do código

```bash
git clone https://github.com/carlosrabelo/vmrsync.git
cd vmrsync
make build
```

Instale em `~/.local/bin` (padrão), ou defina `PREFIX` para um prefixo customizado. Instalação system-wide é via `.deb`:

```bash
make install
make uninstall
make package
sudo dpkg -i dist/*.deb
```

Prefira o `.deb` (binário em `/usr/bin`) antes do completion para o `sudo vmrsync` estar no PATH (`secure_path` costuma omitir `~/.local/bin`):

```bash
make build
make package
sudo dpkg -i dist/*.deb
sudo vmrsync completion bash   # ou zsh / fish

# após instalação local do usuário:
sudo "$(command -v vmrsync)" completion bash
```

Veja o [Guia de Uso](docs/GUIDE-PT.md) para caminhos de limpeza do completion (system-wide e legado user-local).

### Com `go install`

```bash
go install github.com/carlosrabelo/vmrsync/vmrsync/cmd/vmrsync@latest
```

## Uso

Guias completos (também em [GitHub Pages](https://carlosrabelo.github.io/vmrsync/)): [User Guide](docs/GUIDE.md) · [Guia de Uso](docs/GUIDE-PT.md)

```bash
vmrsync prepare my-vm                 # cria /vmrsync no remoto (staging)
vmrsync check my-vm project1
vmrsync backup my-vm project1
vmrsync restore my-vm project1
vmrsync backup my-vm project1 --dry-run
vmrsync backup my-vm project1 --mirror
sudo vmrsync completion bash
```

Ordem fixa dos argumentos: `<command> <machine> [<folder>] [options]`.

A raiz padrão é `$HOME/Sources` (`VMRSYNC_PATH`). No modo staging (padrão) o remoto usa `/vmrsync`; com `--mirror` o remoto usa o mesmo caminho.

## Configuração

| Variável / flag | Padrão | Descrição |
|-----------------|--------|-----------|
| `VMRSYNC_PATH` | `$HOME/Sources` | Raiz de sync local (também remota com `--mirror`) |
| `--mirror` | off | Espelha o caminho local no remoto em vez de `/vmrsync` |
| `--timeout-seconds` | `7200` | Limite rígido do rsync (`0` desativa) |
| `.vmrsyncignore` | — | Padrões de exclusão carregados automaticamente |

## Estrutura do Projeto

```
vmrsync/cmd/vmrsync/   # Ponto de entrada da CLI
vmrsync/internal/      # cli, config, hostcheck, rsyncrun, completion, …
bin/                   # Binários compilados (git-ignored)
.make/                 # Scripts de build e install
docs/                  # Guias de uso + GitHub Pages
```

## Desenvolvimento

```bash
make build           # Compila o binário em bin/vmrsync
make test            # Executa todos os testes
make quality         # Formata, vet e lint
make install         # Instala o binário em ~/.local/bin
make uninstall       # Remove de ~/.local/bin
make package         # Gera .deb em dist/ (binário em /usr/bin)
make deb             # Alias de package
```

## Contribuição

1. Faça um fork do repositório
2. Crie uma branch de feature: `git checkout -b feat/description`
3. Commit com Conventional Commits: `git commit -m "feat: add X"`
4. Envie e abra um pull request

Mantenha a documentação bilíngue (inglês e português).

## Licença

Este projeto é licenciado sob a GNU General Public License v3.0 — veja [LICENSE](LICENSE) para detalhes.
