package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// RegularIrregularParticipleFilter ports surface logic from
// org.languagetool.rules.pt.RegularIrregularParticipleFilter.
// Synthesizer-produced participle candidates are supplied by the caller.
type RegularIrregularParticipleFilter struct{}

func NewRegularIrregularParticipleFilter() *RegularIrregularParticipleFilter {
	return &RegularIrregularParticipleFilter{}
}

// IsRegular reports regular PT participle endings (do/dos/da/das).
func IsRegularParticiple(p string) bool {
	lp := strings.ToLower(p)
	return strings.HasSuffix(lp, "do") || strings.HasSuffix(lp, "dos") ||
		strings.HasSuffix(lp, "da") || strings.HasSuffix(lp, "das")
}

// Suggest picks a regular/irregular alternate from synthesizer candidates.
// direction is "RegularToIrregular" or "IrregularToRegular".
// token is the matched participle form; candidates are synthesised forms.
// template may contain {suggestion}/{Suggestion}/{SUGGESTION}.
func (f *RegularIrregularParticipleFilter) Suggest(direction, token string, candidates []string, template string) string {
	if len(candidates) < 2 {
		return ""
	}
	var replacement string
	dir := strings.ToLower(direction)
	if dir == strings.ToLower("RegularToIrregular") && IsRegularParticiple(token) {
		if !IsRegularParticiple(candidates[0]) {
			replacement = candidates[0]
		} else if !IsRegularParticiple(candidates[1]) {
			replacement = candidates[1]
		}
	} else if dir == strings.ToLower("IrregularToRegular") && !IsRegularParticiple(token) {
		if IsRegularParticiple(candidates[0]) {
			replacement = candidates[0]
		} else if IsRegularParticiple(candidates[1]) {
			replacement = candidates[1]
		}
	}
	if replacement == "" {
		return ""
	}
	if template == "" {
		return replacement
	}
	s := strings.ReplaceAll(template, "{suggestion}", replacement)
	s = strings.ReplaceAll(s, "{Suggestion}", tools.UppercaseFirstChar(replacement))
	s = strings.ReplaceAll(s, "{SUGGESTION}", strings.ToUpper(replacement))
	return s
}
