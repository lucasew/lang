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
	return NewTokenAgreementAdjNounRuleWithMessages(nil)
}

// NewTokenAgreementAdjNounRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementAdjNounRuleWithMessages(messages map[string]string) *TokenAgreementAdjNounRule {
	r := &TokenAgreementAdjNounRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID:       TokenAgreementAdjNounRuleID,
		description:  "Узгодження відмінків, роду і числа прикметника та іменника",
		shortMsg:     "Узгодження прикметника та іменника",
		isLeftToken:  HasAdjReading,
		isRightToken: HasNounReading,
		pairChecker: func(left, right *languagetool.AnalyzedTokenReadings) bool {
			if IsPredicativeAdjException(left) || IsAdjpException(left) {
				return true
			}
			return AdjNounAgree(CollectPOSTags(left), CollectPOSTags(right))
		},
		exception: IsAdjNounException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

func (r *TokenAgreementAdjNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
