package el

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// NumeralStressRule ports org.languagetool.rules.el.NumeralStressRule.
//
// Greek ordinal numerals written as arabic number + suffix may need stress
// (e.g. 20ός, 100ός but 1ος, 10ος). This rule checks incorrect usage or absence of stress.
type NumeralStressRule struct {
	rules.BaseRule
	Messages map[string]string
	// map from stressed to unstressed form and unstressed to stressed form
	suffixMap      map[string]string
	numeral        *regexp.Regexp
	stressedNumber *regexp.Regexp
	stressedSuffix *regexp.Regexp
}

// NewNumeralStressRule ports NumeralStressRule(ResourceBundle messages).
func NewNumeralStressRule(messages map[string]string) *NumeralStressRule {
	// Java unstressedSfx / stressedSfx arrays
	unstressedSfx := []string{
		"ος", "ου", "ο", "ον", "οι", "ων", "ους", "η", "ης", "ην", "ες", "α",
	}
	stressedSfx := []string{
		"ός", "ού", "ό", "όν", "οί", "ών", "ούς", "ή", "ής", "ήν", "ές", "ά",
	}
	suffixMap := make(map[string]string, len(stressedSfx)*2)
	stressedSuffixRE := ""
	for i, s := range stressedSfx {
		if i > 0 {
			stressedSuffixRE += "|"
		}
		stressedSuffixRE += s
		suffixMap[s] = unstressedSfx[i]
		suffixMap[unstressedSfx[i]] = s
	}
	// Java: "([1-9][0-9]*)(" + stressed + unstressed + ")"
	pattern := "([1-9][0-9]*)(" + stressedSuffixRE
	for _, sfx := range unstressedSfx {
		pattern += "|" + sfx
	}
	pattern += ")"

	r := &NumeralStressRule{
		Messages:  messages,
		suffixMap: suffixMap,
		// Java Matcher.matches() is full-string; RE2 MatchString is substring — anchor.
		numeral: regexp.MustCompile("^" + pattern + "$"),
		// Java: Pattern.compile("[0-9]*[0|2-9]0") with Matcher.matches().
		// Character class [0|2-9] is 0, literal '|', and 2-9; for digit tokens equivalent to [02-9].
		stressedNumber: regexp.MustCompile(`^[0-9]*[02-9]0$`),
		stressedSuffix: regexp.MustCompile("^(" + stressedSuffixRE + ")$"),
	}
	// Java: setCategory(new Category(new CategoryId("ORTHOGRAPHY"), "Orthography"));
	r.ID = "GREEK_ORTHOGRAPHY_NUMERAL_STRESS"
	r.Description = "Έλεγχος τονισμού αριθμητικών"
	r.SetCategory(rules.NewCategory(rules.NewCategoryId("ORTHOGRAPHY"), "Orthography"))
	// Java init(): setLocQualityIssueType(ITSIssueType.Misspelling); addExamplePair(...)
	r.SetLocQualityIssueType(rules.ITSMisspelling)
	r.AddExamplePair(
		rules.Wrong("Ο <marker>20ος</marker> αιώνας μαζί με τον 21ο αιώνα κατατάσσεται από τους ιστορικούς στη Σύγχρονη Ιστορία."),
		rules.Fixed("Ο <marker>20ός</marker> αιώνας μαζί με τον 21ο αιώνα κατατάσσεται από τους ιστορικούς στη Σύγχρονη Ιστορία."),
	)
	return r
}

// GetID ports NumeralStressRule.getId.
func (r *NumeralStressRule) GetID() string {
	if r == nil {
		return ""
	}
	return r.ID
}

// GetDescription ports NumeralStressRule.getDescription.
func (r *NumeralStressRule) GetDescription() string {
	if r == nil {
		return ""
	}
	return r.Description
}

// Match ports NumeralStressRule.match.
func (r *NumeralStressRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	var ruleMatches []*rules.RuleMatch
	// Java: sentence.getTokensWithoutWhitespace()
	for _, token := range sentence.GetTokensWithoutWhitespace() {
		// Java: numeral.matcher(token.getToken()).matches()
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
		// Java: suffix = suffixMap.get(suffix); suggestion = number + suffix
		// Map always contains regex-matched suffixes (bidirectional stressed/unstressed).
		suffix = r.suffixMap[suffix]
		suggestion := number + suffix
		msg := "<suggestion>" + suggestion + "</suggestion>"
		// Java: new RuleMatch(this, sentence, start, end, msg, "Πρόβλημα ορθογραφίας")
		// Java RuleMatch extracts <suggestion> tags into getSuggestedReplacements().
		rm := rules.NewRuleMatch(r, sentence, token.GetStartPos(), token.GetEndPos(), msg)
		rm.ShortMessage = "Πρόβλημα ορθογραφίας"
		rm.SetSuggestedReplacement(suggestion)
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}
