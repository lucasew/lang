package rules

// NewLineMatchFilter ports org.languagetool.rules.NewLineMatchFilter.
// Java: implements RuleMatchFilter; trims leading/trailing "\n" and U+2063 from
// match spans and suggestions; drops matches that become no-ops after trim.
type NewLineMatchFilter struct{}

// INVISIBLE_SEPARATOR ports NewLineMatchFilter.INVISIBLE_SEPARATOR ("\u2063").
const invisibleSeparatorNL = "\u2063"

func NewNewLineMatchFilter() *NewLineMatchFilter { return &NewLineMatchFilter{} }

// Filter ports filter(List<RuleMatch>, AnnotatedText).
// originalText is AnnotatedText.getOriginalText(); positions are Java UTF-16 indices.
func (f *NewLineMatchFilter) Filter(ruleMatches []*RuleMatch, originalText string) []*RuleMatch {
	// Java: ruleMatches.stream().filter(...).collect(Collectors.toCollection(ArrayList::new))
	out := make([]*RuleMatch, 0, len(ruleMatches))
	textLen := utf16Len(originalText)
	for _, ruleMatch := range ruleMatches {
		from := ruleMatch.GetFromPos()
		to := ruleMatch.GetToPos()

		// without some old tests failing with ArrayOutOfBoundExceptions. Should be safe to just keep the matches.
		// Java: if (text.getOriginalText().length() < from || text.getOriginalText().length() < to) return true;
		if textLen < from || textLen < to {
			out = append(out, ruleMatch)
			continue
		}

		// Java: var matchText = text.getOriginalText().substring(from, to);
		matchText := utf16Substring(originalText, from, to)

		// remove trailing \n and \u2063 from the matchText and update to
		// Java: while (matchText.endsWith("\n") || matchText.endsWith(INVISIBLE_SEPARATOR))
		drop := false
		for endsWithNLOrInvis(matchText) {
			// matchText = matchText.substring(0, matchText.length() - 1); to--;
			matchText = utf16Substring(matchText, 0, utf16Len(matchText)-1)
			to--
			if to < from {
				// only new line match can be removed
				drop = true
				break
			}
		}
		if drop {
			continue
		}
		// Java: while (matchText.startsWith("\n") || matchText.startsWith(INVISIBLE_SEPARATOR))
		for startsWithNLOrInvis(matchText) {
			// matchText = matchText.substring(1); from++;
			matchText = utf16Substring(matchText, 1, utf16Len(matchText))
			from++
		}

		// Java: var newSuggestionReplacements = new ArrayList<String>();
		// ruleMatch.getSuggestedReplacements().forEach(replacement -> { ... });
		newSuggestions := make([]string, 0, len(ruleMatch.GetSuggestedReplacements()))
		for _, replacement := range ruleMatch.GetSuggestedReplacements() {
			newReplacement := replacement
			for endsWithNLOrInvis(newReplacement) {
				newReplacement = utf16Substring(newReplacement, 0, utf16Len(newReplacement)-1)
			}
			for startsWithNLOrInvis(newReplacement) {
				newReplacement = utf16Substring(newReplacement, 1, utf16Len(newReplacement))
			}
			newSuggestions = append(newSuggestions, newReplacement)
		}

		// Java: if (newSuggestionReplacements.size() == 1 && newSuggestionReplacements.get(0).equals(matchText))
		if len(newSuggestions) == 1 && newSuggestions[0] == matchText {
			// return false — drop
			continue
		}
		// else: setOffsetPosition(from,to); setSuggestedReplacements(...); return true
		ruleMatch.SetOffsetPosition(from, to)
		ruleMatch.SetSuggestedReplacements(newSuggestions)
		out = append(out, ruleMatch)
	}
	return out
}

// endsWithNLOrInvis ports Java endsWith("\n") || endsWith(INVISIBLE_SEPARATOR) on UTF-16 units.
// Empty string → false (same as Java String.endsWith).
func endsWithNLOrInvis(s string) bool {
	n := utf16Len(s)
	if n == 0 {
		return false
	}
	// Both "\n" and U+2063 are single BMP code units.
	last := utf16Substring(s, n-1, n)
	return last == "\n" || last == invisibleSeparatorNL
}

// startsWithNLOrInvis ports Java startsWith("\n") || startsWith(INVISIBLE_SEPARATOR) on UTF-16 units.
// Empty string → false (same as Java String.startsWith).
func startsWithNLOrInvis(s string) bool {
	n := utf16Len(s)
	if n == 0 {
		return false
	}
	first := utf16Substring(s, 0, 1)
	return first == "\n" || first == invisibleSeparatorNL
}
