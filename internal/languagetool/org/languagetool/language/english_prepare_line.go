package language

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// EnglishPrepareLineForSpeller ports English.prepareLineForSpeller
// (twin of rules/en.PrepareLineForSpeller for Language API surface).
// Only NN*/JJ* tagged forms kept; '+' lines dropped (morfologik separator).
func EnglishPrepareLineForSpeller(line string) []string {
	parts := strings.Split(line, "#")
	if len(parts) == 0 {
		return []string{line}
	}
	if strings.Contains(line, "+") {
		// while the morfologik separator is "+", multiwords with '+' can cause undesired results.
		return []string{""}
	}
	formTag := strings.Split(parts[0], "\t")
	// Java: formTag[i].trim()
	form := tools.JavaStringTrim(formTag[0])
	if len(formTag) > 1 {
		tag := tools.JavaStringTrim(formTag[1])
		if strings.HasPrefix(tag, "NN") || strings.HasPrefix(tag, "JJ") {
			return []string{form}
		}
		return []string{""}
	}
	return []string{line}
}

// EnglishHasMinMatchesRules ports English.hasMinMatchesRules (true).
func EnglishHasMinMatchesRules() bool { return true }
