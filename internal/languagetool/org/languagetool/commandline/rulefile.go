package commandline

import (
	"fmt"
	"os"
	"strings"
)

// LoadRuleFile reads an additional grammar/rule file for --rulefile.
// Returns raw content for hooks to parse; validates non-empty path and readable file.
func LoadRuleFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty rulefile path")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// InferLanguageFromRuleFileName ports Java hint: if filename contains a known language code.
func InferLanguageFromRuleFileName(path string) string {
	base := path
	if i := strings.LastIndexAny(path, `/\`); i >= 0 {
		base = path[i+1:]
	}
	lower := strings.ToLower(base)
	for _, code := range []string{"en-us", "en-gb", "en", "de", "fr", "pl", "uk", "es", "pt", "nl", "it", "ca"} {
		if strings.Contains(lower, code) {
			// normalize short
			if strings.HasPrefix(code, "en") {
				return "en"
			}
			return code
		}
	}
	return ""
}
