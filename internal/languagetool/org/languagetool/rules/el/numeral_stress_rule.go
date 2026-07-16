package el

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// NumeralStressRule ports org.languagetool.rules.el.NumeralStressRule.
type NumeralStressRule struct {
	Messages       map[string]string
	suffixMap      map[string]string
	numeral        *regexp.Regexp
	stressedNumber *regexp.Regexp
	stressedSuffix *regexp.Regexp
}

func NewNumeralStressRule(messages map[string]string) *NumeralStressRule {
	unstressed := []string{"ος", "ου", "ο", "ον", "οι", "ων", "ους", "η", "ης", "ην", "ες", "α"}
	stressed := []string{"ός", "ού", "ό", "όν", "οί", "ών", "ούς", "ή", "ής", "ήν", "ές", "ά"}
	suffixMap := map[string]string{}
	stressedRE := ""
	for i, s := range stressed {
		if i > 0 {
			stressedRE += "|"
		}
		stressedRE += s
		suffixMap[s] = unstressed[i]
		suffixMap[unstressed[i]] = s
	}
	pattern := "([1-9][0-9]*)(" + stressedRE
	for _, sfx := range unstressed {
		pattern += "|" + sfx
	}
	pattern += ")"
	return &NumeralStressRule{
		Messages:  messages,
		suffixMap: suffixMap,
		numeral:   regexp.MustCompile("^" + pattern + "$"),
		// Java Matcher.matches() is full-string; RE2 MatchString is substring — anchor.
		stressedNumber: regexp.MustCompile(`^[0-9]*[02-9]0$`),
		stressedSuffix: regexp.MustCompile("^(" + stressedRE + ")$"),
	}
}

func (r *NumeralStressRule) GetID() string { return "GREEK_ORTHOGRAPHY_NUMERAL_STRESS" }

func (r *NumeralStressRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var out []*rules.RuleMatch
	for _, token := range sentence.GetTokensWithoutWhitespace() {
		m := r.numeral.FindStringSubmatch(token.GetToken())
		if m == nil {
			continue
		}
		number, suffix := m[1], m[2]
		needsStress := r.stressedNumber.MatchString(number)
		hasStress := r.stressedSuffix.MatchString(suffix)
		if needsStress == hasStress {
			continue
		}
		alt := r.suffixMap[suffix]
		if alt == "" {
			continue
		}
		suggestion := number + alt
		msg := "<suggestion>" + suggestion + "</suggestion>"
		rm := rules.NewRuleMatch(r, sentence, token.GetStartPos(), token.GetEndPos(), msg)
		rm.ShortMessage = "Πρόβλημα ορθογραφίας"
		rm.SetSuggestedReplacement(suggestion)
		out = append(out, rm)
	}
	return out
}
