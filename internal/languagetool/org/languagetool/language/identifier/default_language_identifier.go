package identifier

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier/detector"
)

// DefaultLanguageIdentifier ports
// org.languagetool.language.identifier.DefaultLanguageIdentifier
// without optimaize profiles — uses pluggable scorers + unicode/common-word detectors.
type DefaultLanguageIdentifier struct {
	BaseLanguageIdentifier
	// ProfileScore is the main ngram/optimaize stand-in: cleanText → lang→score.
	ProfileScore func(cleanText string, preferred []string) map[string]float64
	// FastTextScore optional fastText stand-in.
	FastTextScore func(cleanText string) map[string]float64
	// NGram optional character n-gram detector.
	NGram *detector.CharNGramDetector
	// Unicode optional unicode script detector.
	Unicode *detector.UnicodeBasedDetector
	// CommonWords optional common-words detector.
	CommonWords *detector.CommonWordsDetector

	MinimalConfidence float64
	// IgnoreLangCodes are never returned as top (ast, gl by default).
	IgnoreLangCodes map[string]struct{}
}

func NewDefaultLanguageIdentifier(maxLength int) *DefaultLanguageIdentifier {
	if maxLength <= 0 {
		maxLength = DefaultMaxLength
	}
	d := &DefaultLanguageIdentifier{
		BaseLanguageIdentifier: NewBaseLanguageIdentifier(maxLength),
		MinimalConfidence:      0.0, // surface allows low scores; Java uses 0.9 inside optimaize
		IgnoreLangCodes: map[string]struct{}{
			"ast": {},
			"gl":  {},
		},
		Unicode: detector.NewUnicodeBasedDetector(),
	}
	return d
}

// EnableFastText installs a score function used preferentially for longer text.
func (d *DefaultLanguageIdentifier) EnableFastText(score func(string) map[string]float64) {
	d.FastTextScore = score
}

func (d *DefaultLanguageIdentifier) IsFastTextEnabled() bool {
	return d != nil && d.FastTextScore != nil
}

func (d *DefaultLanguageIdentifier) Detect(cleanText string, noopLangs, preferredLangs []string) *languagetool.DetectedLanguage {
	scores := d.Scores(cleanText, noopLangs, preferredLangs, false, 1)
	if len(scores) == 0 {
		return nil
	}
	return &scores[0]
}

func (d *DefaultLanguageIdentifier) Scores(cleanText string, noopLangs, preferredLangs []string, limitOnPreferred bool, count int) []languagetool.DetectedLanguage {
	if d == nil {
		return nil
	}
	text := strings.TrimSpace(cleanText)
	if text == "" {
		return nil
	}
	preferred := append([]string(nil), preferredLangs...)
	// Java DefaultLanguageIdentifier: text.length() <= CONSIDER_ONLY_PREFERRED_THRESHOLD
	// (UTF-16 units). Forces preferred-lang filter when short.
	if javaStringLen(text) <= ConsiderOnlyPreferredThreshold && len(preferred) > 0 {
		limitOnPreferred = true
	}

	scores := map[string]float64{}
	src := "profile"

	// Prefer fastText when available
	if d.FastTextScore != nil {
		for k, v := range d.FastTextScore(text) {
			scores[k] = v
		}
		src = "fasttext"
	}
	// Merge profile scores if empty or low confidence
	maxScore := maxMap(scores)
	if d.ProfileScore != nil && (len(scores) == 0 || maxScore < 0.85) {
		for k, v := range d.ProfileScore(text, preferred) {
			if cur, ok := scores[k]; !ok || v > cur {
				scores[k] = v
			}
		}
		if src == "fasttext" && maxScore < 0.85 {
			src = "profile"
		}
	}
	// Char n-gram detector
	if d.NGram != nil && len(scores) == 0 {
		for k, v := range d.NGram.DetectLanguages(text) {
			scores[k] = v
		}
		src = "ngram"
	}
	// Unicode fallback for non-latin
	if d.Unicode != nil && (len(scores) == 0 || hasNonLatin(text)) {
		for _, code := range d.Unicode.GetDominantLangCodes(text) {
			if _, ok := scores[code]; !ok {
				scores[code] = 0.6
			} else if scores[code] < 0.6 {
				scores[code] = 0.6
			}
		}
	}
	// Common words boost
	if d.CommonWords != nil {
		counts := d.CommonWords.GetKnownWordsPerLanguage(text)
		var total int
		for _, c := range counts {
			total += c
		}
		if total > 0 {
			for k, c := range counts {
				v := float64(c) / float64(total)
				scores[k] = scores[k]*0.5 + v*0.5
			}
		}
	}

	// Filter ignore / preferred / noop
	type pair struct {
		code  string
		score float64
	}
	var pairs []pair
	for k, v := range scores {
		if _, ign := d.IgnoreLangCodes[k]; ign {
			continue
		}
		if limitOnPreferred && len(preferred) > 0 && !containsStr(preferred, k) {
			continue
		}
		// strip country variants for pairing preferred en-US → en
		pairs = append(pairs, pair{k, v})
	}
	// sort desc
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].score > pairs[i].score {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	if count <= 0 {
		count = len(pairs)
	}
	var out []languagetool.DetectedLanguage
	for i := 0; i < len(pairs) && i < count; i++ {
		code := pairs[i].code
		// noop langs still reported but could be tagged — surface keeps code
		_ = noopLangs
		s := src
		out = append(out, languagetool.NewDetectedLanguageFull(
			"", code, float32(pairs[i].score), &s))
	}
	return out
}

// DetectLanguage is the Java-style entry that cleans text first.
func (d *DefaultLanguageIdentifier) DetectLanguage(text string, noop, preferred []string) *languagetool.DetectedLanguage {
	if d == nil {
		return nil
	}
	return d.Detect(d.CleanAndShortenText(text), noop, preferred)
}

func maxMap(m map[string]float64) float64 {
	var max float64
	for _, v := range m {
		if v > max {
			max = v
		}
	}
	return max
}

func hasNonLatin(text string) bool {
	for _, r := range text {
		if r > unicode.MaxASCII && !unicode.IsSpace(r) && !unicode.IsPunct(r) {
			// rough: non-latin letter
			if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Cyrillic, r) ||
				unicode.Is(unicode.Arabic, r) || unicode.Is(unicode.Hiragana, r) ||
				unicode.Is(unicode.Katakana, r) || unicode.Is(unicode.Greek, r) {
				return true
			}
		}
	}
	return false
}

var _ LanguageIdentifier = (*DefaultLanguageIdentifier)(nil)
