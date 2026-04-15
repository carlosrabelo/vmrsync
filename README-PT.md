# VM RSync

Uma ferramenta CLI em Go para sincronização bidirecional de arquivos entre máquinas locais e remotas via rsync sobre SSH.

## Visão Geral

`vmrsync` encapsula o rsync para sincronizar uma estrutura de diretórios fixa entre a máquina local e uma VM remota:

- `vmrsync in`: Sincroniza DO remoto PARA o local
- `vmrsync out`: Sincroniza DO local PARA o remoto
- `vmrsync setup`: Prepara o diretório remoto (requer sudo na máquina remota)

O diretório remoto é fixo em `/vmrsync`. O diretório local padrão é `$HOME/Sources`.

## Requisitos

- Go (para compilar)
- rsync
- Acesso SSH às máquinas remotas

## Instalação

```bash
make install
```

Instala `vmrsync` em `$HOME/.local/bin` e o bash completion em `$HOME/.local/share/bash-completion/completions/`.

Para desinstalar:

```bash
make uninstall
```

## Configuração Inicial

Antes de sincronizar, `/vmrsync` deve existir na máquina remota com a propriedade correta (UID 1000). Execute:

```bash
vmrsync setup <machine>
```

Isso acessa a máquina via SSH e executa `sudo mkdir -p /vmrsync && sudo chown 1000:1000 /vmrsync`.

Visualizar sem executar:

```bash
vmrsync setup <machine> --dry-run
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
```

## Opções

| Opção                | Descrição                                                  |
|----------------------|------------------------------------------------------------|
| `--dry-run`          | Exibe o comando rsync sem executá-lo                       |
| `--exclude <padrão>` | Exclui arquivos com o padrão (repetível)                   |
| `--ssh-port <porta>` | Porta SSH                                                  |
| `--ssh-key <caminho>`| Caminho da chave privada SSH                               |
| `--verbose`          | Habilita saída detalhada do rsync                          |
| `--no-delete`        | Não deleta arquivos no destino                             |
| `--backup-dir <path>`| Diretório de backup para arquivos deletados/substituídos   |
| `-h, --help`         | Exibe a ajuda                                              |

## Variáveis de Ambiente

| Variável            | Padrão               | Descrição                    |
|---------------------|----------------------|------------------------------|
| `VMRSYNC_LOCAL_ROOT`| `$HOME/Sources`      | Diretório raiz local         |

## Estrutura de Caminhos

```
Local:  $VMRSYNC_LOCAL_ROOT/[pasta]/   →   $HOME/Sources/[pasta]/
Remoto: /vmrsync/[pasta]/
```

Se nenhuma pasta for especificada, toda a raiz é sincronizada.

## Como Funciona

1. Verifica que `/vmrsync` existe na máquina remota (ignorado com `--dry-run`)
2. Monta um comando rsync com `-az --info=progress2 --mkpath --delete`
3. Executa o rsync via SSH

## Desenvolvimento

```bash
make build   # compila para bin/vmrsync
make test    # executa os testes
make lint    # executa go vet
make fmt     # formata o código-fonte
```

## Contribuindo

1. Faça um fork do repositório
2. Crie uma branch de feature: `git checkout -b feature/nome`
3. Certifique-se que os testes passam: `make test`
4. Abra um Pull Request

Por favor, mantenha a documentação bilíngue (inglês e português).

## Licença

Este projeto é open source. Consulte o arquivo LICENSE para detalhes.
