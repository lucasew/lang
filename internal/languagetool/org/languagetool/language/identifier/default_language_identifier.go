package identifier

import (
	"math"
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
	// FastTextScore optional fastText stand-in (tests; no additionalLangs).
	FastTextScore func(cleanText string) map[string]float64
	// fastText is the real detector (Java fastTextDetector) — RunFasttext gets additionalLangs.
	fastText *detector.FastTextDetector
	// NGram ports Java NGramDetector (ZIP model or char-ngram test fallback).
	NGram *detector.NGramDetector
	// Unicode optional unicode script detector (Java UNICODE_BASED_LANG_IDENTIFIER).
	Unicode *detector.UnicodeBasedDetector
	// CommonWords optional common-words detector.
	CommonWords *detector.CommonWordsDetector

	MinimalConfidence float64
	// IgnoreLangCodes are never returned as top (ast, gl by default).
	IgnoreLangCodes map[string]struct{}
	// fasttextInitCounter ports DefaultLanguageIdentifier.fasttextInitCounter.
	fasttextInitCounter int
}

// FastTextConfidenceThreshold ports FASTTEXT_CONFIDENCE_THRESHOLD (0.85f).
const FastTextConfidenceThreshold = 0.85

func NewDefaultLanguageIdentifier(maxLength int) *DefaultLanguageIdentifier {
	if maxLength <= 0 {
		maxLength = DefaultMaxLength
	}
	languagetool.EnsureBuiltInLanguagesRegistered()
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

// GetFasttextInitCounter ports getFasttextInitCounter (tests).
func (d *DefaultLanguageIdentifier) GetFasttextInitCounter() int {
	if d == nil {
		return 0
	}
	return d.fasttextInitCounter
}

// reinitFasttextAfterFailure ports reinitFasttextAfterFailure.
func (d *DefaultLanguageIdentifier) reinitFasttextAfterFailure() {
	if d == nil || d.fastText == nil {
		return
	}
	d.fasttextInitCounter++
	ok, err := d.fastText.RestartProcess()
	if err != nil || !ok {
		// Java: decrement if restart did not reinit
		d.fasttextInitCounter--
	}
}

// EnableFastText installs a score function used preferentially for longer text (tests).
func (d *DefaultLanguageIdentifier) EnableFastText(score func(string) map[string]float64) {
	d.FastTextScore = score
}

func (d *DefaultLanguageIdentifier) IsFastTextEnabled() bool {
	return d != nil && (d.FastTextScore != nil || d.fastText != nil)
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
// Both paths required; creates FastTextDetector (additionalLangs passed at detect time).
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
	d.fastText = ft
	return nil
}

// SetFastTextDetector ports setFastTextDetector (tests).
func (d *DefaultLanguageIdentifier) SetFastTextDetector(ft *detector.FastTextDetector) {
	if d != nil {
		d.fastText = ft
	}
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
	src := ""
	textLen := javaStringLen(text)
	fasttextFailed := false
	usingFastText := false
	hasFTOrNGram := d.fastText != nil || d.FastTextScore != nil || d.NGram != nil

	// Prefer fastText when available (Java: longer text → fasttext, short → ngram)
	// Java: try { … runFasttext / ngram … } catch → fasttextFailed + reinit
	if hasFTOrNGram {
		useFastText := (d.fastText != nil || d.FastTextScore != nil) &&
			(textLen > ShortAlgoThreshold || d.NGram == nil)
		if useFastText {
			usingFastText = true
			if d.fastText != nil {
				m, err := d.fastText.RunFasttext(text, additional)
				if err != nil {
					// FastTextException.isDisabled or other → reinit + fallback
					if fe, ok := err.(*detector.FastTextException); ok && fe.IsDisabled() {
						d.reinitFasttextAfterFailure()
					} else {
						d.reinitFasttextAfterFailure()
					}
					fasttextFailed = true
				} else if m != nil {
					for k, v := range m {
						scores[k] = v
					}
					src += "fasttext"
				}
			} else if d.FastTextScore != nil {
				for k, v := range d.FastTextScore(text) {
					scores[k] = v
				}
				src += "fasttext"
			}
		}
		// NGram for short text when Java would use ngram over fasttext
		if !fasttextFailed && d.NGram != nil && (len(scores) == 0 || textLen <= ShortAlgoThreshold) {
			for k, v := range d.NGram.DetectLanguagesAdditional(text, additional) {
				if cur, ok := scores[k]; !ok || v > cur {
					scores[k] = v
				}
			}
			if src == "" || textLen <= ShortAlgoThreshold {
				if !strings.Contains(src, "ngram") {
					src += "ngram"
				}
			}
		}
		// Common words when fasttext low conf or zz (Java FASTTEXT_CONFIDENCE_THRESHOLD 0.85)
		topCode, topScore := GetHighestScoringResult(scores)
		if (usingFastText && topScore < FastTextConfidenceThreshold) || topCode == "zz" {
			if d.CommonWords != nil {
				for k, c := range d.CommonWords.GetKnownWordsPerLanguage(text) {
					scores[k] = scores[k] + float64(c)
				}
				if !strings.Contains(src, "commonwords") {
					src += "+commonwords"
				}
			}
		}
		// Special case: preferred no without da → drop da
		if containsStr(preferred, "no") && !containsStr(preferred, "da") {
			delete(scores, "da")
		}
		// Preferred filter for short text or forcePreferred
		if len(preferred) > 0 && (textLen <= ConsiderOnlyPreferredThreshold || limitOnPreferred) {
			for k := range scores {
				if !containsStr(preferred, k) {
					delete(scores, k)
				}
			}
			src += "+prefLang(forced: " + boolStr(limitOnPreferred) + ")"
		}
	}

	// Java: if (fastTextDetector == null && ngram == null || fasttextFailed) {
	//   text = textObjectFactory.forText(text).toString();
	//   source += "+fallback";
	//   localResult = detectLanguageCode(text, preferredLangs, limitOnPreferredLangs);
	//   scores.put(localResult);
	// }
	// ProfileScore is the optimaize languageDetector stand-in — only on this path.
	if !hasFTOrNGram || fasttextFailed {
		src += "+fallback"
		factoryText := ApplyTextObjectFactoryFilters(text)
		if code, prob, ok := d.detectLanguageCode(factoryText, preferred, limitOnPreferred); ok {
			scores[code] = prob
		}
	}
	// Common words boost for profile-only path when still low
	maxScore := maxMap(scores)
	topCode, _ := GetHighestScoringResult(scores)
	if d.CommonWords != nil && (maxScore < ScoreThreshold || topCode == "zz" || len(scores) == 0) {
		if !strings.Contains(src, "commonwords") {
			for k, c := range d.CommonWords.GetKnownWordsPerLanguage(text) {
				scores[k] = scores[k] + float64(c)
			}
			if src != "" {
				src += "+commonwords"
			}
		}
	}
	if src == "" {
		src = "profile"
	}

	// Filter ignore (preferred already applied for force path above; re-apply for profile)
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
	// Java: only emit if canLanguageBeDetected(code, additionalLangs)
	for i := 0; i < len(pairs) && i < count; i++ {
		code := pairs[i].code
		if code == "" || !CanLanguageBeDetected(code, nil, additional) {
			continue
		}
		s := src
		score := float32(pairs[i].score)
		// Java: fasttext confidence rewritten for short-text unreliability
		// newScore = 0.99 / (30.0 / min(text.length(), 30))
		if count == 1 && strings.Contains(src, "fasttext") {
			denom := 30.0 / float64(minInt(textLen, 30))
			if denom > 0 {
				score = float32(0.99 / denom)
			}
		} else if count > 1 {
			// Java: Math.round(value * 100.0) / 100.0f
			score = float32(math.Round(float64(pairs[i].score)*100.0) / 100.0)
		}
		out = append(out, languagetool.NewDetectedLanguageFull("", code, score, &s))
	}
	// Java: empty → fallbackToPrefLang with confidence 0.1
	if len(out) == 0 && len(preferred) > 0 {
		pref := tools.JavaStringTrim(preferred[0])
		if pref != "" && CanLanguageBeDetected(pref, nil, nil) {
			s := src + "+fallbackToPrefLang"
			out = append(out, languagetool.NewDetectedLanguageFull("", pref, 0.1, &s))
		}
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// detectLanguageCode ports DefaultLanguageIdentifier.detectLanguageCode.
// Uses ProfileScore as LanguageDetector.getProbabilities stand-in; returns top hit
// after optional preferred-language filter (limitOnPreferredLangs).
func (d *DefaultLanguageIdentifier) detectLanguageCode(text string, preferred []string, limitOnPreferred bool) (code string, prob float64, ok bool) {
	if d == nil || d.ProfileScore == nil {
		return "", 0, false
	}
	// Optimaize getProbabilities → list of (locale, probability); we get a score map.
	raw := d.ProfileScore(text, preferred)
	type hit struct {
		code string
		prob float64
	}
	var list []hit
	for k, v := range raw {
		list = append(list, hit{k, v})
	}
	// Java: if limitOnPreferredLangs && preferred non-empty → remove non-preferred
	if limitOnPreferred && len(preferred) > 0 {
		filtered := list[:0]
		for _, h := range list {
			// Java: preferredLangs.contains(l.getLocale().getLanguage()) — short language code
			lang := h.code
			if i := strings.IndexByte(lang, '-'); i >= 0 {
				lang = lang[:i]
			}
			if containsStr(preferred, lang) || containsStr(preferred, h.code) {
				filtered = append(filtered, h)
			}
		}
		list = filtered
	}
	if len(list) == 0 {
		return "", 0, false
	}
	// highest probability first
	best := list[0]
	for _, h := range list[1:] {
		if h.prob > best.prob {
			best = h
		}
	}
	// Java: code = lang.get(0).getLocale().getLanguage() — ISO language short code
	code = best.code
	if i := strings.IndexByte(code, '-'); i >= 0 {
		code = code[:i]
	}
	return code, best.prob, true
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
