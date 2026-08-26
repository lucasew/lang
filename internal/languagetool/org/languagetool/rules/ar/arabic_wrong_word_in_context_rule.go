package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicWrongWordInContextRuleID ports ArabicWrongWordInContextRule.ARABIC_WRONG_WORD_IN_CONTEXT.
const ArabicWrongWordInContextRuleID = "ARABIC_WRONG_WORD_IN_CONTEXT"

// ArabicWrongWordInContextFilePath ports ArabicWrongWordInContextRule.FILE_NAME / getFilename()
// → "/ar/wrongWordInContext.txt" (leading slash — Java rules-dir resource path).
// Go embeds the same official resource under data/ (path-law twin). Upstream file is
// header comments only; LoadWrongWordInContext yields an empty entry list.
const ArabicWrongWordInContextFilePath = "/ar/wrongWordInContext.txt"

//go:embed data/wrongWordInContext.txt
var wrongWordInContextFS embed.FS

var (
	wwicOnce    sync.Once
	wwicEntries []rules.ContextWords
)

// loadWrongWordInContext ports WrongWordInContextRule cache load of getFilename()
// (Java: JLanguageTool.getDataBroker().getFromRulesDirAsStream("/ar/wrongWordInContext.txt")).
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
// Java ctor: ArabicWrongWordInContextRule(ResourceBundle messages, Language lang).
// Go wires LanguageCode "ar" like other WWIC language ports (lang used for SpecificIdRule ids).
type ArabicWrongWordInContextRule struct {
	*rules.WrongWordInContextRule
}

// NewArabicWrongWordInContextRule ports ArabicWrongWordInContextRule(messages, lang).
// messages is ResourceBundle; Language is represented by LanguageCode "ar" on the base.
func NewArabicWrongWordInContextRule(messages map[string]string) *ArabicWrongWordInContextRule {
	base := &rules.WrongWordInContextRule{
		Messages: messages,
		ID:       ArabicWrongWordInContextRuleID,
		// Java getDescription: no space after شائعة (upstream string exact).
		Description:        "كلمات شائعةمتشابهة (ظل/ضلَ, رؤيا/رؤية الخ.)",
		MessageString:      "احتمال كلمة متشابهة: هل تقصد <suggestion>$SUGGESTION</suggestion> بدلا من '$WRONGWORD'?",
		ShortMessageString: "احتمال كلمة متشابهة",
		LongMessageString:  "احتمال كلمة متشابهة: هل تقصد <suggestion>$SUGGESTION</suggestion> (= $EXPLANATION_SUGGESTION) بدلا من '$WRONGWORD' (= $EXPLANATION_WRONGWORD)?",
		LanguageCode:       "ar",
		Entries:            loadWrongWordInContext(),
	}
	// Java super ctor: Category(CONFUSED_WORDS, getCategoryString()) + ITSIssueType.Misspelling
	// getCategoryString → "كلمات متشابهة"
	rules.InitWrongWordInContextMeta(base, messages, "كلمات متشابهة")
	// Java ctor: addExamplePair(الضن → الظن) — demo only; data file has no pairs yet.
	base.AddExamplePair(
		rules.Wrong("من سوء <marker>الضن</marker> بالله ترك الأمر بالمعروف."),
		rules.Fixed("من سوء <marker>الظن</marker> بالله ترك الأمر بالمعروف."),
	)
	return &ArabicWrongWordInContextRule{WrongWordInContextRule: base}
}

func (r *ArabicWrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WrongWordInContextRule.Match(sentence)
}
