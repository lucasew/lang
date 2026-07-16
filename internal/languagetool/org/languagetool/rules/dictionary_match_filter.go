package rules

// DictionaryMatchFilter ports org.languagetool.rules.DictionaryMatchFilter.
// Drops matches whose covered text is in the user-accepted dictionary.
type DictionaryMatchFilter struct {
	AcceptedWords map[string]struct{}
}

func NewDictionaryMatchFilter(accepted []string) *DictionaryMatchFilter {
	m := make(map[string]struct{}, len(accepted))
	for _, w := range accepted {
		m[w] = struct{}{}
	}
	return &DictionaryMatchFilter{AcceptedWords: m}
}

// Filter ports filter(List<RuleMatch>, AnnotatedText) — text is markup text
// (equals plain original when no markup).
func (f *DictionaryMatchFilter) Filter(ruleMatches []*RuleMatch, textWithMarkup string) []*RuleMatch {
	var out []*RuleMatch
	for _, match := range ruleMatches {
		// Java uses String.substring (UTF-16 indices)
		covered := utf16Substring(textWithMarkup, match.FromPos, match.ToPos)
		if _, ok := f.AcceptedWords[covered]; ok {
			continue
		}
		out = append(out, match)
	}
	return out
}
