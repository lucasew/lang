package identifier

import (
	"math"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier/detector"
)

// SpellerFunc reports whether a word is misspelled for a language.
type SpellerFunc func(word string) bool

// SimpleLanguageIdentifier ports
// org.languagetool.language.identifier.SimpleLanguageIdentifier.
// Uses pluggable spellcheckers instead of full SpellingCheckRule instances.
type SimpleLanguageIdentifier struct {
	BaseLanguageIdentifier
	// Spellers maps short lang code → isMisspelled.
	Spellers map[string]SpellerFunc
	// Unicode optional script dominance helper.
	Unicode *detector.UnicodeBasedDetector
	// CommonWords optional boost.
	CommonWords *detector.CommonWordsDetector
}

func NewSimpleLanguageIdentifier(maxLength int) *SimpleLanguageIdentifier {
	if maxLength <= 0 {
		maxLength = DefaultMaxLength
	}
	return &SimpleLanguageIdentifier{
		BaseLanguageIdentifier: NewBaseLanguageIdentifier(maxLength),
		Spellers:               map[string]SpellerFunc{},
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
		// also keep others if map larger
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

// Detect scores languages by inverse spelling error rate.
func (s *SimpleLanguageIdentifier) Detect(cleanText string, noopLangs, preferredLangs []string) *languagetool.DetectedLanguage {
	if s == nil || strings.TrimSpace(cleanText) == "" {
		return nil
	}
	words := strings.Fields(cleanText)
	if len(words) == 0 {
		return nil
	}

	dominant := map[string]bool{}
	if s.Unicode != nil {
		for _, c := range s.Unicode.GetDominantLangCodes(cleanText) {
			dominant[c] = true
		}
	}

	scores := map[string]float64{}
	source := "spellchecker"
	for code, spell := range s.Spellers {
		if spell == nil {
			continue
		}
		// filter by dominant script when available
		if len(dominant) > 0 {
			if !dominant[code] {
				// allow latin langs when no dominant non-latin
				nonLatin := false
				for d := range dominant {
					if isNonLatinLang(d) {
						nonLatin = true
						break
					}
				}
				if nonLatin {
					continue
				}
			}
		} else if isNonLatinLang(code) {
			// skip non-latin spellers when no dominant non-latin signal
			continue
		}
		var errors float64
		for _, w := range words {
			if spell(w) {
				errors++
			}
		}
		scores[code] = 1.0 - errors/float64(len(words))
	}
	if len(scores) == 0 {
		scores["zz"] = 1.0
	}

	// common-words boost when low confidence / ties
	maxVal, ties := maxScoreStats(scores)
	topCode, topScore := highestScore(scores)
	if topScore < ScoreThreshold || topCode == "zz" || ties > 1 {
		if s.CommonWords != nil {
			// boost known word counts when available
			for code, n := range s.CommonWords.GetKnownWordsPerLanguage(cleanText) {
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
	_ = maxVal

	// preferred-lang filter for short text
	if len(cleanText) < ConsiderOnlyPreferredThreshold && len(preferredLangs) > 0 {
		for k := range scores {
			if !containsStr(preferredLangs, k) {
				delete(scores, k)
			}
		}
		source += "+prefLang"
		topCode, topScore = highestScore(scores)
	}

	// special: no vs da
	if containsStr(preferredLangs, "no") && !containsStr(preferredLangs, "da") {
		delete(scores, "da")
		topCode, topScore = highestScore(scores)
	}

	if topCode == "" || topCode == "zz" {
		return nil
	}
	// noop langs
	if containsStr(noopLangs, topCode) {
		return nil
	}
	src := source
	dl := languagetool.NewDetectedLanguageFull("", topCode, float32(topScore), &src)
	return &dl
}

func (s *SimpleLanguageIdentifier) Scores(cleanText string, noopLangs, preferredLangs []string, limitOnPreferred bool, count int) []languagetool.DetectedLanguage {
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
