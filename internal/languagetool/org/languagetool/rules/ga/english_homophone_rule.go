package ga

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/english-homophones.txt
var englishHomophonesFS embed.FS

var (
	englishHomophonesOnce sync.Once
	englishHomophonesMap  map[string][]string
)

func loadEnglishHomophones() map[string][]string {
	englishHomophonesOnce.Do(func() {
		f, err := englishHomophonesFS.Open("data/english-homophones.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		englishHomophonesMap = m
	})
	return englishHomophonesMap
}

// EnglishHomophoneRule ports org.languagetool.rules.ga.EnglishHomophoneRule.
type EnglishHomophoneRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewEnglishHomophoneRule(messages map[string]string) *EnglishHomophoneRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadEnglishHomophones(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "GA_ENGLISH_HOMOPHONE",
		Description:   "Is homafóin iad na focail, m.sh., \"well\" agus \"bhuel\"",
		ShortMsg:      "Homafón Béarla.",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Is é " + joinCommaGA(replacements) + " an litriú Gaelach ar \"" + tokenStr + "\""
		},
	}
	// Java: sushi → súisí
	base.AddExamplePair(
		rules.Wrong("An bhialann <marker>sushi</marker> sin ba chúis leis."),
		rules.Fixed("An bhialann <marker>súisí</marker> sin ba chúis leis."),
	)
	return &EnglishHomophoneRule{AbstractSimpleReplaceRule: base}
}

func (r *EnglishHomophoneRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
