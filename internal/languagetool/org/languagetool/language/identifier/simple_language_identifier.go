package identifier

import (
	"math"
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier/detector"
)

// SpellerFunc reports whether a word is misspelled for a language (Java SpellingCheckRule.isMisspelled).
type SpellerFunc func(word string) bool

// simpleWhitespace ports SimpleLanguageIdentifier.WHITESPACE = Pattern.compile("\\s+")
// without UNICODE_CHARACTER_CLASS (ASCII only; not Go RE2 \\s / NBSP).
var simpleWhitespace = regexp.MustCompile(`[ \t\n\v\f\r]+`)

// SimpleLanguageIdentifier ports
// org.languagetool.language.identifier.SimpleLanguageIdentifier.
// Uses pluggable spellcheckers instead of full SpellingCheckRule instances.
type SimpleLanguageIdentifier struct {
	BaseLanguageIdentifier
	// Spellers maps short lang code → isMisspelled.
	Spellers map[string]SpellerFunc
	// Unicode optional script dominance helper (Java UNICODE_BASED_LANG_IDENTIFIER).
	Unicode *detector.UnicodeBasedDetector
	// CommonWords optional boost (Java COMMON_WORDS_LANG_IDENTIFIER).
	CommonWords *detector.CommonWordsDetector
}

func NewSimpleLanguageIdentifier(maxLength int) *SimpleLanguageIdentifier {
	if maxLength <= 0 {
		maxLength = DefaultMaxLength
	}
	return &SimpleLanguageIdentifier{
		BaseLanguageIdentifier: NewBaseLanguageIdentifier(maxLength),
		Spellers:               map[string]SpellerFunc{},
		Unicode:                detector.NewUnicodeBasedDetector(),
	}
}

// NewSimpleLanguageIdentifierWith builds with preferred languages' spellers.
func NewSimpleLanguageIdentifierWith(preferred []string, spellers map[string]SpellerFunc) *SimpleLanguageIdentifier {
	s := NewSimpleLanguageIdentifier(DefaultMaxLength)
	if spellers != nil {
		for _, code := range preferred {
			if fn, ok := spellers[code]; ok {
				s.Spellers[code] = fn
			}
		}
		for k, fn := range spellers {
			if _, ok := s.Spellers[k]; !ok {
				s.Spellers[k] = fn
			}
		}
	}
	return s
}

// RegisterSpeller adds/replaces a language speller.
func (s *SimpleLanguageIdentifier) RegisterSpeller(langCode string, fn SpellerFunc) {
	if s.Spellers == nil {
		s.Spellers = map[string]SpellerFunc{}
	}
	s.Spellers[langCode] = fn
}

// Detect ports detectLanguage(cleanText, noop, preferred).
func (s *SimpleLanguageIdentifier) Detect(cleanText string, noopLangs, preferredLangs []string) *languagetool.DetectedLanguage {
	// Java SimpleLanguageIdentifier.detectLanguage(..., limitOnPreferred) ignores the flag.
	return s.DetectLimit(cleanText, noopLangs, preferredLangs, false)
}

// DetectLimit ports Java SimpleLanguageIdentifier — limitOnPreferredLangs is ignored.
func (s *SimpleLanguageIdentifier) DetectLimit(cleanText string, noopLangs, preferredLangs []string, _ bool) *languagetool.DetectedLanguage {
	if s == nil {
		return nil
	}
	if noopLangs == nil {
		noopLangs = []string{}
	}
	if preferredLangs == nil {
		preferredLangs = []string{}
	}

	var unicodeFn func(string) []string
	if s.Unicode != nil {
		unicodeFn = s.Unicode.GetDominantLangCodes
	}
	parsed := PrepareDetectLanguage(cleanText, noopLangs, preferredLangs, unicodeFn)
	if parsed == nil {
		// Java: return new DetectedLanguage(null, new NoopLanguage());
		src := "noop"
		dl := languagetool.NewDetectedLanguageFull("", "zz", 1.0, &src)
		return &dl
	}
	additionalLangs := parsed.AdditionalLangs
	preferred := parsed.PreferredLangs

	// Java: String[] words = WHITESPACE.split(cleanText); (no pre-trim)
	words := simpleWhitespace.Split(cleanText, -1)
	// Pattern.split discards trailing empties; drop pure trailing empties; keep leading empty.
	for len(words) > 0 && words[len(words)-1] == "" {
		words = words[:len(words)-1]
	}
	if len(words) == 0 || (len(words) == 1 && words[0] == "") {
		return nil
	}

	var dominant []string
	if s.Unicode != nil {
		dominant = s.Unicode.GetDominantLangCodes(cleanText)
	}
	dominantSet := map[string]bool{}
	for _, c := range dominant {
		dominantSet[c] = true
	}

	scores := map[string]float64{}
	source := "spellchecker"
	for code, spell := range s.Spellers {
		if spell == nil {
			continue
		}
		// Java: dominant.contains(key) ^ (dominant.isEmpty() && !NON_LATIN.contains(key))
		inDom := dominantSet[code]
		emptyDomAndLatin := len(dominant) == 0 && !isNonLatinLang(code)
		if inDom == emptyDomAndLatin { // !(a XOR b)
			continue
		}
		var errors float64
		for _, w := range words {
			if w == "" {
				continue
			}
			if spell(w) {
				errors++
			}
		}
		nWords := float64(len(words))
		if nWords == 0 {
			continue
		}
		scores[code] = 1.0 - errors/nWords
	}
	if len(scores) == 0 {
		scores["zz"] = 1.0
	}

	// common-words boost when low confidence / ties (Java order before preferred filter)
	_, ties := maxScoreStats(scores)
	topCode, topScore := highestScore(scores)
	if topScore < ScoreThreshold || topCode == "zz" || ties > 1 {
		if s.CommonWords != nil {
			baseHandled := map[string]struct{}{}
			for code, n := range s.CommonWords.GetKnownWordsPerLanguage(cleanText) {
				if _, ok := baseHandled[code]; ok {
					continue
				}
				baseHandled[code] = struct{}{}
				if old, ok := scores[code]; ok {
					scores[code] = old + float64(n)
				} else {
					scores[code] = float64(n)
				}
			}
			source += "+commonwords"
			topCode, topScore = highestScore(scores)
		}
	}

	// Special case no vs da (before preferred filter in Java)
	if containsStr(preferred, "no") && !containsStr(preferred, "da") {
		delete(scores, "da")
		topCode, topScore = highestScore(scores)
	}

	// Java SimpleLanguageIdentifier: cleanText.length() < CONSIDER_ONLY_PREFERRED_THRESHOLD
	// (UTF-16 units, not UTF-8 bytes).
	if javaStringLen(cleanText) < ConsiderOnlyPreferredThreshold && len(preferred) > 0 {
		for k := range scores {
			if !containsStr(preferred, k) {
				delete(scores, k)
			}
		}
		source += "+prefLang"
		topCode, topScore = highestScore(scores)
	}

	if topCode == "" {
		return nil
	}
	// Java: canLanguageBeDetected(key, additionalLangs)
	// Java: LanguageIdentifierService.INSTANCE.canLanguageBeDetected(key, additionalLangs)
	// When GlobalLanguages is empty (tests), treat registered spellers as supported.
	if !CanLanguageBeDetected(topCode, nil, additionalLangs) {
		if _, ok := s.Spellers[topCode]; !ok || topCode == "zz" {
			return nil
		}
	}
	src := source
	dl := languagetool.NewDetectedLanguageFull("", topCode, float32(topScore), &src)
	return &dl
}

func (s *SimpleLanguageIdentifier) Scores(cleanText string, noopLangs, preferredLangs []string, _ bool, count int) []languagetool.DetectedLanguage {
	d := s.Detect(cleanText, noopLangs, preferredLangs)
	if d == nil {
		return nil
	}
	return []languagetool.DetectedLanguage{*d}
}

func (s *SimpleLanguageIdentifier) CleanAndShortenText(text string) string {
	return s.BaseLanguageIdentifier.CleanAndShortenText(text)
}

func isNonLatinLang(code string) bool {
	for _, c := range NonLatinCharsLanguages {
		if c == code {
			return true
		}
	}
	return false
}

func highestScore(scores map[string]float64) (string, float64) {
	var best string
	bestV := math.Inf(-1)
	for k, v := range scores {
		if v > bestV {
			bestV = v
			best = k
		}
	}
	if best == "" {
		return "", 0
	}
	return best, bestV
}

func maxScoreStats(scores map[string]float64) (max float64, ties int) {
	max = math.Inf(-1)
	for _, v := range scores {
		if v > max {
			max = v
			ties = 1
		} else if v == max {
			ties++
		}
	}
	return max, ties
}

var _ LanguageIdentifier = (*SimpleLanguageIdentifier)(nil)
