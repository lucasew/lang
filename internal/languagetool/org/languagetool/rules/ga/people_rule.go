package ga

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/names.txt
var namesFS embed.FS

var (
	namesOnce sync.Once
	namesMap  map[string][]string
)

func loadNames() map[string][]string {
	namesOnce.Do(func() {
		f, err := namesFS.Open("data/names.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		namesMap = m
	})
	return namesMap
}

// PeopleRule ports org.languagetool.rules.ga.PeopleRule (English personal names → Irish forms).
type PeopleRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewPeopleRule(messages map[string]string) *PeopleRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadNames(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "GA_PEOPLE",
		Description:   "Ainm Béarla, m.sh., 'Damocles' in áit 'Dámaicléas'.",
		ShortMsg:      "Ainm Béarla",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Ainm Béarla. \"" + joinCommaGA(replacements) + "\"."
		},
	}
	// Java: Damocles → Dámaicléas
	base.AddExamplePair(
		rules.Wrong("Bhí sí cosúil le claíomh <marker>Damocles</marker> ar crochadh sa spéir."),
		rules.Fixed("Bhí sí cosúil le claíomh <marker>Dámaicléas</marker> ar crochadh sa spéir."),
	)
	return &PeopleRule{AbstractSimpleReplaceRule: base}
}

func (r *PeopleRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
