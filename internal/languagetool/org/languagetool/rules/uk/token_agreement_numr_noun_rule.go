package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

const TokenAgreementNumrNounRuleID = "UK_NUMR_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementNumrNounRule ports TokenAgreementNumrNounRule.
type TokenAgreementNumrNounRule struct {
	*tokenAgreementMatch
}

func hasNumrReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSNumr.Match(p) || taguk.IPOSNumber.Match(p) {
			return true
		}
	}
	return false
}

func NewTokenAgreementNumrNounRule() *TokenAgreementNumrNounRule {
	r := &TokenAgreementNumrNounRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID:       TokenAgreementNumrNounRuleID,
		description:  "Узгодження числівника та іменника",
		shortMsg:     "Узгодження числівника та іменника",
		isLeftToken:  hasNumrReading,
		isRightToken: HasNounReading,
		pairChecker: func(left, right *languagetool.AnalyzedTokenReadings) bool {
			return NumrNounAgree(CollectPOSTags(left), CollectPOSTags(right))
		},
		exception: IsNumrNounException,
	}
	return r
}

func (r *TokenAgreementNumrNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
