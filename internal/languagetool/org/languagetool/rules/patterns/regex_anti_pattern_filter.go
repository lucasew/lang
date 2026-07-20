package patterns

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegexAntiPatternFilter ports org.languagetool.rules.patterns.RegexAntiPatternFilter.
// Limitations (same as Java): antipatterns cannot contain spaces; | delimits patterns.
type RegexAntiPatternFilter struct{}

func (RegexAntiPatternFilter) AcceptRegexMatch(match *rules.RuleMatch, arguments map[string]string, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	antiPatternStr, ok := arguments["antipatterns"]
	if !ok {
		panic("Missing 'antiPatterns:' in 'args' in <filter> of rule")
	}
	antiPatterns := splitAntiPatterns(antiPatternStr)
	text := ""
	if sentence != nil {
		text = sentence.GetText()
	}
	for _, ap := range antiPatterns {
		// Java Pattern.compile throws on bad syntax (not swallowed).
		re, err := regexp.Compile(ap)
		if err != nil {
			panic(err)
		}
		locs := re.FindAllStringIndex(text, -1)
		for _, loc := range locs {
			start, end := loc[0], loc[1]
			// partial overlap is enough to filter out a match
			if (start <= match.GetToPos() && end >= match.GetToPos()) ||
				(start <= match.GetFromPos() && end >= match.GetFromPos()) {
				return nil
			}
		}
	}
	return match
}

func splitAntiPatterns(s string) []string {
	// Java: antiPatternStr.split("\\|") — no empty discard special
	return regexp.MustCompile(`\|`).Split(s, -1)
}
