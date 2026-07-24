package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/en-NZ/replace.txt
var nzReplaceFS embed.FS

var (
	nzReplaceOnce sync.Once
	nzReplaceBase *rules.AbstractSimpleReplaceRule2
)

func loadNZReplace() *rules.AbstractSimpleReplaceRule2 {
	nzReplaceOnce.Do(func() {
		f, err := nzReplaceFS.Open("data/en-NZ/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Java NewZealandReplaceRule: STYLE, LocaleViolation, sidewalk → footpath
		base := &rules.AbstractSimpleReplaceRule2{
			ID:              "EN_NZ_SIMPLE_REPLACE",
			Description:     "English words easily confused in New Zealand English",
			ShortMsg:        "Not a New Zealand English word",
			MessageTemplate: "'$match' is a non-standard expression. Consider using expressions more common to New Zealand English.",
			CaseSens:        rules.CaseInsensitive,
			LanguageCode:    "en-NZ",
			Category:        rules.CatStyle.GetCategory(nil),
			IssueType:       rules.ITSLocaleViolation,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/en/en-NZ/replace.txt"); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("A <marker>sidewalk</marker> is a path along the side of a road."),
			rules.Fixed("A <marker>footpath</marker> is a path along the side of a road."),
		)
		nzReplaceBase = base
	})
	return nzReplaceBase
}

// NewZealandReplaceRule ports org.languagetool.rules.en.NewZealandReplaceRule.
type NewZealandReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewNewZealandReplaceRule(messages map[string]string) *NewZealandReplaceRule {
	base := loadNZReplace()
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	return &NewZealandReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *NewZealandReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
