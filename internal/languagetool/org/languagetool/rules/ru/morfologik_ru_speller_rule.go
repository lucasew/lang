package ru

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	// MorfologikRussianSpellerRuleID ports MorfologikRussianSpellerRule.RULE_ID / getId().
	// Java: "MORFOLOGIK_RULE_RU_RU" (not MORFOLOGIK_RULE_RU).
	MorfologikRussianSpellerRuleID = "MORFOLOGIK_RULE_RU_RU"
	// RussianSpellerDict ports MorfologikRussianSpellerRule.RESOURCE_FILENAME / getFileName().
	// Java: "/ru/hunspell/ru_RU.dict"
	RussianSpellerDict = "/ru/hunspell/ru_RU.dict"
	// Java DEFAULT_MIN_RU_VALUE — 0 = skip non-Russian-letter tokens.
	defaultMinRUValue = 0
)

// Java RUSSIAN_LETTERS (with combining acute/grave on vowels, ё, ʼ, hyphen).
// RE2 has no \u escapes in some builds — use literal combining marks in character class.
var russianLetters = regexp.MustCompile(`^[-а-яёА-ЯЁʼо́а́е́у́и́ы́э́ю́я́о̀а̀ѐу̀ѝы̀э̀ю̀я̀]*$`)

// Java lcDoNotSuggestWords (NOSUGGEST) for MorfologikRussianSpellerRule.
var ruLcDoNotSuggest = map[string]struct{}{
	"блоггер": {}, "дрочим": {}, "анальный": {}, "орочем": {},
}

// MorfologikRussianSpellerRule ports rules.ru.MorfologikRussianSpellerRule.
// filterNoSuggestWords + ignoreToken letter-gate; conf_ru_Value for Latin check.
type MorfologikRussianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	// ConfCheckLatin ports conf_ru_Value: 1 = also check Latin-script tokens (Java RuleOption).
	// Default 0 = ignore tokens that do not fully match RUSSIAN_LETTERS.
	ConfCheckLatin int
	// ExtraDoNotSuggest merges language-variant NOSUGGEST (YO adds "елка").
	ExtraDoNotSuggest map[string]struct{}
	// incorrectExamples / correctExamples port Rule.addExamplePair (not on SpellingCheckRule:
	// import cycle with rules package — same pattern as AbstractEnglishSpellerRule).
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewMorfologikRussianSpellerRule() *MorfologikRussianSpellerRule {
	r := &MorfologikRussianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikRussianSpellerRuleID, "ru", RussianSpellerDict, nil),
		ConfCheckLatin: defaultMinRUValue,
	}
	// Java isLatinScript() = false
	if r.SpellingCheckRule != nil {
		r.NonLatinScript = true
		// Java filterNoSuggestWords (lcDoNotSuggestWords) via shared FilterSuggestions path.
		r.FilterNoSuggestWordsFn = r.filterNoSuggestWords
	}
	r.IgnoreTokenFn = r.ruIgnoreToken
	// Java: каждя → каждая
	r.AddExamplePair(
		rules.Wrong("Все счастливые семьи похожи друг на друга, <marker>каждя</marker> несчастливая семья несчастлива по-своему."),
		rules.Fixed("Все счастливые семьи похожи друг на друга, <marker>каждая</marker> несчастливая семья несчастлива по-своему."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *MorfologikRussianSpellerRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *MorfologikRussianSpellerRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *MorfologikRussianSpellerRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// ruIgnoreToken ports ignoreToken: skip non-Russian-letter tokens when conf != 1.
func (r *MorfologikRussianSpellerRule) ruIgnoreToken(tokens []*languagetool.AnalyzedTokenReadings, idx int) bool {
	if idx < 0 || idx >= len(tokens) || tokens[idx] == nil {
		return false
	}
	word := tokens[idx].GetToken()
	// Java: if (conf_ru_Value != 1) && !RUSSIAN_LETTERS.matcher(word).matches() → true
	conf := defaultMinRUValue
	if r != nil {
		conf = r.ConfCheckLatin
	}
	if conf != 1 && !russianLetters.MatchString(word) {
		return true
	}
	if r != nil && r.SpellingCheckRule != nil {
		return r.IgnoreWord(word)
	}
	return false
}

// Match ports parent Match + filterNoSuggestWords on suggestions.
func (r *MorfologikRussianSpellerRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
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
		sugs := m.GetSuggestedReplacements()
		if len(sugs) == 0 {
			// try filter dict suggestions when map empty
			if FilterDictAvailable() {
				word := matchSurfaceRU(m, sentence)
				sugs = FilterDictSuggest(word)
			}
		}
		if len(sugs) > 0 {
			m.SetSuggestedReplacements(r.filterNoSuggestWords(sugs))
		}
	}
	return base, nil
}

// filterNoSuggestWords ports filterNoSuggestWords (case-insensitive NOSUGGEST set).
func (r *MorfologikRussianSpellerRule) filterNoSuggestWords(suggestions []string) []string {
	if len(suggestions) == 0 {
		return suggestions
	}
	out := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		low := strings.ToLower(s)
		if _, bad := ruLcDoNotSuggest[low]; bad {
			continue
		}
		if r != nil && r.ExtraDoNotSuggest != nil {
			if _, bad := r.ExtraDoNotSuggest[low]; bad {
				continue
			}
		}
		out = append(out, s)
	}
	return out
}

func matchSurfaceRU(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
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
