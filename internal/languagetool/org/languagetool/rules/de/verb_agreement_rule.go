package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// VerbAgreementRule is a surface stand-in for VerbAgreementRule focusing on
// personal-pronoun + wrong finite form of sein/haben.
type VerbAgreementRule struct {
	Messages map[string]string
}

func NewVerbAgreementRule(messages map[string]string) *VerbAgreementRule {
	return &VerbAgreementRule{Messages: messages}
}

func (r *VerbAgreementRule) GetID() string { return "DE_VERBAGREEMENT" }

// pronoun → wrong verb forms (sein)
var wrongSein = map[string]map[string]string{
	"ich": {"bist": "bin", "ist": "bin", "sind": "bin", "seid": "bin"},
	"du":  {"bin": "bist", "ist": "bist", "sind": "bist", "seid": "bist"},
	"er":  {"bin": "ist", "bist": "ist", "sind": "ist", "seid": "ist"},
	"sie": {"bin": "ist", "bist": "ist"}, // 3sg; 3pl uses sind — leave "sind" ok
	"es":  {"bin": "ist", "bist": "ist", "sind": "ist", "seid": "ist"},
	"wir": {"bin": "sind", "bist": "sind", "ist": "sind", "seid": "sind"},
	"ihr": {"bin": "seid", "bist": "seid", "ist": "seid", "sind": "seid"},
}

func (r *VerbAgreementRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens)-1; i++ {
		pro := strings.ToLower(tokens[i].GetToken())
		wrongs, ok := wrongSein[pro]
		if !ok {
			continue
		}
		// skip "Sie" polite / 3pl ambiguity when capitalized mid-sentence as address - still allow
		verb := strings.ToLower(tokens[i+1].GetToken())
		// skip "bin Laden"
		if verb == "laden" || (i+1 < len(tokens) && strings.EqualFold(tokens[i+1].GetToken(), "Laden")) {
			continue
		}
		if sug, bad := wrongs[verb]; bad {
			// "sie sind" is OK for 3pl — we didn't put sind for sie
			msg := "Möglicherweise falsche Verbform zu '" + tokens[i].GetToken() + "'."
			rm := rules.NewRuleMatch(r, sentence, tokens[i+1].GetStartPos(), tokens[i+1].GetEndPos(), msg)
			rm.ShortMessage = "Verb-Kongruenz"
			rm.SetSuggestedReplacement(sug)
			matches = append(matches, rm)
		}
	}
	return matches
}
