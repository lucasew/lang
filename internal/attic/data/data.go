package data

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultDataDir = "inspiration/languagetool"
	EnvData        = "LANG_DATA"
)

// Resolve returns the LanguageTool data root.
// Order: flag (--data-dir) > LANG_DATA > ./inspiration/languagetool
func Resolve(flagDir string) (string, error) {
	if flagDir != "" {
		return absExisting(flagDir, "flag --data-dir")
	}
	if v := os.Getenv(EnvData); v != "" {
		return absExisting(v, "env "+EnvData)
	}
	return absExisting(DefaultDataDir, "default "+DefaultDataDir)
}

func absExisting(path, source string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("data dir (%s): %w", source, err)
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("data dir not found (%s): %s\n  set --data-dir or %s, or init the submodule:\n  git submodule update --init --depth 1", source, abs, EnvData)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("data dir is not a directory (%s): %s", source, abs)
	}
	// Sanity: expect LanguageTool layout.
	if _, err := os.Stat(filepath.Join(abs, "languagetool-language-modules")); err != nil {
		return "", fmt.Errorf("data dir does not look like LanguageTool (%s): missing languagetool-language-modules under %s", source, abs)
	}
	return abs, nil
}

// CoreResources is languagetool-core resources root (MessagesBundle, etc.).
func CoreResources(dataRoot string) string {
	return filepath.Join(dataRoot, "languagetool-core", "src", "main", "resources")
}

// LanguageModules is languagetool-language-modules root.
func LanguageModules(dataRoot string) string {
	return filepath.Join(dataRoot, "languagetool-language-modules")
}
