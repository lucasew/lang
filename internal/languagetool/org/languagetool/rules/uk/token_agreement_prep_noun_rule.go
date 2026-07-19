package uk

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

// Java TokenAgreementPrepNounRule.getId()
const TokenAgreementPrepNounRuleID = "UK_PREP_NOUN_INFLECTION_AGREEMENT"

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

// HasNounOrPronObjectReading treats personal/possessive pronouns as objects for prep government.
func HasNounOrPronObjectReading(tok *languagetool.AnalyzedTokenReadings) bool {
	if HasNounReading(tok) {
		return true
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.Contains(p, "pron") && strings.Contains(p, "v_") {
			return true
		}
	}
	return false
}

func NewTokenAgreementPrepNounRule() *TokenAgreementPrepNounRule {
	return NewTokenAgreementPrepNounRuleWithMessages(nil)
}

// NewTokenAgreementPrepNounRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementPrepNounRuleWithMessages(messages map[string]string) *TokenAgreementPrepNounRule {
	cg := LoadCaseGovernmentHelper()
	r := &TokenAgreementPrepNounRule{CaseGov: cg}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID: TokenAgreementPrepNounRuleID,
		// Java getDescription / getShort
		description:  "Узгодження прийменника та іменника у реченні",
		shortMsg:     "Узгодження прийменника та іменника",
		isLeftToken:  hasPrepReading,
		isRightToken: HasNounOrPronObjectReading,
		pairChecker: func(left, right *languagetool.AnalyzedTokenReadings) bool {
			return prepNounAgree(cg, left, right)
		},
		exception: IsPrepNounException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

// HasVidmPosTag ports TokenAgreementPrepNounRule.hasVidmPosTag.
// posTagsToFind are case substrings like "v_oru"; if no vidminok found on any reading, returns true
// (Java incomplete dictionary path).
func HasVidmPosTag(posTagsToFind []string, tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return true
	}
	rds := tok.GetReadings()
	vidminokFound := false
	for _, token := range rds {
		if token == nil {
			continue
		}
		pos := token.GetPOSTag()
		if pos == nil {
			if len(rds) == 1 {
				return true
			}
			continue
		}
		// Java PosTagHelper.NO_VIDMINOK_SUBSTR
		if strings.Contains(*pos, ":nv") {
			return true
		}
		if strings.Contains(*pos, ":v_") {
			vidminokFound = true
			for _, want := range posTagsToFind {
				if want != "" && strings.Contains(*pos, want) {
					return true
				}
			}
		}
	}
	return !vidminokFound
}

func prepNounAgree(cg *CaseGovernmentHelper, prep, noun *languagetool.AnalyzedTokenReadings) bool {
	if cg == nil || prep == nil || noun == nil {
		return true
	}
	// lemma from prep token surface / lemma
	lemma := prep.GetToken()
	// strip soft hyphen / combining marks from surface lemma
	lemma = CleanIgnoreChars(lemma)
	for _, r := range prep.GetReadings() {
		if r != nil && r.GetLemma() != nil && *r.GetLemma() != "" {
			lemma = CleanIgnoreChars(*r.GetLemma())
			break
		}
	}
	govs := cg.GetCaseGovernments(lemma)
	if len(govs) == 0 {
		return true // unknown prep — no flag
	}
	nounInfs := GetNounCaseInflections(CollectPOSTags(noun))
	if len(nounInfs) == 0 {
		// try free case scan for pron tags
		for _, p := range CollectPOSTags(noun) {
			for _, c := range []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis", "v_kly"} {
				if strings.Contains(p, c) && cg.HasCaseGovernment(lemma, c) {
					return true
				}
			}
		}
		return true // insufficient
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
