package spelling

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier"
)

// ForeignLanguageChecker ports org.languagetool.rules.spelling.ForeignLanguageChecker.
const (
	// ForeignErrorThreshold ports ERROR_THRESHOLD.
	ForeignErrorThreshold = 0.45
	// ForeignMinSentenceThreshold ports MIN_SENTENCE_THRESHOLD.
	ForeignMinSentenceThreshold = 3
	// ForeignMaxScoringLanguages ports MAX_SCORING_LANGUAGES.
	ForeignMaxScoringLanguages = 5
	// NoForeignLangDetected ports NO_FOREIGN_LANG_DETECTED.
	NoForeignLangDetected = "NO_FOREIGN_LANG_DETECTED"
)

// ForeignLanguageChecker ports org.languagetool.rules.spelling.ForeignLanguageChecker.
type ForeignLanguageChecker struct {
	languageShortCode  string
	sentence           string
	sentenceLength     int64
	preferredLanguages []string
}

// NewForeignLanguageChecker ports ForeignLanguageChecker(languageShortCode, sentence, sentenceLength, preferredLanguages).
func NewForeignLanguageChecker(languageShortCode, sentence string, sentenceLength int64, preferred []string) *ForeignLanguageChecker {
	pref := append([]string(nil), preferred...)
	return &ForeignLanguageChecker{
		languageShortCode:  languageShortCode,
		sentence:           sentence,
		sentenceLength:     sentenceLength,
		preferredLanguages: pref,
	}
}

// Check ports ForeignLanguageChecker.check(matchesSoFar).
// Returns language shortCode → confidence when the error ratio is high enough to
// suggest a different language; empty map when not triggered / identifier missing.
// Java returns LinkedHashMap; Go map does not preserve order (callers key-lookup).
func (c *ForeignLanguageChecker) Check(matchesSoFar int) map[string]float32 {
	if c == nil {
		return map[string]float32{}
	}
	// Java: float errorRatio = (float) matchesSoFar / sentenceLength;
	errorRatio := float32(matchesSoFar) / float32(c.sentenceLength)
	if c.sentenceLength >= ForeignMinSentenceThreshold && errorRatio >= ForeignErrorThreshold {
		// Java: LanguageIdentifier langIdent = LanguageIdentifierService.INSTANCE.getInitialized();
		langIdent := identifier.Instance.GetInitialized()
		if langIdent != nil {
			// Java: getDetectedLanguageScores(sentence, Collections.emptyList(), preferredLanguages, true, MAX_SCORING_LANGUAGES)
			detectedLanguageScores := langIdent.Scores(
				c.sentence,
				[]string{},
				c.preferredLanguages,
				true,
				ForeignMaxScoringLanguages,
			)
			// Java: LinkedHashMap — Go map is unordered; values/keys match.
			results := make(map[string]float32, ForeignMaxScoringLanguages)
			if len(detectedLanguageScores) > 0 {
				for i := 0; i < len(detectedLanguageScores); i++ {
					detectedLanguage := detectedLanguageScores[i]
					// Java: Language language = detectedLanguage.getDetectedLanguage();
					// language.getShortCode() — Go DetectedLanguage stores short code in DetectedLanguageCode.
					shortCode := detectedLanguage.GetDetectedLanguageCode()
					// The text main language still has the highest threshold
					if i == 0 && shortCode == c.languageShortCode {
						return map[string]float32{NoForeignLangDetected: 0.99}
					}
					// Java: log.trace(...) — omit (no behavior impact)
					results[shortCode] = detectedLanguage.GetDetectionConfidence()
				}
				return results
			}
			return map[string]float32{NoForeignLangDetected: 0.99}
		}
	}
	return map[string]float32{}
}
