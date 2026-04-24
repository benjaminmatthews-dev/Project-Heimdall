// Package runner provides ScriptServer-compatible v1 runner discovery and parsing.
package runner

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Runner represents a ScriptServer-compatible v1 runner definition.
type Runner struct {
	Name             string `json:"name"`
	ScriptPath       string `json:"script_path"`
	Description      string `json:"description"`
	RequiresTerminal bool   `json:"requires_terminal"`
	// Group is derived from the immediate subdirectory under the runners root;
	// not persisted in JSON.
	Group string `json:"-"`
}

// LoadFromDir recursively discovers all .json runner files under root and
// groups them by the name of their immediate subdirectory relative to root.
// Files at the root level are assigned to the "default" group.
func LoadFromDir(root string) (map[string][]Runner, error) {
	groups := make(map[string][]Runner)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		var r Runner
		if err := json.Unmarshal(data, &r); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		if r.Name == "" || r.ScriptPath == "" {
			// Skip malformed runner files silently.
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		parts := strings.SplitN(rel, string(filepath.Separator), 2)
		group := "default"
		if len(parts) > 1 {
			group = parts[0]
		}
		r.Group = group
		groups[group] = append(groups[group], r)
		return nil
	})

	return groups, err
}

// SortedGroups returns group names in deterministic sorted order.
func SortedGroups(groups map[string][]Runner) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
