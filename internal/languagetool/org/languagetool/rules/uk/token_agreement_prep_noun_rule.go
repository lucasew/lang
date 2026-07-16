package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

const TokenAgreementPrepNounRuleID = "UK_PREP_NOUN_CASE_AGREEMENT"

// TokenAgreementPrepNounRule ports prep→noun case government check.
type TokenAgreementPrepNounRule struct {
	*tokenAgreementMatch
	CaseGov *CaseGovernmentHelper
}

func hasPrepReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSPrep.Match(p) {
			return true
		}
	}
	return false
}

func NewTokenAgreementPrepNounRule() *TokenAgreementPrepNounRule {
	cg := LoadCaseGovernmentHelper()
	r := &TokenAgreementPrepNounRule{CaseGov: cg}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID:       TokenAgreementPrepNounRuleID,
		description:  "Узгодження прийменника та іменника (керування відмінком)",
		shortMsg:     "Узгодження прийменника та іменника",
		isLeftToken:  hasPrepReading,
		isRightToken: HasNounReading,
		pairChecker: func(left, right *languagetool.AnalyzedTokenReadings) bool {
			return prepNounAgree(cg, left, right)
		},
		exception: IsPrepNounException,
	}
	return r
}

func prepNounAgree(cg *CaseGovernmentHelper, prep, noun *languagetool.AnalyzedTokenReadings) bool {
	if cg == nil || prep == nil || noun == nil {
		return true
	}
	// lemma from prep token surface / lemma
	lemma := prep.GetToken()
	for _, r := range prep.GetReadings() {
		if r != nil && r.GetLemma() != nil && *r.GetLemma() != "" {
			lemma = *r.GetLemma()
			break
		}
	}
	govs := cg.GetCaseGovernments(lemma)
	if len(govs) == 0 {
		return true // unknown prep — no flag
	}
	nounInfs := GetNounCaseInflections(CollectPOSTags(noun))
	if len(nounInfs) == 0 {
		return true
	}
	for _, inf := range nounInfs {
		if cg.HasCaseGovernment(lemma, inf.Case) {
			return true
		}
	}
	return false
}

func (r *TokenAgreementPrepNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
