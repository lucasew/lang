package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// SentenceReplacer ports DisambiguationPatternRule.replace surface used by
// Rule.getSentenceWithImmunization (IMMUNIZE / IGNORE_SPELLING anti-patterns).
// Defined here as an interface to avoid import cycle with tagging/disambiguation/rules.
type SentenceReplacer interface {
	Replace(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// SentenceWithImmunization ports Rule.getSentenceWithImmunization:
// when antiPatterns is non-empty, copy the sentence and run each Replace.
// Empty antiPatterns returns the original sentence (Java identity).
func SentenceWithImmunization(sentence *languagetool.AnalyzedSentence, antiPatterns []SentenceReplacer) *languagetool.AnalyzedSentence {
	if sentence == nil || len(antiPatterns) == 0 {
		return sentence
	}
	immunized := sentence.Copy(sentence)
	for _, ap := range antiPatterns {
		if ap == nil {
			continue
		}
		if next := ap.Replace(immunized); next != nil {
			immunized = next
		}
	}
	return immunized
}
