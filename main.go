package main

import (
	"fmt"
	"io"
	"os"

	"github.com/TAbelhaDev/tabelhascaff/ipc"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "ipc":
			os.Exit(runIPC(os.Args[2:]))
		case "--help", "-h":
			printUsage(os.Stdout)
			os.Exit(0)
		}
	}

	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `uso: tamem ipc <método> [key=value...] --json

métodos:
  global                                configura o armazenamento global
                                        compartilhado: cria ~/agent-memory/global/,
                                        migra o AGENTS.md existente, cria symlink
                                        do Claude Code e atualiza o pointer do
                                        OpenCode
  link    project=<slug> repo=<path>    liga a memória compartilhada do
                                        projeto em <repo> a
                                        ~/agent-memory/<slug>/ (Claude
                                        Code via symlink; OpenCode via
                                        AGENTS.md)
  unlink  project=<slug> repo=<path>    desfaz o link: <repo> volta a ter
                                        uma cópia real da memória, e a
                                        seção do AGENTS.md é removida
  status  project=<slug> [repo=<path>]  estado da ponte pra um projeto
  list                                  lista todos os projetos já ligados
  search  query=<termo> [type=<tipo>]
          [project=<slug>]              busca texto nos arquivos de
                                        memória de todos os projetos
  health                                resumo markdown da saúde de todos
                                        os projetos (para workflows taglue)
  search-digest query=<termo> [type=<tipo>]
          [project=<slug>]              busca e devolve markdown formatado
                                        (para workflows taglue)`)
}

// runIPC implements `tamem ipc <method> [key=value...] --json`.
func runIPC(args []string) int {
	parsed, err := ipc.ParseIPCArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "uso: tamem ipc <método> [key=value...] --json")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	switch parsed.Method {
	case "global":
		return ipcGlobal()
	case "link":
		return ipcLink(parsed.Filters)
	case "unlink":
		return ipcUnlink(parsed.Filters)
	case "status":
		return ipcStatus(parsed.Filters)
	case "list":
		return ipcList(parsed.Filters)
	case "search":
		return ipcSearch(parsed.Filters)
	case "health":
		return ipcHealth(parsed.Filters)
	case "search-digest":
		return ipcSearchDigest(parsed.Filters)
	default:
		fmt.Fprintf(os.Stderr, "método desconhecido: %q\n", parsed.Method)
		return 1
	}
}
