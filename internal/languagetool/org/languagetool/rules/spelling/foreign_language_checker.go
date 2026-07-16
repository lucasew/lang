package spelling

// ForeignLanguageChecker ports org.languagetool.rules.spelling.ForeignLanguageChecker
// with a pluggable language-id hook (no LanguageIdentifierService).
const (
	ForeignErrorThreshold       = 0.45
	ForeignMinSentenceThreshold = 3
	ForeignMaxScoringLanguages  = 5
	NoForeignLangDetected       = "NO_FOREIGN_LANG_DETECTED"
)

// DetectedLanguageScore is one scored detection result.
type DetectedLanguageScore struct {
	ShortCode  string
	Confidence float32
	Source     string
}

// DetectScoresFunc returns ranked language scores for a sentence.
type DetectScoresFunc func(sentence string, preferred []string, maxResults int) []DetectedLanguageScore

// ForeignLanguageChecker checks whether high error ratio suggests a different language.
type ForeignLanguageChecker struct {
	LanguageShortCode  string
	Sentence           string
	SentenceLength     int64
	PreferredLanguages []string
	// Detect is optional; when nil, Check returns empty (same as no identifier).
	Detect DetectScoresFunc
}

func NewForeignLanguageChecker(languageShortCode, sentence string, sentenceLength int64, preferred []string) *ForeignLanguageChecker {
	return &ForeignLanguageChecker{
		LanguageShortCode:  languageShortCode,
		Sentence:           sentence,
		SentenceLength:     sentenceLength,
		PreferredLanguages: append([]string(nil), preferred...),
	}
}

// Check returns language→confidence when error ratio is high; empty if not triggered.
func (c *ForeignLanguageChecker) Check(matchesSoFar int) map[string]float32 {
	if c.SentenceLength < ForeignMinSentenceThreshold {
		return map[string]float32{}
	}
	errorRatio := float32(matchesSoFar) / float32(c.SentenceLength)
	if errorRatio < ForeignErrorThreshold {
		return map[string]float32{}
	}
	if c.Detect == nil {
		return map[string]float32{}
	}
	scores := c.Detect(c.Sentence, c.PreferredLanguages, ForeignMaxScoringLanguages)
	if len(scores) == 0 {
		return map[string]float32{NoForeignLangDetected: 0.99}
	}
	if scores[0].ShortCode == c.LanguageShortCode {
		return map[string]float32{NoForeignLangDetected: 0.99}
	}
	out := make(map[string]float32, len(scores))
	for _, s := range scores {
		out[s.ShortCode] = s.Confidence
	}
	return out
}
