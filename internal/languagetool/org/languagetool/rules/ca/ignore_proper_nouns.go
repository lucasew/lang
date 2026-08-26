package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// IgnoreProperNouns ports org.languagetool.rules.ca.IgnoreProperNouns.
// Marks untagged tokens that already appeared as proper nouns (POS NP*) as known.
type IgnoreProperNouns struct {
	rules.TextLevelRuleBase
}

func NewIgnoreProperNouns() *IgnoreProperNouns {
	r := &IgnoreProperNouns{}
	r.ID = "IGNORE_PROPER_NOUNS"
	r.Description = "Ignora noms propis que hagin aparegut en altres parts del text."
	r.SetMinToCheckParagraph(0)
	return r
}

func (r *IgnoreProperNouns) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var out []*rules.RuleMatch
	pos := 0
	var seen []string
	for _, sentence := range sentences {
		if sentence == nil {
			continue
		}
		for _, token := range sentence.GetTokens() {
			if token.HasPosTagStartingWith("NP") {
				seen = append(seen, token.GetToken())
			}
			if !token.IsTagged() && containsStr(seen, token.GetToken()) {
				m := rules.NewRuleMatch(r, sentence, pos+token.GetStartPos(), pos+token.GetEndPos(),
					"Aquesta paraula ja aparegut abans i es pot donar per correcta.")
				out = append(out, m)
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return out
}

func containsStr(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
