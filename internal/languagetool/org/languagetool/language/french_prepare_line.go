package language

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// frenchSpellerExceptions ports French.spellerExceptions (Java singletonList "Ho Chi Minh").
var frenchSpellerExceptions = map[string]struct{}{
	"Ho Chi Minh": {},
}

// FrenchPrepareLineForSpeller ports French.prepareLineForSpeller.
// form\t;tag — keep Z* / N* / A; drop other tags; exception forms → empty.
func FrenchPrepareLineForSpeller(line string) []string {
	parts := strings.Split(line, "#")
	if len(parts) == 0 {
		return []string{line}
	}
	formTag := regexp.MustCompile(`[\t;]`).Split(parts[0], -1)
	// Java: formTag[i].trim()
	form := tools.JavaStringTrim(formTag[0])
	if _, bad := frenchSpellerExceptions[form]; bad {
		return []string{""}
	}
	if len(formTag) > 1 {
		tag := tools.JavaStringTrim(formTag[1])
		if strings.HasPrefix(tag, "Z") || strings.HasPrefix(tag, "N") || tag == "A" {
			return []string{form}
		}
		return []string{""}
	}
	return []string{line}
}

// FrenchHasMinMatchesRules ports French.hasMinMatchesRules (true).
func FrenchHasMinMatchesRules() bool { return true }
