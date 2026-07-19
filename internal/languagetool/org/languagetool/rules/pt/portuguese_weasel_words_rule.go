package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/weaselwords.txt
var weaselFS embed.FS

var (
	weaselOnce sync.Once
	weaselBase *rules.AbstractSimpleReplaceRule2
)

func loadWeasel() *rules.AbstractSimpleReplaceRule2 {
	weaselOnce.Do(func() {
		f, err := weaselFS.Open("data/weaselwords.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "PT_WEASELWORD_REPLACE",
			Description:          "Escrita avançada: Expressões evasivas",
			ShortMsg:             "Expressão evasiva",
			MessageTemplate:      "'$match' é uma expressão ambígua e evasiva. Reconsidere o seu uso, de acordo com o objetivo do seu texto.",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/pt/weaselwords.txt"); err != nil {
			panic(err)
		}
		// Java: Diz-se → XYZ dizem …
		base.AddExamplePair(
			rules.Wrong("<marker>Diz-se</marker> que programas gratuitos não têm qualidade."),
			rules.Fixed("<marker>XYZ</marker> dizem que programas gratuitos não têm qualidade. Por isso vendem programas pagos."),
		)
		weaselBase = base
	})
	return weaselBase
}

// PortugueseWeaselWordsRule ports org.languagetool.rules.pt.PortugueseWeaselWordsRule.
// Suggestions are often rhetorical prompts rather than drop-in replacements.
type PortugueseWeaselWordsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewPortugueseWeaselWordsRule(messages map[string]string) *PortugueseWeaselWordsRule {
	base := loadWeasel()
	r := *base
	r.Messages = messages
	return &PortugueseWeaselWordsRule{AbstractSimpleReplaceRule2: &r}
}

func (r *PortugueseWeaselWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
