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
// Java Pattern.matches() = full-string match → anchored ^...$.
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

// NewMorfologikRussianSpellerRule ports ctor with null UserConfig.
func NewMorfologikRussianSpellerRule() *MorfologikRussianSpellerRule {
	return NewMorfologikRussianSpellerRuleWithUser(nil)
}

// NewMorfologikRussianSpellerRuleWithUser ports
// MorfologikRussianSpellerRule(messages, language, userConfig, altLanguages).
func NewMorfologikRussianSpellerRuleWithUser(userConfig *languagetool.UserConfig) *MorfologikRussianSpellerRule {
	var userWords []string
	var accepted []string
	if userConfig != nil {
		accepted = userConfig.GetAcceptedWords()
		userWords = morfologik.UserDictWordsForMulti(accepted, userConfig.GetPremiumUid())
	}
	r := &MorfologikRussianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikRussianSpellerRuleID, "ru", RussianSpellerDict, nil),
		ConfCheckLatin: defaultMinRUValue,
	}
	// Java: if (userConfig != null) conf_ru_Value = (int) cf[0]
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID(MorfologikRussianSpellerRuleID); len(cf) > 0 && cf[0] != nil {
			if n, ok := confRUValueAsInt(cf[0]); ok {
				r.ConfCheckLatin = n
			}
		}
		r.SetUserConfig(userConfig)
	}
	// Java isLatinScript() = false
	if r.SpellingCheckRule != nil {
		r.NonLatinScript = true
		// Java filterNoSuggestWords (lcDoNotSuggestWords) via shared FilterSuggestions path.
		r.FilterNoSuggestWordsFn = r.filterNoSuggestWords
		// Java SpellingCheckRule: wordsToBeIgnored.addAll(userConfig.getAcceptedWords()) always.
		r.ApplyUserAcceptedWords(accepted)
	}
	r.IgnoreTokenFn = r.ruIgnoreToken
	// Java MorfologikSpellerRule.initSpeller: binary + spelling.txt + custom + global (+ user FSA if premium).
	r.InitSpellersFromGetters(nil, userWords)
	// Java: каждя → каждая
	r.AddExamplePair(
		rules.Wrong("Все счастливые семьи похожи друг на друга, <marker>каждя</marker> несчастливая семья несчастлива по-своему."),
		rules.Fixed("Все счастливые семьи похожи друг на друга, <marker>каждая</marker> несчастливая семья несчастлива по-своему."),
	)
	return r
}

func confRUValueAsInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
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

// GetConfigureText ports MorfologikRussianSpellerRule.getConfigureText.
func (r *MorfologikRussianSpellerRule) GetConfigureText() string {
	return "Проверять слова на латинице, только термины (0/1)"
}

// GetRuleOptions ports MorfologikRussianSpellerRule.getRuleOptions.
// Java: new RuleOption(DEFAULT_MIN_RU_VALUE, getConfigureText(), 0, 1)
func (r *MorfologikRussianSpellerRule) GetRuleOptions() []*rules.RuleOption {
	cfg := "Проверять слова на латинице, только термины (0/1)"
	if r != nil {
		cfg = r.GetConfigureText()
	}
	return []*rules.RuleOption{
		rules.NewRuleOption(defaultMinRUValue, cfg, 0, 1),
	}
}

// ruIgnoreToken ports ignoreToken:
//
//	if (conf_ru_Value != 1) && !RUSSIAN_LETTERS.matcher(word).matches() → true
//	else build words list → ignoreWord(words, idx)
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
	// Java: List<String> words from tokens; return ignoreWord(words, idx)
	words := make([]string, len(tokens))
	for i, tok := range tokens {
		if tok != nil {
			words[i] = tok.GetToken()
		}
	}
	if r != nil && r.SpellingCheckRule != nil {
		return r.IgnoreWordAt(words, idx)
	}
	return false
}

// filterNoSuggestWords ports filterNoSuggestWords (case-insensitive NOSUGGEST set).
// Java: k.getReplacement().toLowerCase() not in lcDoNotSuggestWords.
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
