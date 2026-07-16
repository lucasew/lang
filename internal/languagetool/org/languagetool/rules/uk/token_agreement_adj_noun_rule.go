package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const TokenAgreementAdjNounRuleID = "UK_ADJ_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementAdjNounRule ports org.languagetool.rules.uk.TokenAgreementAdjNounRule.
type TokenAgreementAdjNounRule struct {
	*tokenAgreementMatch
}

func NewTokenAgreementAdjNounRule() *TokenAgreementAdjNounRule {
	r := &TokenAgreementAdjNounRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID:      TokenAgreementAdjNounRuleID,
		description: "Узгодження відмінків, роду і числа прикметника та іменника",
		shortMsg:    "Узгодження прикметника та іменника",
		isLeftToken: HasAdjReading,
		isRightToken: HasNounReading,
		pairChecker: func(left, right *languagetool.AnalyzedTokenReadings) bool {
			return AdjNounAgree(CollectPOSTags(left), CollectPOSTags(right))
		},
		exception: IsAdjNounException,
	}
	return r
}

func (r *TokenAgreementAdjNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
