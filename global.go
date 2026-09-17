package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TAbelhaDev/tabelhascaff/ipc"
	tuiui "github.com/TAbelhaDev/tabelhatuiui"
)

const globalSlug = "global"

// globalSharedDir is ~/agent-memory/global/.
func globalSharedDir() string {
	return filepath.Join(homeDir(), "agent-memory", globalSlug)
}

// globalAGENTSPath is the OpenCode instruction file that points to the shared store.
func globalAGENTSPath() string {
	return filepath.Join(tuiui.ConfigDir(), "opencode", "AGENTS.md")
}

// claudeGlobalMemoryDir is ~/.claude/memory/ for user-level global memory.
func claudeGlobalMemoryDir() string {
	return filepath.Join(homeDir(), ".claude", "memory")
}

// ipcGlobal sets up the shared global memory store:
//   - creates ~/agent-memory/global/ with MEMORY.md
//   - migrates existing ~/.config/opencode/AGENTS.md content into the store
//   - migrates existing ~/.claude/memory/ content into the store
//   - symlinks OpenCode AGENTS.md → shared store
//   - symlinks ~/.claude/memory/ → ~/agent-memory/global/
//
// Idempotent: re-running is safe.
func ipcGlobal() int {
	shared := globalSharedDir()

	// 1. Create shared directory.
	if err := os.MkdirAll(shared, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "erro criando %s: %v\n", shared, err)
		return 1
	}

	// 2. Migrate existing AGENTS.md content into the store (only if store doesn't exist yet).
	storeAGENTS := filepath.Join(shared, "AGENTS.md")
	if _, err := os.Stat(storeAGENTS); os.IsNotExist(err) {
		if data, err := os.ReadFile(globalAGENTSPath()); err == nil && len(data) > 0 {
			if err := os.WriteFile(storeAGENTS, data, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "erro migrando AGENTS.md: %v\n", err)
				return 1
			}
		}
	}

	// 3. Ensure MEMORY.md exists.
	memoryPath := filepath.Join(shared, "MEMORY.md")
	if _, err := os.Stat(memoryPath); os.IsNotExist(err) {
		if err := os.WriteFile(memoryPath, []byte("# Global Memory\n\nRegras e contexto compartilhados entre todos os agentes.\n"), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "erro criando MEMORY.md: %v\n", err)
			return 1
		}
	}

	// 4. Migrate ~/.claude/memory/ real dir → shared store before symlinking.
	if err := migrateClaudeGlobal(claudeGlobalMemoryDir(), shared); err != nil {
		fmt.Fprintf(os.Stderr, "erro migrando Claude memory: %v\n", err)
		return 1
	}

	// 5. Symlink OpenCode AGENTS.md → shared store.
	if err := symlinkAGENTSFile(globalAGENTSPath(), storeAGENTS); err != nil {
		fmt.Fprintf(os.Stderr, "erro criando symlink OpenCode: %v\n", err)
		return 1
	}

	// 6. Symlink ~/.claude/memory/ → ~/agent-memory/global/.
	if err := symlinkClaudeGlobal(claudeGlobalMemoryDir(), shared); err != nil {
		fmt.Fprintf(os.Stderr, "erro criando symlink Claude: %v\n", err)
		return 1
	}

	return ipc.WriteJSON(map[string]string{
		"shared_dir":  shared,
		"agents_md":   globalAGENTSPath(),
		"claude_link": claudeGlobalMemoryDir(),
		"status":      "ok",
	})
}

// migrateClaudeGlobal moves contents of a real ~/.claude/memory/ dir into the
// shared store before the symlink replaces it. Files that already exist in the
// shared store are left in place.
func migrateClaudeGlobal(claudeDir, sharedDir string) error {
	info, err := os.Lstat(claudeDir)
	if err != nil {
		return nil // doesn't exist, nothing to migrate
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil // already a symlink, nothing to migrate
	}
	if !info.IsDir() {
		return nil
	}

	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return fmt.Errorf("lendo %s: %w", claudeDir, err)
	}

	for _, e := range entries {
		src := filepath.Join(claudeDir, e.Name())
		dst := filepath.Join(sharedDir, e.Name())
		if _, err := os.Stat(dst); err == nil {
			continue // already in shared store, skip
		}
		if err := moveFile(src, dst); err != nil {
			return fmt.Errorf("movendo %s: %w", e.Name(), err)
		}
	}

	return nil
}

// symlinkAGENTSFile creates or updates a symlink from path to target.
// If path already exists as a symlink to target, it's a no-op.
// Real files are backed up with a timestamp before being replaced.
func symlinkAGENTSFile(path, target string) error {
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			existing, err := os.Readlink(path)
			if err == nil && existing == target {
				return nil
			}
			os.Remove(path)
		} else if info.Mode().IsRegular() {
			// Real file — back it up with timestamp then replace with symlink.
			ts := time.Now().Format("20060102-150405")
			backup := path + ".bak." + ts
			if err := os.Rename(path, backup); err != nil {
				return fmt.Errorf("fazendo backup de %s: %w", path, err)
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.Symlink(target, path)
}

// symlinkClaudeGlobal creates or updates the symlink from claudeDir to sharedDir.
func symlinkClaudeGlobal(claudeDir, sharedDir string) error {
	if info, err := os.Lstat(claudeDir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(claudeDir)
			if err == nil && target == sharedDir {
				return nil
			}
			os.Remove(claudeDir)
		} else if info.IsDir() {
			fmt.Fprintf(os.Stderr, "aviso: %s já existe como diretório real, não sobrescrevendo\n", claudeDir)
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(claudeDir), 0o755); err != nil {
		return err
	}

	return os.Symlink(sharedDir, claudeDir)
}
