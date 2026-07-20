package en

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

// EnglishConfusionProbabilityRule ports
// org.languagetool.rules.en.EnglishConfusionProbabilityRule.
type EnglishConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

const EnglishConfusionRuleID = "EN_CONFUSION_RULE"

// Java EnglishConfusionProbabilityRule.CONTRACTION
// Note: Java String indices are UTF-16; Go GoogleToken uses UTF-8 byte offsets
// (same as GetGoogleTokens). For ASCII + typographic apostrophe windows this still
// matches the intended "'t …" / "’t …" contraction skip.
var enConfusionContraction = regexp.MustCompile(`['’` + "`" + `´‘]t .*`)

func NewEnglishConfusionProbabilityRule(lm ngrams.LanguageModel) *EnglishConfusionProbabilityRule {
	return NewEnglishConfusionProbabilityRuleGrams(lm, 3)
}

// NewEnglishConfusionProbabilityRuleGrams ports the grams constructor.
func NewEnglishConfusionProbabilityRuleGrams(lm ngrams.LanguageModel, grams int) *EnglishConfusionProbabilityRule {
	base := ngrams.NewConfusionProbabilityRule(lm, grams)
	base.RuleIDOverride = EnglishConfusionRuleID
	// Java: super(..., EXCEPTIONS, ANTI_PATTERNS)
	base.Exceptions = append([]string(nil), EnglishConfusionExceptions...)
	base.IsException = enConfusionIsException
	base.IsCoveredByAntiPattern = enConfusionIsCoveredByAntiPattern
	// Java: breaks → brakes
	base.AddExamplePair(
		rules.Wrong("Don't forget to put on the <marker>breaks</marker>."),
		rules.Fixed("Don't forget to put on the <marker>brakes</marker>."),
	)
	return &EnglishConfusionProbabilityRule{ConfusionProbabilityRule: base}
}

// enConfusionIsException ports EnglishConfusionProbabilityRule.isException (CONTRACTION).
func enConfusionIsException(sentence string, startPos, endPos int) bool {
	// Java: if (startPos > 3) covered = sentence.substring(startPos-3, endPos)
	if startPos <= 3 || endPos < startPos {
		return false
	}
	from := startPos - 3
	if from < 0 {
		from = 0
	}
	if endPos > len(sentence) {
		endPos = len(sentence)
	}
	if from >= endPos {
		return false
	}
	covered := sentence[from:endPos]
	return enConfusionContraction.MatchString(covered)
}

// Match ports ConfusionProbabilityRule.match through the English wrapper.
func (r *EnglishConfusionProbabilityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.Match(sentence)
}

// GetID ports getId (EN_CONFUSION_RULE).
func (r *EnglishConfusionProbabilityRule) GetID() string {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return EnglishConfusionRuleID
	}
	return r.ConfusionProbabilityRule.GetID()
}
