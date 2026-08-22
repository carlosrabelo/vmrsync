---
layout: default
title: Guia de Uso
---

# VM RSync — Guia de Uso

Instalação, configuração, uso e solução de problemas do VM RSync.

## Sumário

- [Instalação](#instalação)
- [Configuração](#configuração)
- [Uso Básico](#uso-básico)
- [Uso Avançado](#uso-avançado)
- [Solução de Problemas](#solução-de-problemas)

## Instalação

### Pré-requisitos

Antes de instalar o VM RSync, certifique-se de ter:
- **Go** (versão 1.26 ou posterior)
- **rsync** (geralmente pré-instalado no Linux/macOS)
- **Cliente SSH** (geralmente pré-instalado no Linux/macOS)

### Verificando Pré-requisitos

```bash
# Verificar versão do Go
go version

# Verificar disponibilidade do rsync
rsync --version

# Verificar disponibilidade do SSH
ssh -V
```

### Instalando VM RSync

1. **Clone o repositório**
```bash
git clone https://github.com/carlosrabelo/vmrsync.git
cd vmrsync
```

2. **Compile e instale**
```bash
make build
make install         # → ~/.local/bin/vmrsync
make install-system  # → /usr/local/bin/vmrsync (sudo só na cópia)
```

3. **Instale o shell completion** (somente root, system-wide em `/usr/local`). Prefira `make install-system` para o `sudo vmrsync` estar no PATH (`secure_path` costuma omitir `~/.local/bin`):
```bash
make install-system
sudo vmrsync completion bash
sudo vmrsync completion zsh
sudo vmrsync completion fish

# após instalação local do usuário:
sudo "$(command -v vmrsync)" completion bash
```

4. **Verifique a instalação**
```bash
vmrsync version
```

### Instalação Manual

```bash
make build
install -D bin/vmrsync $HOME/.local/bin/vmrsync
# ou, como root:
# install -D bin/vmrsync /usr/local/bin/vmrsync
sudo "$(command -v vmrsync)" completion bash
```

### Desinstalação

```bash
make uninstall         # → ~/.local/bin (sudo só se necessário)
make uninstall-system  # → /usr/local/bin (sudo só se necessário)
# Completion (se instalado):
sudo rm -f /usr/local/share/bash-completion/completions/vmrsync
sudo rm -f /usr/local/share/zsh/site-functions/_vmrsync
sudo rm -f /usr/local/share/fish/vendor_completions.d/vmrsync.fish
# Completion antigo do usuário (instalações antigas):
rm -f ~/.local/share/bash-completion/completions/vmrsync
```
## Configuração

### Variáveis de Ambiente

**VMRSYNC_PATH** - Diretório raiz de sincronização, local e remoto

```bash
# Definir raiz de sincronização personalizada
export VMRSYNC_PATH=$HOME/Projetos

# Adicionar ao perfil do shell
echo 'export VMRSYNC_PATH=$HOME/Projetos' >> ~/.bashrc
source ~/.bashrc
```

Padrão: `$HOME/Sources`

### Configuração SSH

#### Usando Chaves SSH (Recomendado)

```bash
# Gerar par de chaves SSH
ssh-keygen -t ed25519 -C "seu_email@example.com"

# Copiar chave pública para máquina remota
ssh-copy-id usuario@maquina-remota

# Testar conexão
ssh usuario@maquina-remota
```

#### Porta SSH Personalizada

```bash
# Especificar porta no comando
vmrsync restore vm21 project1 --ssh-port 2222

# Ou configurar em ~/.ssh/config
Host vm21
    Port 2222
    User seu-usuario
    IdentityFile ~/.ssh/id_rsa
```

#### Arquivo de Configuração SSH

Crie ou edite `~/.ssh/config`:

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

Agora use aliases dos hosts:
```bash
vmrsync restore vm21 project1
vmrsync backup vm22 project2
```

### Preparando Máquinas Remotas

#### Modo Espelhamento (Padrão)

Nenhuma configuração especial necessária. Apenas certifique-se que seu usuário remoto tem acesso de escrita aos diretórios que deseja sincronizar.

#### Modo Staging

Configuração necessária para usar `/vmrsync` como raiz remota:

```bash
# Configurar máquina remota (requer sudo no remoto)
vmrsync prepare <nome-da-maquina>

# Exemplo
vmrsync prepare vm21

# Visualizar sem executar
vmrsync prepare vm21 --dry-run
```

O comando prepare:
- Cria `/vmrsync` na máquina remota
- Define a propriedade para corresponder ao seu UID local
- Verifica se o diretório está pronto para sincronização

## Uso Básico

### Sintaxe do Comando

```
vmrsync <command> <machine> [<folder>] [options]
```

### Comandos

| Comando      | Descrição                                     |
|--------------|-----------------------------------------------|
| `restore`    | Sincroniza DO remoto PARA o local             |
| `backup`     | Sincroniza DO local PARA o remoto             |
| `check`      | Verifica SSH e se os caminhos de sync existem |
| `prepare`    | Cria e configura `/vmrsync` no remoto         |
| `completion` | Instala completion system-wide (somente root): bash, zsh ou fish |
| `version`    | Exibe a versão                                |
| `help`       | Exibe a ajuda (`-h`, `--help`)                |

### Opções

| Opção                   | Descrição                                                  |
|-------------------------|------------------------------------------------------------|
| `--dry-run`             | Exibe o comando rsync sem executá-lo                       |
| `--exclude <padrão>`    | Exclui arquivos com o padrão (repetível)                   |
| `--exclude-from <file>` | Lê padrões de exclusão de um arquivo                       |
| `--ssh-port <porta>`    | Porta SSH                                                  |
| `--ssh-key <caminho>`   | Caminho da chave privada SSH                               |
| `--verbose`             | Habilita saída detalhada do rsync                          |
| `--quiet`               | Suprime saída de progresso (erros ainda são exibidos)      |
| `--no-delete`           | Não deleta arquivos no destino                             |
| `--staging`             | Usa /vmrsync como raiz remota em vez de espelhar caminho local |
| `--timeout-seconds <n>` | Limite rígido de tempo do rsync em segundos (padrão `7200`; `0` desativa) |
| `-h`, `--help`, `help`  | Exibe a ajuda (exit 0)                                     |

### Exemplos

```bash
# Sincronizar toda a árvore de diretórios
vmrsync restore vm21
vmrsync backup vm21

# Sincronizar uma pasta específica
vmrsync restore vm21 project1
vmrsync backup vm21 project1

# Visualizar sem sincronizar
vmrsync backup vm21 project1 --dry-run

# Excluir arquivos
vmrsync backup vm21 project1 --exclude "*.log" --exclude "node_modules"

# Opções SSH personalizadas
vmrsync restore vm21 project1 --ssh-port 2222 --ssh-key ~/.ssh/id_rsa

# Usar modo staging
vmrsync backup vm21 project1 --staging
```

### Estrutura de Caminhos

**Modo espelhamento (padrão):**
```
Local:  $VMRSYNC_PATH/[pasta]/   →   Remoto: $VMRSYNC_PATH/[pasta]/
```

**Modo staging (--staging):**
```
Local:  $VMRSYNC_PATH/[pasta]/   →   Remoto: /vmrsync/[pasta]/
```

Se nenhuma pasta for especificada, toda a raiz é sincronizada.

#### .vmrsyncignore

Coloque um arquivo `.vmrsyncignore` na pasta sincronizada (mesmo formato do rsync, um padrão por linha). Os padrões são carregados automaticamente a cada sync, junto com `--exclude-from` e `--exclude`.

```bash
# $VMRSYNC_PATH/project1/.vmrsyncignore
node_modules/
*.log
.git/
```

#### Verificação prévia

Antes da primeira sincronização, verifique SSH e caminhos:

```bash
vmrsync check vm21 project1
```

### Primeira Sincronização

Sempre teste com `--dry-run` primeiro:

```bash
# Testar sincronização do remoto para local
vmrsync restore vm21 project1 --dry-run

# Testar sincronização do local para remoto
vmrsync backup vm21 project1 --dry-run

# Executar sincronização real
vmrsync restore vm21 project1
```

## Uso Avançado

### Modos de Sincronização

#### Modo Espelhamento (Padrão)

Replica sua estrutura de diretórios local exatamente na máquina remota.

```bash
# Local: /home/user/Sources/project1/
# Remoto: /home/user/Sources/project1/

vmrsync backup vm21 project1
```

**Casos de uso:**
- Ambientes de desenvolvimento que espelham produção
- Trabalhar com múltiplas VMs idênticas
- Manter estruturas de diretórios consistentes

#### Modo Staging

Sincroniza tudo para `/vmrsync` independentemente do caminho local.

```bash
# Local: /home/user/Sources/project1/
# Remoto: /vmrsync/project1/

vmrsync backup vm21 project1 --staging
```

**Casos de uso:**
- Ambiente de teste centralizado
- Workspace compartilhado entre membros da equipe
- Staging temporário antes do deployment

### Filtragem Avançada

#### Múltiplos Padrões de Exclusão

```bash
vmrsync backup vm21 project1 \
  --exclude "*.log" \
  --exclude "*.tmp" \
  --exclude "node_modules/" \
  --exclude ".git/" \
  --exclude "*.pyc" \
  --exclude "__pycache__/"
```

#### Exclusões Específicas por Linguagem

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

### Otimização de Performance

#### Manipulação de Arquivos Grandes

```bash
# Use --no-delete para evitar perda acidental de dados
vmrsync backup vm21 project1 --no-delete

# Visualize o que será sincronizado
vmrsync backup vm21 project1 --dry-run --verbose

# Sincronize múltiplos diretórios menores em vez de um grande
vmrsync backup vm21 src --exclude "assets/"
vmrsync backup vm21 assets
```

#### Otimização de Rede

**Conexões lentas:**
```bash
vmrsync restore vm21 project1 \
  --exclude "*.log" \
  --exclude "*.tmp" \
  --exclude "node_modules/"
```

**Rede local rápida:**
```bash
vmrsync backup vm21 project1 --verbose
```

### Automação e Scripts

#### Hook Pre-commit

Crie `.git/hooks/pre-commit`:
```bash
#!/bin/bash
BRANCH=$(git rev-parse --abbrev-ref HEAD)

if [ "$BRANCH" = "main" ]; then
    vmrsync backup vm21 . --exclude ".git/" --exclude "*.log"
fi
```

Torne executável:
```bash
chmod +x .git/hooks/pre-commit
```

#### Hook Post-commit

Crie `.git/hooks/post-commit`:
```bash
#!/bin/bash
vmrsync backup vm21 . --staging --exclude ".git/"
```

#### Tarefas Cron

```bash
# Sincronizar a cada 15 minutos (quiet para cron)
*/15 * * * * $HOME/.local/bin/vmrsync restore vm21 project1 --quiet

# Sincronizar a cada hora com logging
0 * * * * $HOME/.local/bin/vmrsync restore vm21 project1 >> $HOME/vmrsync.log 2>&1

# Sincronizar todas as manhãs às 8 AM
0 8 * * * $HOME/.local/bin/vmrsync restore vm21 project1
```

### Workflows com Múltiplas Máquinas

#### Sincronização em Múltiplas VMs

```bash
#!/bin/bash
MACHINES="vm21 vm22 vm23"
PROJECT="project1"

for machine in $MACHINES; do
    echo "Sincronizando para $machine..."
    vmrsync backup "$machine" "$PROJECT" --exclude "*.log"
done
```

#### Teste Round-Robin

```bash
#!/bin/bash
MACHINES="vm21 vm22 vm23"
PROJECT="project1"

for machine in $MACHINES; do
    echo "Executando testes em $machine..."
    ssh "$machine" "cd /vmrsync/$PROJECT && make test"
done
```

### Backup e Recuperação

#### Estratégia de Backup

```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="$HOME/backups/vmrsync/$DATE"

mkdir -p "$BACKUP_DIR"
vmrsync restore vm21 . --dry-run | tee "$BACKUP_DIR/rsync-preview.log"

read -p "Prosseguir com backup? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    vmrsync restore vm21 . | tee "$BACKUP_DIR/rsync.log"
    echo "Backup completado: $BACKUP_DIR"
fi
```

### Considerações de Segurança

#### Gerenciamento de Chaves SSH

```bash
# Chave de desenvolvimento
vmrsync backup dev-server project1 --ssh-key ~/.ssh/id_dev

# Chave de produção
vmrsync backup prod-server project1 --ssh-key ~/.ssh/id_prod
```

#### Segurança de Rede

O `vmrsync` recusa alvos que se referem a esta máquina: `localhost` (e variantes comuns), endereços de loopback (`127.0.0.1`, `::1`, etc.), o hostname local (`os.Hostname()`), e qualquer endereço atribuído a uma interface não-loopback da máquina onde o comando está rodando. Não existe flag para bypass. Isso evita sincronizações acidentais (principalmente com `--delete`) contra o próprio host.

Use um **hostname/IP remoto real** como argumento `machine` (não `localhost`). Se você acessa a VM por bastion/jump host, configure **`ProxyJump`** ou **`ProxyCommand`** no `~/.ssh/config`. O `vmrsync` invoca `ssh`/`rsync` de forma padrão, então essas entradas se aplicam automaticamente quando você usa o alias `Host`:

```bash
# Exemplo ~/.ssh/config:
# Host minha-vm
#   HostName 10.0.0.50
#   User dev
#   ProxyJump user@jump-server

vmrsync backup minha-vm project1
```

Para um destino fixo na próxima máquina (em `/vmrsync` após `vmrsync prepare`), use **`--staging`** como documentado acima — as mesmas regras de segurança do host ainda valem.

#### Notas de robustez do rsync

- Todas as invocações SSH usam `-o BatchMode=yes` e `-o ConnectTimeout=10` para falhar rápido em vez de pedir senha.
- Falhas de conexão SSH são reportadas separadamente de diretório remoto inexistente.
- O `vmrsync` passa `--protect-args` ao rsync para evitar problemas de interpretação de argumentos no remoto.
- O `vmrsync` usa `--mkpath` quando disponível; em versões antigas do rsync faz fallback criando o caminho remoto via SSH no `backup`.
- `--timeout-seconds` define um limite rígido de tempo do rsync (padrão: 7200 segundos; `0` desativa).

## Solução de Problemas

### Problemas de Instalação

#### Binário Não Encontrado

```bash
# Verificar se o binário existe
ls -la $HOME/.local/bin/vmrsync

# Adicionar ao PATH se necessário
export PATH="$HOME/.local/bin:$PATH"

# Recarregar shell
source ~/.bashrc
```

#### Falha na Compilação

```bash
# Verificar instalação do Go
go version

# Limpar e recompilar
make clean
make build
```

### Problemas de Conexão

#### Conexão SSH Recusada

```bash
# Testar SSH independentemente
ssh <maquina>

# Verificar se servidor SSH está rodando
ssh <maquina> "systemctl status ssh"

# Tentar com porta personalizada
vmrsync restore <maquina> project1 --ssh-port 2222
```

#### Autenticação Falhou

```bash
# Testar SSH com saída verbose
ssh -v <maquina>

# Verificar permissões da chave SSH
ls -la ~/.ssh/id_rsa
chmod 600 ~/.ssh/id_rsa

# Copiar chave pública para remoto
ssh-copy-id <maquina>

# Especificar chave personalizada
vmrsync restore <maquina> project1 --ssh-key ~/.ssh/chave_personalizada
```

#### Timeout de Conexão

```bash
# Testar conectividade de rede
ping <maquina>

# Verificar resolução DNS
nslookup <maquina>

# Tentar usar endereço IP
vmrsync restore 192.168.1.100 project1
```

### Erros de Sincronização

#### Diretório Não Existe

```bash
# Para modo staging
vmrsync prepare <maquina>

# Para modo espelhamento, criar manualmente
ssh <maquina> "mkdir -p $HOME/Sources/project1"

# Verificar se diretório existe
ssh <maquina> "ls -la $HOME/Sources/"
```

#### rsync Não Encontrado

```bash
# Instalar rsync no local
sudo apt-get install rsync  # Ubuntu/Debian
sudo yum install rsync      # CentOS/RHEL

# Instalar rsync no remoto
ssh <maquina> "sudo apt-get install rsync"
```

#### Deleções Inesperadas de Arquivos

```bash
# Sempre teste com dry-run primeiro
vmrsync backup <maquina> project1 --dry-run --verbose

# Use --no-delete para prevenir deleções
vmrsync backup <maquina> project1 --no-delete

# Verifique direção
vmrsync restore <maquina> project1  # Remoto -> Local
vmrsync backup <maquina> project1  # Local -> Remoto
```

### Problemas de Performance

#### Sincronização Muito Lenta

```bash
# Verifique o que está sendo transferido
vmrsync backup <maquina> project1 --dry-run --verbose

# Exclua arquivos desnecessários
vmrsync backup <maquina> project1 \
  --exclude "*.log" \
  --exclude "*.tmp" \
  --exclude "node_modules/" \
  --exclude ".git/"

# Sincronize subdiretórios separadamente
vmrsync backup <maquina> project1/src
vmrsync backup <maquina> project1/tests
```

#### Alto Uso de CPU

```bash
# Reduza verbosidade
vmrsync backup <maquina> project1  # sem --verbose

# Agende fora do horário de pico
# Adicione ao crontab: 0 2 * * * vmrsync backup <maquina> project1

# Exclua mais arquivos
vmrsync backup <maquina> project1 \
  --exclude "*.log" \
  --exclude "build/" \
  --exclude "dist/"
```

### Problemas de Permissão

#### Permissão Negada no Remoto

```bash
# Verificar permissões do diretório remoto
ssh <maquina> "ls -la $HOME/Sources/project1"

# Corrigir permissões
ssh <maquina> "chown -R $USER:$USER $HOME/Sources/project1"
ssh <maquina> "chmod -R u+rw $HOME/Sources/project1"

# Para modo staging, execute prepare
vmrsync prepare <maquina>
```

#### Erro de Sudo Requerido

```bash
# Execute prepare manualmente
ssh <maquina> "sudo mkdir -p /vmrsync && sudo chown $UID:$(id -g) /vmrsync"

# Ou use modo espelhamento
vmrsync backup <maquina> project1  # sem --staging
```

### Obtendo Ajuda

#### Coletar Informações de Diagnóstico

```bash
# Informações do sistema
uname -a
go version
rsync --version
ssh -V

# Informações do VM RSync
vmrsync version

# Teste de rede
ping -c 4 <maquina>
ssh -v <maquina> "echo 'Conexão SSH funciona'"

# Dry run
vmrsync backup <maquina> project1 --dry-run --verbose

# Variáveis de ambiente
echo "VMRSYNC_PATH=$VMRSYNC_PATH"
```

#### Habilitar Logging

```bash
# Redirecionar saída para arquivo de log
vmrsync backup <maquina> project1 --verbose > vmrsync.log 2>&1
```

### Mensagens de Erro Comuns

| Erro | Causa Comum | Solução |
|-------|-------------|---------|
| `command not found` | VM RSync não está no PATH | Adicione `$HOME/.local/bin` ao PATH |
| `connection refused` | Servidor SSH não está rodando | Inicie servidor SSH no remoto |
| `permission denied` | Chave SSH não configurada | Configure autenticação com chave SSH |
| `directory does not exist` | Diretório remoto ausente | Execute `vmrsync prepare <maquina>` |
| `rsync: command not found` | rsync não instalado | Instale rsync no local/remoto |
| `timeout` | Problemas de rede | Verifique conectividade de rede |

### Melhores Práticas

1. **Sempre teste com --dry-run primeiro**
2. **Use padrões de exclusão apropriados**
3. **Backups regulares antes de mudanças importantes**
4. **Monitore operações de sincronização com --verbose**
5. **Mantenha chaves SSH seguras (chmod 600)**
6. **Documente seu workflow de sincronização**
7. **Teste procedimentos de recuperação de desastres**
8. **Mantenha software atualizado**

## Recursos Adicionais

- [README Principal](../README-PT.md)
- [Repositório GitHub](https://github.com/carlosrabelo/vmrsync)
- [Documentação do rsync](https://linux.die.net/man/1/rsync)
- [Documentação do SSH](https://www.openssh.com/manual.html)

---

*Última atualização: Abril 2026*