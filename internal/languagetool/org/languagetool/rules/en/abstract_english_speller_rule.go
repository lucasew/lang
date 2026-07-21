package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ner"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// English spelling resource paths used by AbstractEnglishSpellerRule.
// Match Java AbstractEnglishSpellerRule.getAdditionalSpellingFileNames:
// short + CUSTOM_SPELLING_FILE, GLOBAL_SPELLING_FILE, "/en/multiwords.txt".
const (
	EnglishCustomSpellingFile = "en/hunspell/spelling_custom.txt"
	EnglishGlobalSpellingFile = "spelling_global.txt"
	EnglishMultiwordsFile     = "/en/multiwords.txt"
)

// AbstractEnglishSpellerRule ports
// org.languagetool.rules.en.AbstractEnglishSpellerRule surface over MorfologikSpellerRule.
type AbstractEnglishSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	LanguageShortCode string // e.g. "en"
	// VariantCode for file paths (en-US etc.).
	VariantCode string
	// LanguageName ports language.getName() (short description on variant suggestions).
	LanguageName string
	// Synthesize optional English synthesizer for irregular-form suggestions (fail-closed when nil).
	Synthesize SynthesizeFn
	// IsValidInOtherVariantFn ports isValidInOtherVariant (variant spellers set this).
	IsValidInOtherVariantFn func(word string) *VariantInfo
	// LanguageModel ports BaseLanguageModel for NER filter (nil → skip NER arm).
	LanguageModel CountProvider
	// NERPipe ports AbstractEnglishSpellerRule.nerPipe (nil → skip NER arm).
	NERPipe *ner.NERService
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
	r := &AbstractEnglishSpellerRule{
		MorfologikSpellerRule: base,
		LanguageShortCode:     short,
		VariantCode:           variantCode,
		LanguageName:          englishLanguageName(variantCode),
	}
	// Java AbstractEnglishSpellerRule: super.ignoreWordsWithLength = 1
	// tokenizeNewWords() = false — re-load lists as whole-line ignores.
	// Path getters: EN multiwords via getAdditionalSpellingFileNames (or SpellingCheckRule en case).
	// languageSpecificIgnoreFile = spelling.txt → spelling_{variant}.txt (same as variant file).
	if base.SpellingCheckRule != nil {
		base.IgnoreWordsWithLength = 1
		base.DisableTokenizeNewWords = true
		// Prefer full variant for getLanguageVariantSpellingFileName / languageSpecificIgnoreFile.
		if variantCode != "" {
			base.LanguageCode = variantCode
		}
		base.GetAdditionalSpellingFileNamesFn = r.GetAdditionalSpellingFileNames
		base.GetLanguageVariantSpellingFileNameFn = func() string {
			return spelling.LanguageVariantSpellingClasspath(r.VariantCode)
		}
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

// Match ports parent Match + getRuleMatches irregular forms / other-variant rewrite
// + Match-level variant blog URLs (Java AbstractEnglishSpellerRule.match).
func (r *AbstractEnglishSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || r.MorfologikSpellerRule == nil {
		return nil, nil
	}
	base, err := r.MorfologikSpellerRule.Match(sentence)
	if err != nil || len(base) == 0 {
		return base, err
	}
	// Java Match: NER filter when languageModel is BaseLanguageModel && nerPipe != null
	// and sentenceText.length() <= 250.
	if r.LanguageModel != nil && r.NERPipe != nil && sentence != nil {
		sentenceText := sentence.GetText()
		if len(sentenceText) <= 250 {
			// Java: catch Exception → warn and keep matches
			func() {
				defer func() { _ = recover() }()
				named := r.NERPipe.RunNER(sentenceText)
				base = filterNERMatches(base, sentenceText, named, r.LanguageModel)
			}()
		}
	}
	for _, m := range base {
		if m == nil {
			continue
		}
		word := matchSurfaceEN(m, sentence)
		if word == "" {
			continue
		}
		// Java getRuleMatches: irregular forms preferred over dialect variant
		if forms := EnglishIrregularForms(word, r.IsMisspelled, r.Synthesize); forms != nil && len(forms.Forms) > 0 {
			// Java addFormsToFirstMatch: message + forms prepended to existing suggestions
			m.Message = "Possible spelling mistake. Did you mean <suggestion>" + forms.Forms[0] +
				"</suggestion>, the " + forms.FormName + " form of the " + forms.PosName +
				" '" + forms.BaseForm + "'?"
			old := m.GetSuggestedReplacements()
			merged := append([]string(nil), forms.Forms...)
			seen := map[string]bool{}
			for _, f := range forms.Forms {
				seen[f] = true
			}
			for _, o := range old {
				if !seen[o] {
					merged = append(merged, o)
					seen[o] = true
				}
			}
			m.SetSuggestedReplacements(merged)
			if sugs := m.GetSuggestedReplacements(); len(sugs) > 0 {
				m.SetSuggestedReplacements(EnglishCleanSuggestions(sugs))
			}
			continue
		}
		if vi := r.isValidInOtherVariant(word); vi != nil {
			// Java replaceFormsOfFirstMatch
			m.Message = "Possible spelling mistake. '" + word + "' is " + vi.GetVariantName() + "."
			sug := vi.GetOtherVariant()
			if tools.StartsWithUppercase(word) {
				sug = tools.UppercaseFirstChar(sug)
			}
			// Java: sugg.setShortDescription(language.getName())
			desc := r.LanguageName
			if desc == "" {
				desc = englishLanguageName(r.VariantCode)
			}
			obj := rules.NewSuggestedReplacementWithDesc(sug, &desc)
			m.SetSuggestedReplacementObjects([]*rules.SuggestedReplacement{obj})
			// Java Match: setUrl for variant blog when isValidInOtherVariant
			if u := enVariantBlogURL(word); u != "" {
				m.SetURL(u)
			}
			// Java getRuleMatches still wraps with cleanSuggestions for all matches
			if sugs := m.GetSuggestedReplacements(); len(sugs) > 0 {
				m.SetSuggestedReplacements(EnglishCleanSuggestions(sugs))
			}
			continue
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

// matchSurfaceEN extracts match surface using Java UTF-16 from/to positions.
func matchSurfaceEN(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
	if m == nil || sent == nil {
		return ""
	}
	return rules.UTF16Substring(sent.GetText(), m.GetFromPos(), m.GetToPos())
}

// GetAdditionalSpellingFileNames ports AbstractEnglishSpellerRule.getAdditionalSpellingFileNames:
// shortCode + CUSTOM_SPELLING_FILE, GLOBAL_SPELLING_FILE, "/en/multiwords.txt".
func (r *AbstractEnglishSpellerRule) GetAdditionalSpellingFileNames() []string {
	sc := r.LanguageShortCode
	if sc == "" {
		sc = "en"
	}
	return []string{sc + "/hunspell/spelling_custom.txt", EnglishGlobalSpellingFile, EnglishMultiwordsFile}
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

// englishLanguageName maps variant codes to Java Language.getName() twins used by
// replaceFormsOfFirstMatch short descriptions.
func englishLanguageName(variantCode string) string {
	switch variantCode {
	case "en-US":
		return "English (US)"
	case "en-GB":
		return "English (GB)"
	case "en-CA":
		return "English (Canadian)"
	case "en-AU":
		return "English (Australian)"
	case "en-NZ":
		return "English (New Zealand)"
	case "en-ZA":
		return "English (South African)"
	case "en", "":
		return "English"
	default:
		return "English"
	}
}
