package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// LoadDotenv loads backend environment from disk before other packages read os.Getenv.
//
// Resolution order:
//  1. BACKEND_ENV_FILE — explicit path (absolute or relative to the process cwd)
//  2. {cwd}/backend/.env — when the process is started from the repository root
//  3. {cwd}/.env — when the process is started from the backend module directory
//
// When multiple files from (2) and (3) exist, the first path wins per key (same as godotenv.Load),
// so backend/.env takes precedence over a sibling .env at the repo root.
func LoadDotenv() {
	if p := strings.TrimSpace(os.Getenv("BACKEND_ENV_FILE")); p != "" {
		_ = godotenv.Load(p)
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		_ = godotenv.Load()
		return
	}
	var paths []string
	for _, p := range []string{
		filepath.Join(wd, "backend", ".env"),
		filepath.Join(wd, ".env"),
	} {
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			continue
		}
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		return
	}
	_ = godotenv.Load(paths...)
}
