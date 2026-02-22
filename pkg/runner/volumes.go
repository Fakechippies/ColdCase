package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const containerBase = "/data"

// detectMounts scans args for existing host paths, generates "-v host:container"
// mount strings, and returns remapped args with container-side paths substituted.
func detectMounts(args []string) (mounts []string, remapped []string) {
	seen := map[string]string{} // hostDir → containerDir

	for _, arg := range args {
		// Strip flag prefix so "-f /path/to/file" is handled too.
		val := strings.TrimLeft(arg, "-")
		if !looksLikePath(val) {
			remapped = append(remapped, arg)
			continue
		}

		abs, err := filepath.Abs(val)
		if err != nil || !pathExists(abs) {
			remapped = append(remapped, arg)
			continue
		}

		// Mount the parent directory; remap the arg to the container path.
		dir := filepath.Dir(abs)
		base := filepath.Base(abs)

		if _, ok := seen[dir]; !ok {
			idx := len(seen)
			containerDir := fmt.Sprintf("%s/vol%d", containerBase, idx)
			seen[dir] = containerDir
			mounts = append(mounts, fmt.Sprintf("%s:%s:ro", dir, containerDir))
		}

		containerPath := filepath.Join(seen[dir], base)
		// Preserve any leading flag prefix.
		prefix := arg[:len(arg)-len(val)]
		remapped = append(remapped, prefix+containerPath)
	}
	return mounts, remapped
}

func looksLikePath(s string) bool {
	return strings.HasPrefix(s, "/") ||
		strings.HasPrefix(s, "./") ||
		strings.HasPrefix(s, "../") ||
		strings.Contains(s, string(os.PathSeparator))
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
