package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MissingCommaRelativeClauseRule is a surface stand-in for MissingCommaRelativeClauseRule.
// Behind=false: missing comma before a relative pronoun after a noun-like token.
// Behind=true: missing comma after a relative clause (heuristic).
type MissingCommaRelativeClauseRule struct {
	Messages map[string]string
	Behind   bool
}

func NewMissingCommaRelativeClauseRule(messages map[string]string) *MissingCommaRelativeClauseRule {
	return &MissingCommaRelativeClauseRule{Messages: messages}
}

func NewMissingCommaRelativeClauseRuleBehind(messages map[string]string) *MissingCommaRelativeClauseRule {
	return &MissingCommaRelativeClauseRule{Messages: messages, Behind: true}
}

func (r *MissingCommaRelativeClauseRule) GetID() string {
	if r.Behind {
		return "COMMA_BEHIND_RELATIVE_CLAUSE"
	}
	return "COMMA_RELATIVE_CLAUSE"
}

var relativePronouns = map[string]struct{}{
	"der": {}, "die": {}, "das": {}, "dem": {}, "den": {}, "des": {},
	"dessen": {}, "deren": {}, "denen": {},
	"welche": {}, "welcher": {}, "welches": {}, "welchem": {}, "welchen": {},
	"was": {}, "wessen": {},
}

func isRelativePronounDE(w string) bool {
	_, ok := relativePronouns[strings.ToLower(w)]
	return ok
}

func isNounLikeDE(w string) bool {
	return tools.StartsWithUppercase(w) && utf8.RuneCountInString(w) >= 3
}

func isPrepLikeDE(w string) bool {
	switch strings.ToLower(w) {
	case "in", "an", "auf", "mit", "von", "zu", "bei", "nach", "vor", "über", "unter", "aus", "für":
		return true
	}
	return false
}

func (r *MissingCommaRelativeClauseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r.Behind {
		return r.matchBehind(sentence)
	}
	return r.matchBefore(sentence)
}

func (r *MissingCommaRelativeClauseRule) matchBefore(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 2; i < len(tokens)-1; i++ {
		if !isRelativePronounDE(tokens[i].GetToken()) {
			continue
		}
		if tokens[i-1].GetToken() == "," {
			continue
		}
		fromIdx := -1
		// "Auto das" / "Alles was"
		if isNounLikeDE(tokens[i-1].GetToken()) {
			fromIdx = i
		}
		// "Auto in dem"
		if isPrepLikeDE(tokens[i-1].GetToken()) && i >= 2 && isNounLikeDE(tokens[i-2].GetToken()) && tokens[i-2].GetToken() != "," {
			fromIdx = i - 1
		}
		if fromIdx < 0 {
			continue
		}
		toIdx := i
		for j := i + 1; j < len(tokens) && j <= i+3; j++ {
			if tokens[j].GetToken() == "," {
				break
			}
			toIdx = j
		}
		msg := "Fehlendes Komma vor dem Relativsatz?"
		rm := rules.NewRuleMatch(r, sentence, tokens[fromIdx].GetStartPos(), tokens[toIdx].GetEndPos(), msg)
		rm.ShortMessage = "fehlendes Komma"
		return []*rules.RuleMatch{rm}
	}
	return nil
}

func (r *MissingCommaRelativeClauseRule) matchBehind(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens)-2; i++ {
		if tokens[i].GetToken() != "," {
			continue
		}
		// ", das" / ", die" / ", in dem"
		relIdx := -1
		if i+1 < len(tokens) && isRelativePronounDE(tokens[i+1].GetToken()) {
			relIdx = i + 1
		} else if i+2 < len(tokens) && isPrepLikeDE(tokens[i+1].GetToken()) && isRelativePronounDE(tokens[i+2].GetToken()) {
			relIdx = i + 2
		}
		if relIdx < 0 {
			continue
		}
		for j := relIdx + 1; j < len(tokens)-1; j++ {
			if tokens[j].GetToken() == "," {
				break
			}
			if !looksLikeFiniteVerbDE(tokens[j].GetToken()) {
				continue
			}
			// main clause continues without comma after relative-clause verb
			if j+1 < len(tokens) && tokens[j+1].GetToken() != "," && !isPunctDE(tokens[j+1].GetToken()) {
				to := j + 1
				msg := "Fehlendes Komma hinter dem Relativsatz?"
				rm := rules.NewRuleMatch(r, sentence, tokens[j].GetStartPos(), tokens[to].GetEndPos(), msg)
				rm.ShortMessage = "fehlendes Komma"
				return []*rules.RuleMatch{rm}
			}
		}
	}
	return nil
}

func looksLikeFiniteVerbDE(w string) bool {
	if w == "" || tools.StartsWithUppercase(w) {
		return false
	}
	lc := strings.ToLower(w)
	if utf8.RuneCountInString(lc) < 4 {
		return false
	}
	return strings.HasSuffix(lc, "t") || strings.HasSuffix(lc, "en") || strings.HasSuffix(lc, "te") ||
		strings.HasSuffix(lc, "ten") || strings.HasSuffix(lc, "st")
}

func isPunctDE(w string) bool {
	if w == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(w)
	return unicode.IsPunct(r)
}
