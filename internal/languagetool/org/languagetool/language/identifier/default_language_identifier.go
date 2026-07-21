package identifier

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier/detector"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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
	// NGram ports Java NGramDetector (ZIP model or char-ngram test fallback).
	NGram *detector.NGramDetector
	// Unicode optional unicode script detector (Java UNICODE_BASED_LANG_IDENTIFIER).
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
	return d.DetectLimit(cleanText, noopLangs, preferredLangs, false)
}

// DetectLimit ports detectLanguage(..., limitOnPreferredLangs).
func (d *DefaultLanguageIdentifier) DetectLimit(cleanText string, noopLangs, preferredLangs []string, limitOnPreferred bool) *languagetool.DetectedLanguage {
	scores := d.Scores(cleanText, noopLangs, preferredLangs, limitOnPreferred, 1)
	if len(scores) == 0 {
		return nil
	}
	return &scores[0]
}

// EnableFastTextFromPaths ports enableFasttext(binary, model).
// Both paths required; creates FastTextDetector and wires FastTextScore.
// Nil/empty paths leave fasttext disabled (Java logs warn when either null).
func (d *DefaultLanguageIdentifier) EnableFastTextFromPaths(binaryPath, modelPath string) error {
	if d == nil {
		return nil
	}
	if binaryPath == "" || modelPath == "" {
		return nil
	}
	ft, err := detector.NewFastTextDetector(modelPath, binaryPath)
	if err != nil {
		return err
	}
	d.EnableFastText(func(text string) map[string]float64 {
		m, err := ft.RunFasttext(text, nil)
		if err != nil || m == nil {
			return map[string]float64{}
		}
		return m
	})
	return nil
}

// EnableNgramsFromPath ports enableNgrams(File) — loads NGramDetector from ZIP (maxLength 50).
// Empty path leaves ngram disabled. Load failure returns error (Java RuntimeException).
func (d *DefaultLanguageIdentifier) EnableNgramsFromPath(ngramPath string) error {
	if d == nil || ngramPath == "" {
		return nil
	}
	ng, err := detector.NewNGramDetectorFromZip(ngramPath, 50)
	if err != nil {
		return err
	}
	d.NGram = ng
	return nil
}

func (d *DefaultLanguageIdentifier) Scores(cleanText string, noopLangs, preferredLangs []string, limitOnPreferred bool, count int) []languagetool.DetectedLanguage {
	if d == nil {
		return nil
	}
	// Java requireNonNull lists
	if noopLangs == nil {
		noopLangs = []string{}
	}
	if preferredLangs == nil {
		preferredLangs = []string{}
	}
	text := tools.JavaStringTrim(cleanText)
	if text == "" {
		return nil
	}

	// Java: prepareDetectLanguage first (nb→no, dominant unicode expand preferred, th/he/ko → null)
	var unicodeFn func(string) []string
	if d.Unicode != nil {
		unicodeFn = d.Unicode.GetDominantLangCodes
	}
	parsed := PrepareDetectLanguage(text, noopLangs, preferredLangs, unicodeFn)
	if parsed == nil {
		// Java: return singleton NoopLanguage when prepareDetect returns null
		src := "noop"
		return []languagetool.DetectedLanguage{
			languagetool.NewDetectedLanguageFull("", "zz", 1.0, &src),
		}
	}
	preferred := append([]string(nil), parsed.PreferredLangs...)
	additional := parsed.AdditionalLangs

	// Java DefaultLanguageIdentifier: text.length() <= CONSIDER_ONLY_PREFERRED_THRESHOLD
	// (UTF-16 units). Forces preferred-lang filter when short.
	if javaStringLen(text) <= ConsiderOnlyPreferredThreshold && len(preferred) > 0 {
		limitOnPreferred = true
	}

	scores := map[string]float64{}
	src := "profile"

	// Prefer fastText when available (Java: longer text → fasttext, short → ngram)
	if d.FastTextScore != nil && (javaStringLen(text) > ShortAlgoThreshold || d.NGram == nil) {
		for k, v := range d.FastTextScore(text) {
			scores[k] = v
		}
		src = "fasttext"
	}
	// NGram for short text when Java would use ngram over fasttext
	// Java: ngram.detectLanguages(text.trim(), additionalLangs)
	if d.NGram != nil && (len(scores) == 0 || javaStringLen(text) <= ShortAlgoThreshold) {
		for k, v := range d.NGram.DetectLanguagesAdditional(text, additional) {
			if cur, ok := scores[k]; !ok || v > cur {
				scores[k] = v
			}
		}
		if src != "fasttext" || javaStringLen(text) <= ShortAlgoThreshold {
			src = "ngram"
		}
	}
	// Merge profile scores if empty or low confidence (Java optimaize fallback ~0.85)
	maxScore := maxMap(scores)
	if d.ProfileScore != nil && (len(scores) == 0 || maxScore < 0.85) {
		for k, v := range d.ProfileScore(text, preferred) {
			if cur, ok := scores[k]; !ok || v > cur {
				scores[k] = v
			}
		}
		if src == "fasttext" && maxScore < 0.85 {
			src = "profile"
		} else if src == "" || len(scores) == 0 {
			src = "profile"
		}
	}
	// Common words boost when low confidence / zz (Java FASTTEXT_CONFIDENCE_THRESHOLD path)
	maxScore = maxMap(scores)
	topCode, _ := GetHighestScoringResult(scores)
	if d.CommonWords != nil && (maxScore < 0.85 || topCode == "zz" || len(scores) == 0) {
		counts := d.CommonWords.GetKnownWordsPerLanguage(text)
		// Java: scores.put(langCode, scores.get + count) then re-normalize later via ordering
		for k, c := range counts {
			scores[k] = scores[k] + float64(c)
		}
		if src != "" && !strings.Contains(src, "commonwords") {
			src = src + "+commonwords"
		}
	}

	// Filter ignore / preferred
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

var _ LanguageIdentifier = (*DefaultLanguageIdentifier)(nil)
