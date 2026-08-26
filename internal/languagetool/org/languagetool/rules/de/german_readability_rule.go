package de

import (
	"strconv"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanReadabilityRule ports org.languagetool.rules.de.GermanReadabilityRule
// (extends ReadabilityRule). Does not override match() — inherits core MatchList
// with polymorphic FRE / simpleSyllablesCount / getMessage via override hooks.
//
// Java ctor after super: Category CREATIVE_WRITING (category_creative_writing),
// Location.INTERNAL, false. Default off when defaultOn=false.
type GermanReadabilityRule struct {
	*rules.ReadabilityRule
}

// NewGermanReadabilityRule ports
// GermanReadabilityRule(messages, lang, userConfig, tooEasyTest)
// → full ctor with level=-1, defaultOn=false.
func NewGermanReadabilityRule(messages map[string]string, tooEasy bool) *GermanReadabilityRule {
	return NewGermanReadabilityRuleFull(messages, nil, tooEasy, -1, false)
}

// NewGermanReadabilityRuleFull ports
// GermanReadabilityRule(ResourceBundle, Language, UserConfig, boolean tooEasyTest, int level, boolean defaultOn).
func NewGermanReadabilityRuleFull(
	messages map[string]string,
	userConfig *languagetool.UserConfig,
	tooEasyTest bool,
	level int,
	defaultOn bool,
) *GermanReadabilityRule {
	// Java: super(messages, lang, userConfig, tooEasyTest, level, defaultOn);
	// then setCategory(CREATIVE_WRITING, …).
	base := rules.NewReadabilityRuleFull(messages, userConfig, tooEasyTest, level, defaultOn)
	base.Category = rules.CreativeWritingCategory(messages)

	r := &GermanReadabilityRule{ReadabilityRule: base}

	// Java polymorphic dispatch from ReadabilityRule.match() → DE overrides.
	base.FleschReadingEaseOverride = func(asl, asw float64) float64 {
		return r.GetFleschReadingEase(asl, asw)
	}
	base.SimpleSyllablesCountOverride = func(word string) int {
		return r.SimpleSyllablesCount(word)
	}
	base.MessageOverride = func(level, fre, asl, asw int) string {
		return r.GetMessage(level, fre, asl, asw)
	}
	return r
}

// GetID ports getId() / getId(tooEasyTest) with DE suffixes.
func (r *GermanReadabilityRule) GetID() string {
	tooEasy := r != nil && r.ReadabilityRule != nil && r.TooEasyTest
	return r.GetIDFor(tooEasy)
}

// GetIDFor ports getId(boolean tooEasyTest).
func (r *GermanReadabilityRule) GetIDFor(tooEasyTest bool) string {
	if tooEasyTest {
		return "READABILITY_RULE_SIMPLE_DE"
	}
	return "READABILITY_RULE_DIFFICULT_DE"
}

// GetDescription ports getDescription().
func (r *GermanReadabilityRule) GetDescription() string {
	if r != nil && r.ReadabilityRule != nil && r.TooEasyTest {
		return "Lesbarkeit: Zu einfacher Text"
	}
	return "Lesbarkeit: Zu schwieriger Text"
}

// GetConfigureText ports getConfigureText().
func (r *GermanReadabilityRule) GetConfigureText() string {
	return "Grad der Lesbarkeit 0 (sehr schwierig) bis 6 (sehr einfach):"
}

// GetRuleOptions ports getRuleOptions() — uses DE configure text (Java polymorphic getConfigureText).
func (r *GermanReadabilityRule) GetRuleOptions() []*rules.RuleOption {
	return []*rules.RuleOption{rules.NewRuleOption(3, r.GetConfigureText(), 0, 6)}
}

// PrintMessageLevel ports printMessageLevel(level) — German "Grad" labels.
func (r *GermanReadabilityRule) PrintMessageLevel(level int) string {
	var sLevel string
	switch level {
	case 0:
		sLevel = "Sehr schwer"
	case 1:
		sLevel = "Schwer"
	case 2:
		sLevel = "Mittelschwer"
	case 3:
		sLevel = "Mittel"
	case 4:
		sLevel = "Mittelleicht"
	case 5:
		sLevel = "Leicht"
	case 6:
		sLevel = "Sehr leicht"
	default:
		return ""
	}
	return " {Grad " + strconv.Itoa(level) + ": " + sLevel + "}"
}

// GetMessage ports getMessage(level, fre, asl, asw).
func (r *GermanReadabilityRule) GetMessage(level, fre, asl, asw int) string {
	simple, few := "schwierig", "viele"
	if r != nil && r.ReadabilityRule != nil && r.TooEasyTest {
		simple, few = "einfach", "wenige"
	}
	return "Lesbarkeit: Der Text dieses Absatzes ist zu " + simple + r.PrintMessageLevel(level) +
		". Zu " + few + " Wörter pro Satz und zu " + few + " Silben pro Wort."
}

// GetFleschReadingEase ports getFleschReadingEase — Amstad German FRE:
// 180 - asl - (58.5 * asw).
func (r *GermanReadabilityRule) GetFleschReadingEase(asl, asw float64) float64 {
	return 180 - asl - (58.5 * asw)
}

// SimpleSyllablesCount ports protected simpleSyllablesCount(String word):
// UTF-16 charAt/length; GermanTools.isVowel; diphthong lastDouble rules;
// nSyllables==0 → 1.
func (r *GermanReadabilityRule) SimpleSyllablesCount(word string) int {
	return simpleSyllablesCountDE(word)
}

// simpleSyllablesCountDE is the Java GermanReadabilityRule.simpleSyllablesCount body.
func simpleSyllablesCountDE(word string) int {
	n := utf16LenDE(word)
	if n == 0 {
		return 0
	}
	nSyllables := 0
	if IsVowel(javaCharAtDE(word, 0)) {
		nSyllables++
	}
	lastDouble := false
	for i := 1; i < n; i++ {
		c := javaCharAtDE(word, i)
		if IsVowel(c) {
			cl := javaCharAtDE(word, i-1)
			if lastDouble {
				nSyllables++
				lastDouble = false
			} else if ((c == 'i' || c == 'y') && (cl == 'a' || cl == 'e' || cl == 'A' || cl == 'E')) ||
				(c == 'u' && (cl == 'a' || cl == 'e' || cl == 'o' || cl == 'A' || cl == 'E' || cl == 'O')) ||
				(c == 'e' && (cl == 'e' || cl == 'i' || cl == 'E' || cl == 'I')) ||
				(c == 'a' && (cl == 'a' || cl == 'A')) {
				lastDouble = true
			} else {
				nSyllables++
				lastDouble = false
			}
		} else {
			lastDouble = false
		}
	}
	if nSyllables == 0 {
		return 1
	}
	return nSyllables
}
