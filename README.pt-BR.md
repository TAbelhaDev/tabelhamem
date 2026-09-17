<div align="center">

# tabelhamem

**Ponte de memória compartilhada entre Claude Code e OpenCode.**

[English](README.md) · **Português**

[![Go Version](https://img.shields.io/github/go-mod/go-version/TAbelhaDev/tabelhamem?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)
[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![Powered by tabelhatuiui](https://img.shields.io/badge/theme-tabelhatuiui-d6b4f7?style=flat-square)](https://github.com/TAbelhaDev/tabelhatuiui)
[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)

</div>

---

## Por quê

O Claude Code tem um sistema de memória automático e embutido: injeta um
índice `MEMORY.md` por projeto mais arquivos-tópico tipados
(`feedback_*.md`/`project_*.md`/`reference_*.md`/`user_*.md`, com
frontmatter YAML) no início de cada sessão, guardados em
`~/.claude/projects/<cwd-escapado>/memory/` — um diretório por diretório de
trabalho exato, sem opção de configurar outro lugar. O OpenCode não tem
diretório de memória por projeto, mas carrega o `AGENTS.md` no início de
cada sessão — o `tamem` usa isso pra ensinar a ele o mesmo armazenamento.

O `tamem` faz a ponte entre os dois apontando ambos pro mesmo armazenamento
em markdown puro: `~/agent-memory/<projeto>/`, irmão de `~/jobs` (automação
do próprio usuário, não amarrada a nenhuma ferramenta específica). O
diretório de memória do Claude Code vira um symlink pra lá (transparente —
o Claude só faz I/O de arquivo normal); o OpenCode lê/escreve no mesmo lugar
através de um bloco que o `tamem` mantém no `AGENTS.md` do repo (carregado
automaticamente pelo opencode a cada sessão).

Quando o repo é um repositório git, o `tamem` detecta automaticamente todos
os worktrees via `git worktree list` e liga, desfaz ou confere a saúde da
ponte pra cada um deles em uma única invocação, sem precisar rodar o
comando por worktree.

## Instalação

```bash
go install github.com/TAbelhaDev/tabelhamem/cmd/tamem@latest
```

### Desenvolvimento local

Um hook `post-commit` em `.githooks/` reconstrói e reinstala o `tamem` em
`~/.local/bin/tamem` a cada commit, então o comando local nunca fica
desatualizado. O git não ativa o `.githooks/` de um repo sozinho — rode isso
uma vez por clone:

```bash
git config core.hooksPath .githooks
```

## Uso

Rodar `tamem` sem argumentos abre a TUI interativa pra navegar, buscar,
ligar e desligar projetos. O subcomando `ipc` continua disponível pra
scripts.

### TUI

```bash
# Abre a TUI interativa
tamem

# Configure seus projetos em ~/.config/tabelhamem/config.toml
cat <<'EOF'
[[projects]]
slug = "tabelharadar"
repo = "/home/ianptkcs/codigo/tabelhadev/tabelharadar"

[layout]
sidebar_width_share = 1
right_width_share = 4
stats_height_share = 1
memory_height_share = 4

[general]
editor = ""
EOF
```

Atalhos (reconfiguráveis em `~/.config/tabelhamem/keybindings.json`):

| Tecla | Ação |
|---|---|
| `q` | Sair |
| `?` | Ajuda |
| `,` | Reconfigurar atalhos |
| `r` | Rescan projetos |
| `ctrl+shift+r` | Recarregar config |
| `ctrl+h` / `ctrl+l` | Navegar painéis |
| `j` / `k` | Mover cursor / rolar conteúdo |
| `enter` | Abrir arquivo no editor |
| `/` | Buscar memória |
| `l` | Ligar projeto |
| `u` | Desligar projeto |
| `esc` | Voltar / sair |

### IPC (JSON scriptável)

```bash
# Liga a memória de um projeto: migra os arquivos de memória do Claude Code
# já existentes pra ~/agent-memory/<projeto>/, faz o symlink do diretório
# do Claude Code pra lá, e escreve/atualiza as instruções no AGENTS.md do
# repo. Idempotente.
tamem ipc link project=tabelharadar repo=/home/ianptkcs/codigo/tabelhadev/tabelharadar --json

# Desfaz: o diretório de memória do Claude Code em <repo> volta a ter uma
# cópia real do conteúdo compartilhado (o diretório compartilhado em si não
# é tocado), e a seção do AGENTS.md é removida.
tamem ipc unlink project=tabelharadar repo=/home/ianptkcs/codigo/tabelhadev/tabelharadar --json

# Confere a saúde da ponte pra um projeto
tamem ipc status project=tabelharadar repo=/home/ianptkcs/codigo/tabelhadev/tabelharadar --json

# Lista todos os projetos já em ponte
tamem ipc list --json

# Busca na memória de todos os projetos já ligados
tamem ipc search query=worktree type=feedback --json
```

## Métodos IPC

| Método | Filtros | Descrição |
|---|---|---|
| `global` | (nenhum) | Configura o armazenamento global compartilhado: cria `~/agent-memory/global/`, migra o AGENTS.md existente, cria symlink do Claude Code e atualiza o pointer do OpenCode |
| `link` | `project=`, `repo=` | Cria/atualiza a ponte de um projeto: migra + symlink + seção no AGENTS.md |
| `unlink` | `project=`, `repo=` | Desfaz o `link` de um repo: restaura um diretório real, remove a seção do AGENTS.md, não mexe no diretório compartilhado |
| `status` | `project=`, `repo=` (opcional) | Reporta se o symlink e a seção do AGENTS.md estão certos |
| `list` | (nenhum) | Lista todos os projetos sob `~/agent-memory/` |
| `search` | `query=`, `type=` (opcional), `project=` (opcional) | Busca texto em todos os projetos já ligados |

## Limitações

- O lado OpenCode da ponte é dirigido por instrução: o opencode carrega o
  bloco do `AGENTS.md`, mas ler/escrever de fato o armazenamento
  compartilhado ainda depende do modelo seguir essa instrução a cada
  sessão.
- A detecção de worktrees requer `git` no `$PATH`. Se `git` não estiver
  disponível, o `tamem` opera em um único diretório (o caminho de `repo=`).
- Depois de um `unlink`, rodar `link` de novo recusa sobrescrever se a
  cópia local divergiu do compartilhado nesse meio tempo (mesma checagem de
  segurança de um diretório nunca ligado) — resolva à mão (compare os dois,
  remova a cópia local só depois de confirmar que nada se perde).
