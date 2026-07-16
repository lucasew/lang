package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const TokenAgreementVerbNounRuleID = "UK_VERB_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementVerbNounRule ports verb + object/noun agreement surface (simplified).
type TokenAgreementVerbNounRule struct {
	*tokenAgreementMatch
}

func NewTokenAgreementVerbNounRule() *TokenAgreementVerbNounRule {
	r := &TokenAgreementVerbNounRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID:       TokenAgreementVerbNounRuleID,
		description:  "Узгодження дієслова та іменника",
		shortMsg:     "Узгодження дієслова та іменника",
		isLeftToken:  hasVerbReading,
		isRightToken: HasNounReading,
		pairChecker:  verbNounAgree,
		exception:    IsVerbNounException,
	}
	return r
}

func verbNounAgree(verb, noun *languagetool.AnalyzedTokenReadings) bool {
	// Full government tables deferred; only flag when both have tags and case gov fails for known prep-like lemmas.
	// Default: no flag without case-government data on the verb lemma.
	_ = verb
	_ = noun
	return true
}

func (r *TokenAgreementVerbNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
