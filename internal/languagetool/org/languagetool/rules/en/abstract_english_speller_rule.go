package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// English spelling resource paths used by AbstractEnglishSpellerRule.
const (
	EnglishCustomSpellingFile = "/en/hunspell/spelling.txt" // language-relative custom
	EnglishGlobalSpellingFile = "/spelling.txt"
	EnglishMultiwordsFile     = "/en/multiwords.txt"
)

// AbstractEnglishSpellerRule ports
// org.languagetool.rules.en.AbstractEnglishSpellerRule surface over MorfologikSpellerRule.
type AbstractEnglishSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	LanguageShortCode string // e.g. "en"
	// VariantCode for file paths (en-US etc.).
	VariantCode string
	// Synthesize optional English synthesizer for irregular-form suggestions (fail-closed when nil).
	Synthesize SynthesizeFn
	// IsValidInOtherVariantFn ports isValidInOtherVariant (variant spellers set this).
	IsValidInOtherVariantFn func(word string) *VariantInfo
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewAbstractEnglishSpellerRule(id, variantCode string, speller *morfologik.MorfologikSpeller) *AbstractEnglishSpellerRule {
	short := "en"
	if i := strings.Index(variantCode, "-"); i > 0 {
		short = variantCode[:i]
	} else if variantCode != "" {
		short = variantCode
	}
	base := morfologik.NewMorfologikSpellerRule(id, short, "", speller)
	// Java AbstractEnglishSpellerRule: setCheckCompound(true) with default compoundRegex "-"
	base.SetCheckCompound(true)
	// Java AbstractEnglishSpellerRule: super.ignoreWordsWithLength = 1
	// tokenizeNewWords() = false — re-load lists as whole-line ignores.
	if base.SpellingCheckRule != nil {
		base.IgnoreWordsWithLength = 1
		base.DisableTokenizeNewWords = true
		spelling.ReapplyDefaultSpellingWordLists(base.SpellingCheckRule)
		// Java AbstractEnglishSpellerRule.filterNoSuggestWords (lcDoNotSuggestWords).
		base.FilterNoSuggestWordsFn = filterEnglishNoSuggestWords
		// Java filterSuggestions CONTAINS_TOKEN arm after super.
		base.FilterSuggestionsExtraFn = filterEnglishContainsToken
		// Java getOnlySuggestions early-return in calcSpellerSuggestions.
		base.GetOnlySuggestionsFn = EnglishOnlySuggestions
		// Java getAdditionalTopSuggestions curated maps + ys→ies.
		base.GetAdditionalTopSuggestionsFn = func(existing []string, word string) []string {
			return EnglishAdditionalTopSuggestions(word, base.IsMisspelled)
		}
		// Java addHyphenSuggestions for multi-part hyphenated misspellings.
		base.AddHyphenSuggestionsFn = func(parts []string) []string {
			return EnglishAddHyphenSuggestions(parts, base.IsMisspelled, func(w string) []string {
				if base.Speller == nil {
					return nil
				}
				return base.Speller.FindReplacements(w)
			})
		}
	}
	r := &AbstractEnglishSpellerRule{
		MorfologikSpellerRule: base,
		LanguageShortCode:     short,
		VariantCode:           variantCode,
	}
	// Java AbstractEnglishSpellerRule: sentenc → sentence
	r.AddExamplePair(
		rules.Wrong("This <marker>sentenc</marker> contains a spelling mistake."),
		rules.Fixed("This <marker>sentence</marker> contains a spelling mistake."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *AbstractEnglishSpellerRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AbstractEnglishSpellerRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AbstractEnglishSpellerRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// Match ports parent Match + getRuleMatches irregular forms / other-variant rewrite.
func (r *AbstractEnglishSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil {
		return nil, nil
	}
	base, err := r.MorfologikSpellerRule.Match(sentence)
	if err != nil || len(base) == 0 {
		return base, err
	}
	for _, m := range base {
		if m == nil {
			continue
		}
		word := matchSurfaceEN(m, sentence)
		if word == "" {
			continue
		}
		// Java: irregular forms preferred over dialect variant
		if forms := EnglishIrregularForms(word, r.IsMisspelled, r.Synthesize); forms != nil && len(forms.Forms) > 0 {
			m.SetSuggestedReplacements(append([]string(nil), forms.Forms...))
			continue
		}
		if vi := r.isValidInOtherVariant(word); vi != nil {
			sug := vi.GetOtherVariant()
			if tools.StartsWithUppercase(word) {
				sug = tools.UppercaseFirstChar(sug)
			}
			m.SetSuggestedReplacements([]string{sug})
		}
		// Java: setLazySuggestedReplacements(() -> cleanSuggestions(m))
		if sugs := m.GetSuggestedReplacements(); len(sugs) > 0 {
			m.SetSuggestedReplacements(EnglishCleanSuggestions(sugs))
		}
	}
	return base, nil
}

func (r *AbstractEnglishSpellerRule) isValidInOtherVariant(word string) *VariantInfo {
	if r == nil || r.IsValidInOtherVariantFn == nil {
		return nil
	}
	return r.IsValidInOtherVariantFn(word)
}

func matchSurfaceEN(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
	if m == nil || sent == nil {
		return ""
	}
	text := sent.GetText()
	from, to := m.GetFromPos(), m.GetToPos()
	if from < 0 || from >= to {
		return ""
	}
	runes := []rune(text)
	if to <= len(runes) {
		return string(runes[from:to])
	}
	if to <= len(text) {
		return text[from:to]
	}
	return ""
}

// GetAdditionalSpellingFileNames ports getAdditionalSpellingFileNames.
func (r *AbstractEnglishSpellerRule) GetAdditionalSpellingFileNames() []string {
	custom := r.LanguageShortCode + "/hunspell/spelling.txt"
	if r.VariantCode != "" {
		// Java: language.getShortCode() + CUSTOM_SPELLING_FILE
		custom = r.LanguageShortCode + "/hunspell/spelling.txt"
	}
	return []string{custom, EnglishGlobalSpellingFile, EnglishMultiwordsFile}
}

// IsDoNotSuggest reports whether the word is blocked from suggestions.
func IsDoNotSuggest(word string) bool {
	_, ok := EnglishDoNotSuggestWords[strings.ToLower(strings.TrimSpace(word))]
	return ok
}

// FilterEnglishSuggestions drops blocked suggestions (NOSUGGEST list).
func FilterEnglishSuggestions(suggestions []string) []string {
	var out []string
	for _, s := range suggestions {
		if IsDoNotSuggest(s) {
			continue
		}
		out = append(out, s)
	}
	return out
}

// filterEnglishNoSuggestWords ports AbstractEnglishSpellerRule.filterNoSuggestWords.
func filterEnglishNoSuggestWords(suggestions []string) []string {
	return FilterEnglishSuggestions(suggestions)
}
