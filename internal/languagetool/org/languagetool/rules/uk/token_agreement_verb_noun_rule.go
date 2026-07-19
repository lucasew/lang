package uk

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const TokenAgreementVerbNounRuleID = "UK_VERB_NOUN_INFLECTION_AGREEMENT"

// TokenAgreementVerbNounRule ports verb + object/noun agreement surface (simplified).
type TokenAgreementVerbNounRule struct {
	*tokenAgreementMatch
	// CaseGov optional inject; nil → LoadCaseGovernmentHelper().
	CaseGov *CaseGovernmentHelper
}

func NewTokenAgreementVerbNounRule() *TokenAgreementVerbNounRule {
	return NewTokenAgreementVerbNounRuleWithMessages(nil)
}

// NewTokenAgreementVerbNounRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementVerbNounRuleWithMessages(messages map[string]string) *TokenAgreementVerbNounRule {
	r := &TokenAgreementVerbNounRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID: TokenAgreementVerbNounRuleID,
		// Java getDescription / getShort
		description:  "Узгодження дієслова з іменником",
		shortMsg:     "Узгодження дієслова з іменником",
		isLeftToken:  hasVerbReading,
		isRightToken: HasNounReading,
		pairChecker:  r.verbNounAgree,
		exception:    IsVerbNounException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

func (r *TokenAgreementVerbNounRule) caseGov() *CaseGovernmentHelper {
	if r != nil && r.CaseGov != nil {
		return r.CaseGov
	}
	return LoadCaseGovernmentHelper()
}

func (r *TokenAgreementVerbNounRule) verbNounAgree(verb, noun *languagetool.AnalyzedTokenReadings) bool {
	return VerbNounCaseAgree(r.caseGov(), verb, noun)
}

// VerbNounCaseAgree returns false when a known verb government conflicts with all noun cases.
func VerbNounCaseAgree(cg *CaseGovernmentHelper, verb, noun *languagetool.AnalyzedTokenReadings) bool {
	if verb == nil || noun == nil || cg == nil {
		return true
	}
	// gather governed cases from verb lemmas that appear in the map
	var gov []string
	hasGovLemma := false
	for _, r := range verb.GetReadings() {
		if r == nil || r.GetPOSTag() == nil || !strings.HasPrefix(*r.GetPOSTag(), "verb") {
			continue
		}
		if r.GetLemma() == nil || *r.GetLemma() == "" {
			continue
		}
		cases := cg.GetCaseGovernments(*r.GetLemma())
		if len(cases) == 0 {
			continue
		}
		hasGovLemma = true
		gov = append(gov, cases...)
	}
	if !hasGovLemma {
		// no government data — do not flag
		return true
	}
	// noun cases present
	nounCases := map[string]struct{}{}
	for _, n := range noun.GetReadings() {
		if n == nil || n.GetPOSTag() == nil {
			continue
		}
		pos := *n.GetPOSTag()
		if !strings.HasPrefix(pos, "noun") && !strings.HasPrefix(pos, "adj") {
			continue
		}
		for _, c := range []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis", "v_kly", "v_inf"} {
			if strings.Contains(pos, c) {
				nounCases[c] = struct{}{}
			}
		}
	}
	if len(nounCases) == 0 {
		return true
	}
	// agree if any non-infinitive governed case is among noun cases
	hasNonInf := false
	for _, g := range gov {
		if g == "v_inf" {
			continue // infinitive complement — not noun-case checked here
		}
		hasNonInf = true
		if _, ok := nounCases[g]; ok {
			return true
		}
	}
	if !hasNonInf {
		// only v_inf government — do not flag noun objects
		return true
	}
	return false
}

func (r *TokenAgreementVerbNounRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
