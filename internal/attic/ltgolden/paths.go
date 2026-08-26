package ltgolden

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/attic/data"
)

// AllRuleXMLPaths returns every rules XML under language modules (grammar, style, punctuation, variants).
func AllRuleXMLPaths(dataRoot string) []string {
	root := data.LanguageModules(dataRoot)
	var paths []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".xml") {
			return nil
		}
		// only under .../rules/...
		if !strings.Contains(filepath.ToSlash(path), "/rules/") {
			return nil
		}
		// skip print/xsl related if any
		name := d.Name()
		if name == "print.xml" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths
}

// AllDisambiguationXMLPaths returns every disambiguation.xml in language modules + core xx test.
func AllDisambiguationXMLPaths(dataRoot string) []string {
	var paths []string
	// language modules
	root := data.LanguageModules(dataRoot)
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if d.Name() == "disambiguation.xml" {
			paths = append(paths, path)
		}
		return nil
	})
	// core demo
	xx := filepath.Join(dataRoot, "languagetool-core", "src", "test", "resources",
		"org", "languagetool", "resource", "xx", "disambiguation.xml")
	if _, err := os.Stat(xx); err == nil {
		paths = append(paths, xx)
	}
	return paths
}

// AllJavaTestPaths returns every *Test.java under LT sources.
func AllJavaTestPaths(dataRoot string) []string {
	var paths []string
	_ = filepath.WalkDir(dataRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), "Test.java") {
			return nil
		}
		if !strings.Contains(filepath.ToSlash(path), "/src/test/java/") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths
}

// EnglishGrammarPaths kept for compatibility.
func EnglishGrammarPaths(dataRoot string) []string {
	var out []string
	for _, p := range AllRuleXMLPaths(dataRoot) {
		if strings.Contains(filepath.ToSlash(p), "/language-modules/en/") ||
			strings.Contains(filepath.ToSlash(p), "/languagetool-language-modules/en/") {
			out = append(out, p)
		}
	}
	return out
}
