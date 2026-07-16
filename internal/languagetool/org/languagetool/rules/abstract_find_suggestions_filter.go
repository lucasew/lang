package rules

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

const findSuggestionsMax = 10

// AbstractFindSuggestionsFilter ports org.languagetool.rules.AbstractFindSuggestionsFilter
// with pluggable spelling suggestions and optional POS filter.
type AbstractFindSuggestionsFilter struct {
	// SpellingSuggestions returns candidates for the focused token.
	SpellingSuggestions func(atr *languagetool.AnalyzedTokenReadings) []string
	// MatchesDesiredPostag optional; when set, filters candidates by desiredPostag regex.
	MatchesDesiredPostag func(suggestion string, desiredPostag string) bool
}

// AcceptRuleMatch filters/ranks suggestions for wordFrom token.
// Required args: wordFrom, desiredPostag. Optional: removeSuggestionsRegexp, suppressMatch, Mode=diacritics.
func (f *AbstractFindSuggestionsFilter) AcceptRuleMatch(match *RuleMatch, args map[string]string,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *RuleMatch {
	if f == nil || f.SpellingSuggestions == nil || match == nil {
		return nil
	}
	wordFrom, ok := args["wordFrom"]
	if !ok {
		panic("Missing key 'wordFrom'")
	}
	desiredPostag, ok := args["desiredPostag"]
	if !ok {
		panic("Missing key 'desiredPostag'")
	}
	atr := resolveWordFrom(wordFrom, match, patternTokens, tokenPositions)
	if atr == nil {
		return nil
	}
	cands := f.SpellingSuggestions(atr)
	var removeRE *regexp.Regexp
	if re, ok := args["removeSuggestionsRegexp"]; ok && re != "" {
		removeRE, _ = regexp.Compile(re)
	}
	var out []string
	for _, c := range cands {
		if removeRE != nil && removeRE.MatchString(c) {
			continue
		}
		if f.MatchesDesiredPostag != nil && !f.MatchesDesiredPostag(c, desiredPostag) {
			continue
		}
		out = append(out, c)
		if len(out) >= findSuggestionsMax {
			break
		}
	}
	suppress := strings.EqualFold(args["suppressMatch"], "true")
	diacriticsMode := args["Mode"] == "diacritics"
	if diacriticsMode {
		out = filterDiacriticOnly(atr.GetToken(), out)
	}
	if len(out) == 0 && suppress {
		return nil
	}
	if len(out) == 0 && diacriticsMode {
		return nil
	}
	match.SetSuggestedReplacements(out)
	return match
}

func resolveWordFrom(wordFrom string, match *RuleMatch, patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *languagetool.AnalyzedTokenReadings {
	if wordFrom == "marker" || wordFrom == "inmarker" {
		for _, t := range patternTokens {
			if t.GetStartPos() >= match.GetFromPos() && t.GetStartPos() < match.GetToPos() {
				return t
			}
		}
		if len(patternTokens) > 0 {
			return patternTokens[0]
		}
		return nil
	}
	n, err := strconv.Atoi(wordFrom)
	if err != nil {
		return nil
	}
	idx := skipCorrectedRef(tokenPositions, n)
	if idx < 0 || idx >= len(patternTokens) {
		return nil
	}
	return patternTokens[idx]
}

func filterDiacriticOnly(original string, cands []string) []string {
	base := stripDiacriticsLight(original)
	var out []string
	for _, c := range cands {
		if stripDiacriticsLight(c) == base && c != original {
			out = append(out, c)
		}
	}
	return out
}

func stripDiacriticsLight(s string) string {
	// reuse simple map from multitoken-style replacements
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch r {
		case 'á', 'à', 'â', 'ä', 'ã':
			b.WriteByte('a')
		case 'é', 'è', 'ê', 'ë':
			b.WriteByte('e')
		case 'í', 'ì', 'î', 'ï':
			b.WriteByte('i')
		case 'ó', 'ò', 'ô', 'ö', 'õ':
			b.WriteByte('o')
		case 'ú', 'ù', 'û', 'ü':
			b.WriteByte('u')
		case 'ç':
			b.WriteByte('c')
		case 'ñ':
			b.WriteByte('n')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
