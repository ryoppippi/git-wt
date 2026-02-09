package cmd

import (
	"encoding/json"
	"io"

	"github.com/k1LoW/git-wt/internal/git"
)

type worktreeJSON struct {
	Path    string `json:"path"`
	Branch  string `json:"branch"`
	Head    string `json:"head"`
	Bare    bool   `json:"bare"`
	Current bool   `json:"current"`
}

func printJSON(w io.Writer, worktrees []git.Worktree, currentPath string) error {
	items := make([]worktreeJSON, len(worktrees))
	for i, wt := range worktrees {
		items[i] = worktreeJSON{
			Path:    wt.Path,
			Branch:  wt.Branch,
			Head:    wt.Head,
			Bare:    wt.Bare,
			Current: wt.Path == currentPath,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}
