package hunspell

import (
	"os"
	"path/filepath"
	"strings"
)

// DiscoverHunspellDic finds a Java resource-path Hunspell .dic on disk.
// classpath is e.g. "/da/hunspell/da_DK.dic" (Java HunspellRule.getDictFilenameInResources).
// Walks inspiration language-modules resource trees. Empty if missing (no invent path).
func DiscoverHunspellDic(classpath string) string {
	rel := strings.TrimPrefix(classpath, "/")
	if rel == "" || !strings.HasSuffix(rel, ".dic") {
		return ""
	}
	lang, rest, ok := strings.Cut(rel, "/")
	if !ok || lang == "" || rest == "" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		candidates := []string{
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", lang,
				"src", "main", "resources", "org", "languagetool", "resource", lang, rest),
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", lang,
				"src", "main", "resources", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "languagetool-dicts", "org", "languagetool", "resource", rel),
		}
		for _, p := range candidates {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// companionAff returns path with .dic → .aff when that file exists.
func companionAff(dicPath string) string {
	if !strings.HasSuffix(dicPath, ".dic") {
		return ""
	}
	aff := dicPath[:len(dicPath)-4] + ".aff"
	if st, err := os.Stat(aff); err == nil && st.Mode().IsRegular() {
		return aff
	}
	return ""
}

// TryOpenFromClasspath opens a FileHunspellDictionary for a Java resource path.
// Returns nil if missing or unreadable (fail closed — same as nil Dict on HunspellRule).
// Affix file is discovered beside the .dic when present; affix rules are not expanded
// (pure-Go word-list port, incomplete vs full native Hunspell).
func TryOpenFromClasspath(classpath string) HunspellDictionary {
	dicPath := DiscoverHunspellDic(classpath)
	if dicPath == "" {
		return nil
	}
	d, err := NewFileHunspellDictionary(dicPath, companionAff(dicPath), false)
	if err != nil || d == nil {
		return nil
	}
	return d
}
