package de

import (
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AgreementRule is a partial surface stand-in for AgreementRule focusing on
// open compounds written as two capitalized tokens (e.g. "Original Mail").
// Full DET-ADJ-NOUN morphology needs the German tagger.
type AgreementRule struct {
	Messages map[string]string
}

func NewAgreementRule(messages map[string]string) *AgreementRule {
	return &AgreementRule{Messages: messages}
}

func (r *AgreementRule) GetID() string { return "DE_AGREEMENT" }

func (r *AgreementRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 2; i < len(tokens); i++ {
		a, b := tokens[i-1], tokens[i]
		wa, wb := a.GetToken(), b.GetToken()
		if !isOpenCompoundPart(wa) || !isOpenCompoundPart(wb) {
			continue
		}
		// skip sentence start pairs and titles after numbers
		if i-1 <= 1 {
			continue
		}
		// skip if either is a pure determiner
		if isDet(wa) || isDet(wb) {
			continue
		}
		// skip known multiword exceptions lightly
		if IsCaseRuleException(wa + " " + wb) {
			continue
		}
		msg := "Möglicherweise fehlender Bindestrich oder Zusammenschreibung (offenes Kompositum)."
		rm := rules.NewRuleMatch(r, sentence, a.GetStartPos(), b.GetEndPos(), msg)
		rm.ShortMessage = "Kompositum"
		rm.SetSuggestedReplacement(wa + "-" + wb)
		matches = append(matches, rm)
	}
	return matches
}

func isOpenCompoundPart(w string) bool {
	if utf8.RuneCountInString(w) < 3 || !tools.StartsWithUppercase(w) {
		return false
	}
	for _, r := range w {
		if !unicode.IsLetter(r) && r != '-' {
			return false
		}
	}
	// skip all-caps abbreviations
	allUpper := true
	for _, r := range w {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			allUpper = false
			break
		}
	}
	return !allUpper
}
