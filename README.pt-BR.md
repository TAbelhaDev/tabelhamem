<div align="center">

# tabelhamem

**Ponte de memória compartilhada entre Claude Code e OpenCode.**

[English](README.md) · **Português**

[![Go Version](https://img.shields.io/github/go-mod/go-version/TAbelhaDev/tabelhamem?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)
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
equivalente automático: só um diário global `memory.md` mantido à mão e
arquivos `AGENTS.md` por repositório.

O `tamem` faz a ponte entre os dois apontando ambos pro mesmo armazenamento
em markdown puro: `~/agent-memory/<projeto>/`, irmão de `~/jobs` (automação
do próprio usuário, não amarrada a nenhuma ferramenta específica). O
diretório de memória do Claude Code vira um symlink pra lá (transparente —
o Claude só faz I/O de arquivo normal); o OpenCode é instruído a ler/
escrever no mesmo lugar através de um bloco que o `tamem` mantém no
`AGENTS.md` do repo.

Isso também conserta de quebra uma peculiaridade do Claude Code em que cada
worktree git do mesmo repo ganha um diretório de memória próprio,
desconectado dos outros — rodar `tamem run link` de novo a partir de um
worktree aponta ele pro mesmo bucket compartilhado.

## Instalação

```bash
go install github.com/TAbelhaDev/tabelhamem@latest
```

## Uso

```bash
# Liga a memória de um projeto: migra os arquivos de memória do Claude Code
# já existentes pra ~/agent-memory/<projeto>/, faz o symlink do diretório
# do Claude Code pra lá, e escreve/atualiza as instruções no AGENTS.md do
# repo. Idempotente.
tamem ipc link project=tabelharadar repo=/home/ianptkcs/codigo/tabelhadev/tabelharadar --json

# Confere a saúde da ponte pra um projeto
tamem ipc status project=tabelharadar repo=/home/ianptkcs/codigo/tabelhadev/tabelharadar --json

# Lista todos os projetos já em ponte
tamem ipc list --json
```

## Métodos IPC

| Método | Filtros | Descrição |
|---|---|---|
| `link` | `project=`, `repo=` | Cria/atualiza a ponte de um projeto: migra + symlink + seção no AGENTS.md |
| `status` | `project=`, `repo=` (opcional) | Reporta se o symlink e a seção do AGENTS.md estão certos |
| `list` | (nenhum) | Lista todos os projetos sob `~/agent-memory/` |

## Limitações

- O OpenCode não tem mecanismo embutido pra ler `AGENTS.md` automaticamente
  como o Claude Code carrega o `MEMORY.md` sozinho — a ponte funciona tão
  bem quanto o modelo seguir essa instrução a cada sessão.
- `link` precisa ser rodado de novo por worktree git; o diretório de
  memória do Claude Code de um worktree não é tocado automaticamente só
  porque o checkout principal foi ligado.
