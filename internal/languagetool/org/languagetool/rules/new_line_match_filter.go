package rules

// NewLineMatchFilter ports org.languagetool.rules.NewLineMatchFilter.
// Trims leading/trailing newlines and U+2063 from match spans and suggestions;
// drops matches that become no-ops after trim.
type NewLineMatchFilter struct{}

func NewNewLineMatchFilter() *NewLineMatchFilter { return &NewLineMatchFilter{} }

// Filter ports filter(List<RuleMatch>, AnnotatedText) — originalText is plain text.
func (f *NewLineMatchFilter) Filter(ruleMatches []*RuleMatch, originalText string) []*RuleMatch {
	textLen := utf16Len(originalText)
	var out []*RuleMatch
	for _, ruleMatch := range ruleMatches {
		from := ruleMatch.FromPos
		to := ruleMatch.ToPos
		if textLen < from || textLen < to {
			out = append(out, ruleMatch)
			continue
		}
		matchText := utf16Substring(originalText, from, to)
		drop := false
		for endsWithNLOrInvis(matchText) {
			matchText = trimLastUTF16Char(matchText)
			to--
			if to < from {
				drop = true
				break
			}
		}
		if drop {
			continue
		}
		for startsWithNLOrInvis(matchText) {
			matchText = trimFirstUTF16Char(matchText)
			from++
		}
		var newSuggestions []string
		for _, replacement := range ruleMatch.SuggestedReplacements {
			nr := replacement
			for endsWithNLOrInvis(nr) {
				nr = trimLastUTF16Char(nr)
			}
			for startsWithNLOrInvis(nr) {
				nr = trimFirstUTF16Char(nr)
			}
			newSuggestions = append(newSuggestions, nr)
		}
		if len(newSuggestions) == 1 && newSuggestions[0] == matchText {
			continue
		}
		ruleMatch.SetOffsetPosition(from, to)
		ruleMatch.SetSuggestedReplacements(newSuggestions)
		out = append(out, ruleMatch)
	}
	return out
}

func endsWithNLOrInvis(s string) bool {
	if s == "" {
		return false
	}
	rs := []rune(s)
	last := rs[len(rs)-1]
	return last == '\n' || last == '\u2063'
}

func startsWithNLOrInvis(s string) bool {
	if s == "" {
		return false
	}
	rs := []rune(s)
	return rs[0] == '\n' || rs[0] == '\u2063'
}

func trimLastUTF16Char(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return s
	}
	return string(rs[:len(rs)-1])
}

func trimFirstUTF16Char(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return s
	}
	return string(rs[1:])
}
