package el

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/specific_case.txt
var specificCaseFS embed.FS

var (
	specificCaseOnce sync.Once
	specificCaseMap  map[string]string
	specificCaseMax  int
)

func loadSpecificCase() (map[string]string, int) {
	specificCaseOnce.Do(func() {
		f, err := specificCaseFS.Open("data/specific_case.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, maxLen, err := rules.LoadSpecificCasePhrases(f)
		if err != nil {
			panic(err)
		}
		specificCaseMap = m
		specificCaseMax = maxLen
	})
	return specificCaseMap, specificCaseMax
}

// GreekSpecificCaseRule ports org.languagetool.rules.el.GreekSpecificCaseRule.
type GreekSpecificCaseRule struct {
	*rules.AbstractSpecificCaseRule
}

func NewGreekSpecificCaseRule(messages map[string]string) *GreekSpecificCaseRule {
	m, maxLen := loadSpecificCase()
	base := &rules.AbstractSpecificCaseRule{
		Messages:                   messages,
		LcToProper:                 m,
		MaxPhraseLen:               maxLen,
		ID:                         "EL_SPECIFIC_CASE",
		Description:                "Ελέγχει αν κάποιες λέξεις χρειάζονται κεφαλαίο το πρώτο τους γράμμα",
		InitialCapitalMessage:      "Οι λέξεις της συγκεκριμένης έκφρασης χρείαζεται να ξεκινούν με κεφαλαία γράμματα.",
		OtherCapitalizationMessage: "Η συγκεκριμένη έκφραση γράφεται σύμφωνα με την προτεινόμενη κεφαλαιοποίηση.",
		ShortMsg:                   "Ειδική κεφαλαιοποίηση",
	}
	// Java: Ηνωμένες πολιτείες → Ηνωμένες Πολιτείες
	base.AddExamplePair(
		rules.Wrong("Κατοικώ στις <marker>Ηνωμένες πολιτείες</marker>."),
		rules.Fixed("Κατοικώ στις <marker>Ηνωμένες Πολιτείες</marker>."),
	)
	return &GreekSpecificCaseRule{AbstractSpecificCaseRule: base}
}

func (r *GreekSpecificCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpecificCaseRule.Match(sentence)
}
