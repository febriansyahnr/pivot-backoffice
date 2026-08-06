package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EnsureTmpDir(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		return os.MkdirAll(dirName, 0750)
	}
	return nil
}

func FindProjectRoot(startPath, projectName string) (string, error) {
	dir := startPath
	for {
		if matchProjectDir(dir, projectName) {
			return dir, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", fmt.Errorf("project root not found")
		}
		dir = parentDir
	}
}

// matchProjectDir reports whether dir is the project root for projectName.
// Besides an exact directory name match, it also recognises worktree-style
// checkout directories whose name is "<projectName>" followed by a separator
// (".", "-", or "_") and a suffix — for example "backend-portal.my-branch".
func matchProjectDir(dir, projectName string) bool {
	base := filepath.Base(dir)
	if base == projectName {
		return true
	}
	for _, sep := range []string{".", "-", "_"} {
		if strings.HasPrefix(base, projectName+sep) {
			return true
		}
	}
	return false
}
