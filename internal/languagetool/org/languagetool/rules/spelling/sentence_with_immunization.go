package spelling

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// SentenceWithImmunization ports Rule.getSentenceWithImmunization:
// copy the sentence and run getAntiPatterns() DisambiguationPatternRule.replace
// (IMMUNIZE / IGNORE_SPELLING) so match() sees immunized or ignore-spelling tokens.
func (r *SpellingCheckRule) SentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if r == nil || sentence == nil {
		return sentence
	}
	aps := r.GetAntiPatterns()
	if len(aps) == 0 {
		return sentence
	}
	immunized := sentence.Copy(sentence)
	for _, ap := range aps {
		if ap == nil {
			continue
		}
		if next := ap.Replace(immunized); next != nil {
			immunized = next
		}
	}
	return immunized
}

// ApplyAntiPatterns is an alias used by spellers that hold DisambiguationPatternRule lists.
func ApplyAntiPatterns(sentence *languagetool.AnalyzedSentence, aps []*disambigrules.DisambiguationPatternRule) *languagetool.AnalyzedSentence {
	if sentence == nil || len(aps) == 0 {
		return sentence
	}
	immunized := sentence.Copy(sentence)
	for _, ap := range aps {
		if ap == nil {
			continue
		}
		if next := ap.Replace(immunized); next != nil {
			immunized = next
		}
	}
	return immunized
}
