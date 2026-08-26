package pt

import (
	"strconv"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseReadabilityRule ports org.languagetool.rules.pt.PortugueseReadabilityRule
// (extends ReadabilityRule). Does not override match() — inherits core MatchList
// with polymorphic FRE / simpleSyllablesCount / getMessage via override hooks.
//
// Java ctor after super: Category TEXT_ANALYSIS ("Análise de Texto"),
// Location.INTERNAL, false. Default off when defaultOn=false.
type PortugueseReadabilityRule struct {
	*rules.ReadabilityRule
}

// NewPortugueseReadabilityRule ports
// PortugueseReadabilityRule(messages, lang, userConfig, tooEasyTest, level)
// → full ctor with defaultOn=false (common call site: level=-1 → default 3).
func NewPortugueseReadabilityRule(tooEasy bool, level int) *PortugueseReadabilityRule {
	return NewPortugueseReadabilityRuleFull(nil, nil, tooEasy, level, false)
}

// NewPortugueseReadabilityRuleFull ports
// PortugueseReadabilityRule(ResourceBundle, Language, UserConfig, boolean tooEasyTest, int level, boolean defaultOn).
func NewPortugueseReadabilityRuleFull(
	messages map[string]string,
	userConfig *languagetool.UserConfig,
	tooEasyTest bool,
	level int,
	defaultOn bool,
) *PortugueseReadabilityRule {
	// Java: super(messages, lang, userConfig, tooEasyTest, level, defaultOn);
	// then setCategory(TEXT_ANALYSIS, "Análise de Texto", INTERNAL, false).
	base := rules.NewReadabilityRuleFull(messages, userConfig, tooEasyTest, level, defaultOn)
	base.Category = rules.NewCategoryFull(
		rules.NewCategoryId("TEXT_ANALYSIS"),
		"Análise de Texto",
		rules.CategoryInternal,
		false,
		"",
	)

	r := &PortugueseReadabilityRule{ReadabilityRule: base}

	// Java polymorphic dispatch from ReadabilityRule.match() → PT overrides.
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

// GetID ports getId() / getId(tooEasyTest) with PT suffixes.
func (r *PortugueseReadabilityRule) GetID() string {
	tooEasy := r != nil && r.ReadabilityRule != nil && r.TooEasyTest
	return r.GetIDFor(tooEasy)
}

// GetIDFor ports getId(boolean tooEasyTest).
func (r *PortugueseReadabilityRule) GetIDFor(tooEasyTest bool) string {
	if tooEasyTest {
		return "READABILITY_RULE_SIMPLE_PT"
	}
	return "READABILITY_RULE_DIFFICULT_PT"
}

// GetDescription ports getDescription().
func (r *PortugueseReadabilityRule) GetDescription() string {
	if r != nil && r.ReadabilityRule != nil && r.TooEasyTest {
		return "Legibilidade: texto demasiado simples"
	}
	return "Legibilidade: texto demasiado complexo"
}

// GetConfigureText ports getConfigureText().
func (r *PortugueseReadabilityRule) GetConfigureText() string {
	return "Nível de legibilidade 0 (muito difícil) a 6 (muito fácil):"
}

// GetRuleOptions ports getRuleOptions() — uses PT configure text (Java polymorphic getConfigureText).
func (r *PortugueseReadabilityRule) GetRuleOptions() []*rules.RuleOption {
	return []*rules.RuleOption{rules.NewRuleOption(3, r.GetConfigureText(), 0, 6)}
}

// PrintMessageLevel ports printMessageLevel(level) — Portuguese "Nível" labels.
func (r *PortugueseReadabilityRule) PrintMessageLevel(level int) string {
	var sLevel string
	switch level {
	case 0:
		sLevel = "Muito complexo"
	case 1:
		sLevel = "Complexo"
	case 2:
		sLevel = "Moderadamente complexo"
	case 3:
		sLevel = "Meio-termo"
	case 4:
		sLevel = "Moderadamente simples"
	case 5:
		sLevel = "Simples"
	case 6:
		sLevel = "Muito simples"
	default:
		return ""
	}
	return " {Nível " + strconv.Itoa(level) + ": " + sLevel + "}"
}

// GetMessage ports getMessage(level, fre, asl, asw).
// Java includes FRE/ASL/ASW numbers in the template (unlike English/DE).
func (r *PortugueseReadabilityRule) GetMessage(level, fre, asl, asw int) string {
	simple, few := "difícil", "muitas"
	if r != nil && r.ReadabilityRule != nil && r.TooEasyTest {
		simple, few = "fácil", "poucas"
	}
	return "Legibilidade {FRE: " + strconv.Itoa(fre) + ", ASL: " + strconv.Itoa(asl) + ", ASW: " + strconv.Itoa(asw) +
		"}: O texto deste parágrafo é " + simple + r.PrintMessageLevel(level) + ". Tem " +
		few + " palavras por frase e " + few + " sílabas por palavra."
}

// GetFleschReadingEase ports getFleschReadingEase — Portuguese FRE:
// 206.84 - (1.02 * asl) - (60.0 * asw).
func (r *PortugueseReadabilityRule) GetFleschReadingEase(asl, asw float64) float64 {
	return 206.84 - (1.02 * asl) - (60.0 * asw)
}

// SimpleSyllablesCount ports protected simpleSyllablesCount(String word):
// UTF-16 charAt/length; PT isVowel; diphthong lastDouble rules; nSyllables==0 → 1.
func (r *PortugueseReadabilityRule) SimpleSyllablesCount(word string) int {
	return simpleSyllablesCountPT(word)
}

// isVowelPT ports private static isVowel(char c) — exact Java char set.
// Note: includes uppercase Ü only (not ü); does not include ã/õ (Java quirk:
// diphthong branches for ã/õ are unreachable because isVowel rejects them).
func isVowelPT(c rune) bool {
	return c == 'a' || c == 'e' || c == 'i' || c == 'o' || c == 'u' || c == 'y' ||
		c == 'A' || c == 'E' || c == 'I' || c == 'O' || c == 'U' || c == 'Y' ||
		c == 'á' || c == 'é' || c == 'í' || c == 'ó' || c == 'ú' || c == 'à' ||
		c == 'Á' || c == 'É' || c == 'Í' || c == 'Ó' || c == 'Ú' || c == 'À' ||
		c == 'â' || c == 'ê' || c == 'ô' ||
		c == 'Â' || c == 'Ê' || c == 'Ô' ||
		c == 'Ü'
}

// simpleSyllablesCountPT is the Java PortugueseReadabilityRule.simpleSyllablesCount body.
func simpleSyllablesCountPT(word string) int {
	n := utf16LenPT(word)
	if n == 0 {
		return 0
	}
	nSyllables := 0
	if isVowelPT(javaCharAtPT(word, 0)) {
		nSyllables++
	}
	lastDouble := false
	for i := 1; i < n; i++ {
		c := javaCharAtPT(word, i)
		if isVowelPT(c) {
			cl := javaCharAtPT(word, i-1)
			if lastDouble {
				nSyllables++
				lastDouble = false
			} else if ((c == 'ã' || c == 'õ') && (cl == 'e' || cl == 'o')) ||
				(c == 'a' && (cl == 'e' || cl == 'i' || cl == 'í' || cl == 'o' || cl == 'u' || cl == 'ú')) ||
				(c == 'e' && (cl == 'e' || cl == 'i' || cl == 'í' || cl == 'o' || cl == 'a' || cl == 'u')) ||
				(c == 'i' && (cl == 'a' || cl == 'e' || cl == 'o' || cl == 'u' || cl == 'á' || cl == 'é')) ||
				(c == 'í' && (cl == 'a' || cl == 'e')) ||
				(c == 'o' && (cl == 'a' || cl == 'á' || cl == 'e' || cl == 'é' || cl == 'i' || cl == 'í' || cl == 'u')) ||
				(c == 'u' && (cl == 'a' || cl == 'á' || cl == 'e' || cl == 'é' || cl == 'i' || cl == 'o')) ||
				(c == 'ú' && (cl == 'a' || cl == 'e' || cl == 'o')) {
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

func utf16LenPT(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func javaCharAtPT(s string, i int) rune {
	u := utf16.Encode([]rune(s))
	if i < 0 || i >= len(u) {
		return 0
	}
	return rune(u[i])
}
