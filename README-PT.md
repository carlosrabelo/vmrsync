# VM RSync

Sincronização bidirecional de arquivos entre uma árvore de trabalho local e máquinas remotas, com rsync sobre SSH e os comandos `in`, `out` e `setup`.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.26%2B-blue.svg)](https://go.dev/)

## Destaques

- Puxe ou envie para o host remoto com `vmrsync in` e `vmrsync out`
- Inicialize `/vmrsync` no remoto com `vmrsync setup` (usa `sudo` no remoto)
- Espelhe caminhos sob `VMRSYNC_PATH` ou use `--staging` para usar `/vmrsync` no remoto
- Recusa alvos que se resolvem para esta máquina (localhost, loopback, hostname local e endereços de interfaces locais) para reduzir sincronizações destrutivas acidentais
- Pré-visualize comandos com `--dry-run`, limite o tempo com `--timeout-seconds` e ajuste SSH com `--ssh-port` e `--ssh-key`
- `--exclude` repetível, `--no-delete` e `--verbose` para a saída do rsync
- A instalação copia o binário para `~/.local/bin` e instala bash completion ao rodar `make install`

## Visão Geral

O `vmrsync` encapsula o `rsync` para manter o mesmo layout relativo no computador local e na VM. A raiz local e remota padrão é `$HOME/Sources`, salvo se você definir `VMRSYNC_PATH`. Para árvores fixas no remoto sob `/vmrsync`, passe `--staging`. Veja [docs/GUIDE-PT.md](docs/GUIDE-PT.md) para notas de segurança de rede (por exemplo bastion e `ProxyJump`).

## Pré-requisitos

- **Go 1.26+** — compilar a partir do código; veja [go.dev/dl](https://go.dev/dl/)
- **rsync** e **cliente OpenSSH** (`ssh`) na máquina onde você executa o `vmrsync`
- **Acesso SSH** ao host remoto como um usuário que possa ler e gravar nos caminhos sincronizados (e `sudo` para o `setup` ao criar `/vmrsync`)

## Instalação

### Compilar a partir do código

```bash
git clone https://github.com/carlosrabelo/vmrsync.git
cd vmrsync
make build
```

### Instalar em `~/.local/bin`

```bash
make install
```

Isso instala o `vmrsync` em `$HOME/.local/bin` e o bash completion em `$HOME/.local/share/bash-completion/completions/` quando o arquivo de completion estiver presente.

### Desinstalar

```bash
make uninstall
```

### Com `go install`

```bash
go install github.com/carlosrabelo/vmrsync/vmrsync/cmd/vmrsync@latest
```

## Início Rápido

Garanta que `/vmrsync` exista no remoto quando for usar `--staging`:

```bash
vmrsync setup my-vm
```

Depois sincronize uma pasta:

```bash
vmrsync out my-vm project1
vmrsync in my-vm project1
```

Pré-visualize o comando rsync sem executar:

```bash
vmrsync out my-vm project1 --dry-run
```

## Uso

```
vmrsync <command> <machine> [<folder>] [options]
```

### Comandos

| Comando   | Descrição                                    |
|-----------|----------------------------------------------|
| `in`      | Sincroniza DO remoto PARA o local            |
| `out`     | Sincroniza DO local PARA o remoto            |
| `setup`   | Cria e configura `/vmrsync` no remoto        |
| `version` | Exibe a versão                               |

### Exemplos

```bash
# Sincronizar toda a árvore de diretórios
vmrsync in vm21
vmrsync out vm21

# Sincronizar uma pasta específica
vmrsync in vm21 project1
vmrsync out vm21 project1

# Visualizar sem sincronizar
vmrsync out vm21 project1 --dry-run

# Excluir arquivos
vmrsync out vm21 project1 --exclude "*.log" --exclude "node_modules"

# Opções SSH personalizadas
vmrsync in vm21 project1 --ssh-port 2222 --ssh-key ~/.ssh/id_rsa

# Usar modo staging (/vmrsync em vez de espelhar caminho local)
vmrsync out vm21 project1 --staging
```

### Caminhos de sincronização

Por padrão (modo espelhamento):

```
Local:  $VMRSYNC_PATH/[folder]/   →   Remoto: $VMRSYNC_PATH/[folder]/
```

Com `--staging`:

```
Local:  $VMRSYNC_PATH/[folder]/   →   Remoto: /vmrsync/[folder]/
```

Se `folder` for omitida, sincroniza-se toda a raiz sob `VMRSYNC_PATH`.

### Comportamento

1. Verifica se o caminho de destino existe no remoto (ignorado com `--dry-run`): no modo espelho usa o mesmo caminho que no local; com `--staging` usa `/vmrsync`
2. Monta uma chamada ao `rsync` com `-az --info=progress2 --mkpath` e remoção no destino salvo se `--no-delete` estiver definido
3. Executa o `rsync` via SSH

## Configuração

### Flags

| Opção                   | Descrição                                                        |
|-------------------------|------------------------------------------------------------------|
| `--dry-run`             | Exibe o comando rsync sem executá-lo                            |
| `--exclude <padrão>`    | Exclui arquivos que casam com o padrão (repetível)               |
| `--ssh-port <porta>`    | Porta SSH                                                        |
| `--ssh-key <caminho>`   | Caminho da chave privada SSH                                     |
| `--verbose`             | Habilita saída detalhada do rsync                                |
| `--no-delete`           | Não remove arquivos no destino                                   |
| `--staging`             | Usa `/vmrsync` como raiz remota em vez de espelhar o caminho local |
| `--timeout-seconds <n>` | Limite rígido de tempo do rsync em segundos (padrão `7200`; `0` desativa) |
| `-h`, `--help`          | Exibe a ajuda                                                    |

### Variáveis de ambiente

| Variável       | Padrão          | Descrição                                      |
|----------------|-----------------|------------------------------------------------|
| `VMRSYNC_PATH` | `$HOME/Sources` | Diretório raiz da sincronização, local e remoto |

## Estrutura do Projeto

```
vmrsync/cmd/vmrsync/   # Ponto de entrada Go (pacote `main`)
.make/                 # Scripts shell de build, teste, instalação e desinstalação
docs/                  # Guias longos (inglês e português)
bin/                   # Binário compilado (ignorado pelo git; criado por `make build`)
vmrsync.bash-completion
Makefile
go.mod
LICENSE
```

## Desenvolvimento

```bash
make build      # Compila o binário em bin/vmrsync
make test       # Executa os testes unitários em Go
make lint       # Executa go vet
make fmt        # Formata o código com gofmt
make clean      # Remove artefatos de build em bin/
make install    # Compila e instala em ~/.local/bin
make uninstall  # Remove o binário e o completion de ~/.local
make help       # Lista os alvos do Makefile
```

## Contribuindo

1. Faça um fork do repositório
2. Crie uma branch de feature: `git checkout -b feature/nome`
3. Certifique-se de que os testes passam: `make test`
4. Abra um Pull Request

Por favor, mantenha a documentação bilíngue (inglês e português).

## Licença

Este projeto é licenciado sob a Licença MIT — veja [LICENSE](LICENSE) para detalhes.
