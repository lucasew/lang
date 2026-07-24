package el

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var replaceFS embed.FS

var (
	homonymOnce sync.Once
	homonymBase *rules.AbstractSimpleReplaceRule2
)

func loadHomonyms() *rules.AbstractSimpleReplaceRule2 {
	homonymOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "GREEK_HOMONYMS_REPLACE",
			Description:          "Έλεγχος για λανθασμένη χρήση ομόηχων λέξεων σε μια πρόταση",
			ShortMsg:             "Λανθασμένη χρήση της λέξης",
			MessageTemplate:      "Μήπως εννοούσατε $suggestions?",
			SuggestionsSeparator: " ή ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "el",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/el/replace.txt"); err != nil {
			panic(err)
		}
		homonymBase = base
	})
	return homonymBase
}

// ReplaceHomonymsRule ports org.languagetool.rules.el.ReplaceHomonymsRule.
type ReplaceHomonymsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewReplaceHomonymsRule(messages map[string]string) *ReplaceHomonymsRule {
	base := loadHomonyms()
	r := *base
	r.Messages = messages
	return &ReplaceHomonymsRule{AbstractSimpleReplaceRule2: &r}
}

func (r *ReplaceHomonymsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
