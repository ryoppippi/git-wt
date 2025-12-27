package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	configKeyBaseDir = "wt.basedir"
)

// GetConfig retrieves a git config value.
func GetConfig(ctx context.Context, key string) (string, error) {
	cmd, err := gitCommand(ctx, "config", "--get", key)
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		// git config returns exit code 1 if key is not found
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRepoRoot returns the root directory of the current git repository (or worktree).
func GetRepoRoot(ctx context.Context) (string, error) {
	cmd, err := gitCommand(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetMainRepoRoot returns the root directory of the main git repository.
// Unlike GetRepoRoot, this returns the main repository root even when called from a worktree.
func GetMainRepoRoot(ctx context.Context) (string, error) {
	cmd, err := gitCommand(ctx, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	gitCommonDir := strings.TrimSpace(string(out))

	// If git-common-dir is relative (e.g., ".git"), resolve it from current repo root
	if !filepath.IsAbs(gitCommonDir) {
		repoRoot, err := GetRepoRoot(ctx)
		if err != nil {
			return "", err
		}
		gitCommonDir = filepath.Join(repoRoot, gitCommonDir)
	}

	// The main repo root is the parent of the .git directory
	return filepath.Dir(gitCommonDir), nil
}

// GetRepoName returns the name of the current git repository (directory name).
func GetRepoName(ctx context.Context) (string, error) {
	root, err := GetRepoRoot(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Base(root), nil
}

// GetBaseDir returns the base directory pattern for worktrees.
// It checks git config (local, then global) and falls back to default.
// Note: This returns the raw pattern. Use GetWorktreePath to get the full path with branch expanded.
func GetBaseDir(ctx context.Context) (string, error) {
	// Check git config
	baseDir, err := GetConfig(ctx, configKeyBaseDir)
	if err != nil {
		return "", err
	}

	// If not set, use default
	if baseDir == "" {
		baseDir = "../{gitroot}-wt"
	}

	return baseDir, nil
}

// expandTemplate expands template variables in a string.
// Supported variables:
//   - {gitroot}: repository root directory name
func expandTemplate(ctx context.Context, s string) (string, error) {
	// Expand {gitroot}
	if strings.Contains(s, "{gitroot}") {
		repoName, err := GetRepoName(ctx)
		if err != nil {
			return "", err
		}
		s = strings.ReplaceAll(s, "{gitroot}", repoName)
	}

	return s, nil
}

// ExpandPath expands ~ to home directory and resolves relative paths.
// Relative paths are resolved from the main repository root, not the current worktree.
func ExpandPath(ctx context.Context, path string) (string, error) {
	// Expand ~
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	} else if path == "~" {
		return os.UserHomeDir()
	}

	// If already absolute, return as is
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	// Resolve relative path from main repo root (not current worktree)
	repoRoot, err := GetMainRepoRoot(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(repoRoot, path)), nil
}

// GetWorktreeBaseDir returns the expanded base directory for worktrees.
func GetWorktreeBaseDir(ctx context.Context) (string, error) {
	baseDir, err := GetBaseDir(ctx)
	if err != nil {
		return "", err
	}

	// Expand template variables
	baseDir, err = expandTemplate(ctx, baseDir)
	if err != nil {
		return "", err
	}

	// Expand path (~ and relative paths)
	baseDir, err = ExpandPath(ctx, baseDir)
	if err != nil {
		return "", err
	}

	return baseDir, nil
}

// GetWorktreePath returns the full path for a worktree given a branch name.
func GetWorktreePath(ctx context.Context, branch string) (string, error) {
	baseDir, err := GetWorktreeBaseDir(ctx)
	if err != nil {
		return "", err
	}

	return filepath.Join(baseDir, branch), nil
}
