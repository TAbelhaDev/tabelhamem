package main

import (
	"fmt"
	"io"
	"os"

	"github.com/TAbelhaDev/tabelhascaff/ipc"
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

	printUsage(os.Stderr)
	os.Exit(1)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `uso: tamem ipc <método> [key=value...] --json

métodos:
  link    project=<slug> repo=<path>      liga a memória do Claude Code em
                                           <repo> a ~/agent-memory/<slug>/ e
                                           garante a seção no AGENTS.md do repo
  unlink  project=<slug> repo=<path>      desfaz o link: <repo> volta a ter
                                           uma cópia real da memória, e a
                                           seção do AGENTS.md é removida
  status  project=<slug> [repo=<path>]    estado da ponte pra um projeto
  list                                    lista todos os projetos já ligados
  search  query=<termo> [type=<tipo>]
          [project=<slug>]                busca texto nos arquivos de
                                           memória de todos os projetos`)
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
	default:
		fmt.Fprintf(os.Stderr, "método desconhecido: %q\n", parsed.Method)
		return 1
	}
}
