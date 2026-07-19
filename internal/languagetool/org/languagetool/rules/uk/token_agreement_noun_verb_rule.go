package uk

import (
	"strings"

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
	return NewTokenAgreementNounVerbRuleWithMessages(nil)
}

// NewTokenAgreementNounVerbRuleWithMessages ports the Java ctor (ResourceBundle messages).
func NewTokenAgreementNounVerbRuleWithMessages(messages map[string]string) *TokenAgreementNounVerbRule {
	r := &TokenAgreementNounVerbRule{}
	r.tokenAgreementMatch = &tokenAgreementMatch{
		ruleID:       TokenAgreementNounVerbRuleID,
		// Java getDescription / getShort
		description:  "Узгодження іменника та дієслова за родом, числом та особою",
		shortMsg:     "Узгодження іменника з дієсловом",
		isLeftToken:  HasNounOrPronSubjectReading,
		isRightToken: hasVerbReading,
		pairChecker:  nounVerbAgree,
		exception:    IsNounVerbException,
	}
	initTokenAgreementMeta(r.tokenAgreementMatch, messages)
	return r
}

func nounVerbAgree(noun, verb *languagetool.AnalyzedTokenReadings) bool {
	nTags := CollectPOSTags(noun)
	vTags := CollectPOSTags(verb)
	// proper names soft: prop without clear person/number often skip
	if isProperNameOnly(nTags) && len(GetNounInflections(nTags)) == 0 {
		return true
	}
	// personal pronouns: use person/number soft matrix
	if hasPronPers(nTags) {
		return pronVerbAgree(nTags, vTags)
	}
	if len(GetNounInflections(nTags)) == 0 || len(GetVerbInflections(vTags)) == 0 {
		return true // insufficient data
	}
	return VerbInflectionsOverlap(vTags, nTags)
}

func hasPronPers(tags []string) bool {
	for _, t := range tags {
		if strings.Contains(t, "pron:pers") {
			return true
		}
	}
	return false
}

func isProperNameOnly(tags []string) bool {
	if len(tags) == 0 {
		return false
	}
	for _, t := range tags {
		if !strings.Contains(t, "prop") {
			return false
		}
	}
	return true
}

// pronVerbAgree soft-matches personal pronouns to verb person/number.
func pronVerbAgree(nTags, vTags []string) bool {
	// extract :1/:2/:3 and s/p from both
	var nPers, nNum, vPers, vNum string
	for _, t := range nTags {
		if !strings.Contains(t, "pron:pers") {
			continue
		}
		for _, p := range []string{":1", ":2", ":3"} {
			if strings.Contains(t, p) {
				nPers = p
			}
		}
		if strings.Contains(t, ":p:") || strings.HasSuffix(t, ":p") || strings.Contains(t, ":p:v_") {
			nNum = "p"
		} else if strings.Contains(t, ":s:") || strings.Contains(t, ":m:") || strings.Contains(t, ":f:") || strings.Contains(t, ":n:") {
			nNum = "s"
		}
		// Ukrainian: noun:…:p:v_naz:pron:pers:1
		if strings.Contains(t, ":p:") {
			nNum = "p"
		}
	}
	for _, t := range vTags {
		if !strings.HasPrefix(t, "verb") {
			continue
		}
		for _, p := range []string{":1", ":2", ":3"} {
			if strings.Contains(t, p) {
				vPers = p
			}
		}
		if strings.Contains(t, ":p:") || strings.Contains(t, ":p:3") || strings.Contains(t, "past:p") {
			vNum = "p"
		} else if strings.Contains(t, ":s:") || strings.Contains(t, "past:m") || strings.Contains(t, "past:f") || strings.Contains(t, "past:n") {
			vNum = "s"
		}
	}
	if nPers == "" || vPers == "" {
		return true // insufficient
	}
	if nPers != vPers {
		return false
	}
	if nNum != "" && vNum != "" && nNum != vNum {
		return false
	}
	return true
}

func (r *TokenAgreementNounVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.tokenAgreementMatch.Match(sentence)
}
