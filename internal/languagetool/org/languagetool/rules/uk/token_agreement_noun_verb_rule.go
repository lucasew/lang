package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

const TokenAgreementNounVerbRuleID = "UK_NOUN_VERB_INFLECTION_AGREEMENT"

// TokenAgreementNounVerbRule ports subject-noun + verb person/number agreement (simplified).
type TokenAgreementNounVerbRule struct {
	*tokenAgreementMatch
}

func hasVerbReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSVerb.Match(p) {
			return true
		}
	}
	return false
}

func NewTokenAgreementNounVerbRule() *TokenAgreementNounVerbRule {
	r := &TokenAgreementNounVerbRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID:       TokenAgreementNounVerbRuleID,
		description:  "Узгодження іменника-підмета та дієслова",
		shortMsg:     "Узгодження іменника та дієслова",
		isLeftToken:  HasNounReading,
		isRightToken: hasVerbReading,
		pairChecker:  nounVerbAgree,
		exception:    IsNounVerbException,
	}
	return r
}

func nounVerbAgree(noun, verb *languagetool.AnalyzedTokenReadings) bool {
	nTags := CollectPOSTags(noun)
	vTags := CollectPOSTags(verb)
	if len(GetNounInflections(nTags)) == 0 || len(GetVerbInflections(vTags)) == 0 {
		return true
	}
	return VerbInflectionsOverlap(vTags, nTags)
}

func (r *TokenAgreementNounVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
