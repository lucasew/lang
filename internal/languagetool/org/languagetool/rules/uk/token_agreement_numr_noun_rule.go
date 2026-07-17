package uk

import (
	"strings"

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

// ForceNounLemmas soft-skip special measure nouns (тон/…) with numr.
var ForceNounLemmas = map[string]struct{}{
	"тон": {}, "тони": {},
}

// FractionalNumrLemmas soft fractional heads.
var FractionalNumrLemmas = map[string]struct{}{
	"півтора": {}, "півтори": {}, "пів": {},
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
			if IsForceNounException(left, right) {
				return true
			}
			if IsFractionalNumrException(left, right) {
				return true
			}
			return NumrNounAgree(CollectPOSTags(left), CollectPOSTags(right))
		},
		exception: IsNumrNounException,
	}
	return r
}

// IsForceNounException soft-skips agreement for known force-noun lemmas.
func IsForceNounException(numr, noun *languagetool.AnalyzedTokenReadings) bool {
	if noun == nil {
		return false
	}
	// surface or lemma in force list
	if _, ok := ForceNounLemmas[strings.ToLower(noun.GetToken())]; ok {
		return true
	}
	for _, r := range noun.GetReadings() {
		if r != nil && r.GetLemma() != nil {
			if _, ok := ForceNounLemmas[strings.ToLower(*r.GetLemma())]; ok {
				return true
			}
		}
	}
	return false
}

// IsFractionalNumrException soft-skips fractional numeral + noun pairs.
func IsFractionalNumrException(numr, noun *languagetool.AnalyzedTokenReadings) bool {
	if numr == nil {
		return false
	}
	if _, ok := FractionalNumrLemmas[strings.ToLower(numr.GetToken())]; ok {
		return true
	}
	for _, r := range numr.GetReadings() {
		if r != nil && r.GetLemma() != nil {
			if _, ok := FractionalNumrLemmas[strings.ToLower(*r.GetLemma())]; ok {
				return true
			}
		}
	}
	return false
}

func (r *TokenAgreementNumrNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
