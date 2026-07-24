package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/wrongWordInContext.txt
var wrongWordInContextFS embed.FS

var (
	wwicOnce    sync.Once
	wwicEntries []rules.ContextWords
)

func loadWrongWordInContext() []rules.ContextWords {
	wwicOnce.Do(func() {
		f, err := wrongWordInContextFS.Open("data/wrongWordInContext.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		entries, err := rules.LoadWrongWordInContext(f)
		if err != nil {
			panic(err)
		}
		wwicEntries = entries
	})
	return wwicEntries
}

// ArabicWrongWordInContextRule ports org.languagetool.rules.ar.ArabicWrongWordInContextRule.
type ArabicWrongWordInContextRule struct {
	*rules.WrongWordInContextRule
}

func NewArabicWrongWordInContextRule(messages map[string]string) *ArabicWrongWordInContextRule {
	base := &rules.WrongWordInContextRule{
		Messages:           messages,
		ID:                 "ARABIC_WRONG_WORD_IN_CONTEXT",
		Description:        "كلمات شائعةمتشابهة (ظل/ضلَ, رؤيا/رؤية الخ.)",
		MessageString:      "احتمال كلمة متشابهة: هل تقصد <suggestion>$SUGGESTION</suggestion> بدلا من '$WRONGWORD'?",
		ShortMessageString: "احتمال كلمة متشابهة",
		LongMessageString:  "احتمال كلمة متشابهة: هل تقصد <suggestion>$SUGGESTION</suggestion> (= $EXPLANATION_SUGGESTION) بدلا من '$WRONGWORD' (= $EXPLANATION_WRONGWORD)?",
		LanguageCode:       "ar",
		Entries:            loadWrongWordInContext(),
	}
	// Java: getCategoryString → "كلمات متشابهة"
	rules.InitWrongWordInContextMeta(base, messages, "كلمات متشابهة")
	// Java: الضن → الظن
	base.AddExamplePair(
		rules.Wrong("من سوء <marker>الضن</marker> بالله ترك الأمر بالمعروف."),
		rules.Fixed("من سوء <marker>الظن</marker> بالله ترك الأمر بالمعروف."),
	)
	return &ArabicWrongWordInContextRule{WrongWordInContextRule: base}
}

func (r *ArabicWrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WrongWordInContextRule.Match(sentence)
}
