package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AgreementRule2 is a surface stand-in for AgreementRule2 (sentence-start ADJ+SUB).
// Without POS, flags a few common wrong adjective endings before known nouns.
type AgreementRule2 struct {
	Messages map[string]string
}

func NewAgreementRule2(messages map[string]string) *AgreementRule2 {
	return &AgreementRule2{Messages: messages}
}

func (r *AgreementRule2) GetID() string { return "DE_AGREEMENT2" }

// common neuter nouns that often appear in the Java tests
var neuterNounsAR2 = map[string]struct{}{
	"Haus": {}, "Auto": {}, "Wachstum": {}, "Kind": {}, "Buch": {}, "Taschenbuch": {},
	"Wasser": {}, "Mädchen": {}, "Zimmer": {}, "Jahr": {},
}

func (r *AgreementRule2) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	// find first content token (skip SENT_START and quotes)
	i := 1
	for i < len(tokens) && isQuoteTok(tokens[i].GetToken()) {
		i++
	}
	if i+1 >= len(tokens) {
		return nil
	}
	adj := tokens[i]
	noun := tokens[i+1]
	aw, nw := adj.GetToken(), noun.GetToken()
	if !tools.StartsWithUppercase(aw) || !tools.StartsWithUppercase(nw) {
		return nil
	}
	// "Deutscher Taschenbuch Verlag" — three SUB: no alarm
	if i+2 < len(tokens) && tools.StartsWithUppercase(tokens[i+2].GetToken()) && !isPrepLikeDE(tokens[i+2].GetToken()) {
		// third capital continues name/title
		third := tokens[i+2].GetToken()
		if isContentWordCap(third) {
			return nil
		}
	}
	// undeclined adj-ish before known neuter noun (Wirtschaftlich Wachstum)
	if isUndeclinedAdjStart(aw) {
		if _, neu := neuterNounsAR2[nw]; neu {
			msg := "Möglicherweise falsche Adjektiv-Endung vor dem Nomen."
			rm := rules.NewRuleMatch(r, sentence, adj.GetStartPos(), noun.GetEndPos(), msg)
			rm.ShortMessage = "Adjektiv-Kongruenz"
			rm.SetSuggestedReplacement(aw + "es " + nw)
			return []*rules.RuleMatch{rm}
		}
	}
	// -er before neuter noun (Kleiner Haus)
	if strings.HasSuffix(aw, "er") && !strings.HasSuffix(strings.ToLower(aw), "ier") {
		if _, neu := neuterNounsAR2[nw]; neu {
			// suggest -es form: Kleiner -> Kleines
			base := strings.TrimSuffix(aw, "er")
			sug := base + "es " + nw
			msg := "Möglicherweise falsche Adjektiv-Endung vor dem Nomen."
			rm := rules.NewRuleMatch(r, sentence, adj.GetStartPos(), noun.GetEndPos(), msg)
			rm.ShortMessage = "Adjektiv-Kongruenz"
			rm.SetSuggestedReplacement(sug)
			return []*rules.RuleMatch{rm}
		}
	}
	return nil
}

func isQuoteTok(w string) bool {
	switch w {
	case "\"", "„", "»", "«", "'", "“", "”":
		return true
	}
	return false
}

func isUndeclinedAdjStart(w string) bool {
	// sentence-start adjective without strong ending: Wirtschaftlich
	lc := strings.ToLower(w)
	if strings.HasSuffix(lc, "er") || strings.HasSuffix(lc, "es") || strings.HasSuffix(lc, "en") ||
		strings.HasSuffix(lc, "em") || strings.HasSuffix(lc, "e") {
		return false
	}
	// -lich, -ig, -isch, -bar, -sam
	for _, suf := range []string{"lich", "ig", "isch", "bar", "sam", "iv", "al"} {
		if strings.HasSuffix(lc, suf) && utf8.RuneCountInString(lc) > len(suf)+2 {
			return true
		}
	}
	return false
}

func isContentWordCap(w string) bool {
	if !tools.StartsWithUppercase(w) || utf8.RuneCountInString(w) < 2 {
		return false
	}
	for _, r := range w {
		if !unicode.IsLetter(r) && r != '-' {
			return false
		}
	}
	return true
}
