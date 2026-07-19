package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// FindSuggestionsEsFilter ports org.languagetool.rules.ca.FindSuggestionsEsFilter (1:1 Accept).
//
// SpellingSuggestions + Tag (or TagPOS) required; fail-closed without them.
type FindSuggestionsEsFilter struct {
	// SpellingSuggestions ports getSpellingSuggestions(atr).
	SpellingSuggestions func(atr *languagetool.AnalyzedTokenReadings) []string
	// Tag tags a cleaned candidate (Java getTagger().tag).
	Tag func(word string) *languagetool.AnalyzedTokenReadings
	// TagPOS is a legacy hook returning POS strings; used when Tag is nil.
	TagPOS func(form string) []string
	// MaxSuggestions caps raw replacement pairs (Java 2 * MAX_SUGGESTIONS = 20).
	MaxSuggestions int
}

func NewFindSuggestionsEsFilter() *FindSuggestionsEsFilter {
	return &FindSuggestionsEsFilter{MaxSuggestions: 10}
}

// Java patterns (Matcher.matches = full string).
var (
	pApostropheNeededES = regexp.MustCompile(`(?i)h?[aeiouàèéíòóú].*`)
	pPostagNominal      = regexp.MustCompile(`NP..[^0].*|NC.[SN].*|A...[SN].|V.P..S..|V.[NG].*|RG|PX..S...`)
	pPostagVerb3person  = regexp.MustCompile(`V...3.*`)
)

func esFullMatch(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

// RewriteEsSuggestions builds "és "+nominal / "es "+verb3 (legacy unit helper).
func (f *FindSuggestionsEsFilter) RewriteEsSuggestions(candidates []struct{ Form, POS string }, max int) []string {
	if max <= 0 {
		max = 20
	}
	var out []string
	seen := map[string]struct{}{}
	for _, c := range candidates {
		if len(out) >= max {
			break
		}
		if esFullMatch(pPostagNominal, c.POS) {
			s := "és " + c.Form
			if _, ok := seen[s]; !ok {
				seen[s] = struct{}{}
				out = append(out, s)
			}
		}
		if esFullMatch(pPostagVerb3person, c.POS) {
			// Java: only when NOT vowel-initial (no apostrophe needed)
			if !esFullMatch(pApostropheNeededES, c.Form) {
				s := "es " + strings.ToLower(c.Form)
				if _, ok := seen[s]; !ok {
					seen[s] = struct{}{}
					out = append(out, s)
				}
			}
		}
	}
	return out
}

// AcceptRuleMatch ports FindSuggestionsEsFilter.acceptRuleMatch.
func (f *FindSuggestionsEsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = arguments
	_ = patternTokenPos
	_ = tokenPositions
	if f == nil || match == nil {
		return nil
	}
	if f.SpellingSuggestions == nil && (f.Tag == nil && f.TagPOS == nil) {
		return nil
	}
	// Prefer sentence tokens when available (more faithful); fall back to patternTokens.
	var tokens []*languagetool.AnalyzedTokenReadings
	if match.Sentence != nil {
		tokens = match.Sentence.GetTokensWithoutWhitespace()
	}
	if len(tokens) == 0 {
		tokens = patternTokens
	}
	if len(tokens) == 0 {
		return nil
	}

	posWord := 0
	for posWord < len(tokens) && tokens[posWord] != nil &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	posWord++ // token after "es"
	if posWord < 1 || posWord >= len(tokens) || tokens[posWord] == nil || tokens[posWord-1] == nil {
		return nil
	}
	atrWord := tokens[posWord]

	var suggestions []string
	if f.SpellingSuggestions != nil {
		suggestions = f.SpellingSuggestions(atrWord)
	}
	maxRaw := 2 * f.MaxSuggestions
	if f.MaxSuggestions <= 0 {
		maxRaw = 20
	}

	var replacements []string
	usedEsAccent := false
	usedEs := false
	for _, suggestion := range suggestions {
		if len(replacements) >= maxRaw {
			break
		}
		clean := suggestion // cleanSuggestion identity in parent
		if f.Tag != nil {
			analyzed := f.Tag(clean)
			if analyzed == nil {
				continue
			}
			// may have multiple readings
			for _, reading := range analyzed.GetReadings() {
				if reading == nil || reading.GetPOSTag() == nil {
					continue
				}
				if len(replacements) >= maxRaw {
					break
				}
				pos := *reading.GetPOSTag()
				tok := analyzed.GetToken()
				if esFullMatch(pPostagNominal, pos) {
					replacements = append(replacements, "és "+tok)
					usedEsAccent = true
				}
				if esFullMatch(pPostagVerb3person, pos) {
					if !esFullMatch(pApostropheNeededES, tok) {
						replacements = append(replacements, "es "+strings.ToLower(tok))
						usedEs = true
					}
				}
			}
		} else if f.TagPOS != nil {
			for _, pos := range f.TagPOS(clean) {
				if len(replacements) >= maxRaw {
					break
				}
				if esFullMatch(pPostagNominal, pos) {
					replacements = append(replacements, "és "+clean)
					usedEsAccent = true
				}
				if esFullMatch(pPostagVerb3person, pos) {
					if !esFullMatch(pApostropheNeededES, clean) {
						replacements = append(replacements, "es "+strings.ToLower(clean))
						usedEs = true
					}
				}
			}
		}
	}
	if len(replacements) == 0 {
		return nil
	}

	// Capitalize if "es" token is capitalized
	esTok := tokens[posWord-1].GetToken()
	firstCh := ""
	if esTok != "" {
		r := []rune(esTok)
		firstCh = string(r[0])
	}
	var definitive []string
	if firstCh != "" && strings.ToUpper(firstCh) == firstCh {
		for _, r := range replacements {
			definitive = append(definitive, tools.UppercaseFirstChar(r))
		}
	} else {
		definitive = append(definitive, replacements...)
	}

	isFirstEsAccent := strings.EqualFold(esTok, "és")
	if isFirstEsAccent && usedEsAccent && !usedEs {
		// show just the spelling rule
		return nil
	}

	message := match.GetMessage()
	if usedEsAccent {
		message = message + " \"És\" (del verb 'ser') s'escriu amb accent."
	}
	if usedEs {
		message = message + " \"Es\" (pronom) acompanya un verb en tercera persona."
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), message)
	out.ShortMessage = match.GetShortMessage()
	out.SetSuggestedReplacements(definitive)
	return out
}
