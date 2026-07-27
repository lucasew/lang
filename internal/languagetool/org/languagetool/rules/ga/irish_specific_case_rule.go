package ga

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	gatag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
)

//go:embed data/specific_case.txt
var specificCaseFS embed.FS

// IrishSpecificCaseRuleID ports IrishSpecificCaseRule.getId() → "GA_SPECIFIC_CASE".
const IrishSpecificCaseRuleID = "GA_SPECIFIC_CASE"

// IrishSpecificCaseFilePath is the official resource-dir path Java loads
// (getFromResourceDirAsLines("/ga/specific_case.txt")).
const IrishSpecificCaseFilePath = "/ga/specific_case.txt"

var (
	specificCaseOnce sync.Once
	specificCaseMap  map[string]string
	specificCaseMax  int
)

// loadSpecificCase ports static phrases/lcToProperSpelling from /ga/specific_case.txt.
func loadSpecificCase() (map[string]string, int) {
	specificCaseOnce.Do(func() {
		f, err := specificCaseFS.Open("data/specific_case.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Java: JLanguageTool.getDataBroker().getFromResourceDirAsLines("/ga/specific_case.txt")
		m, maxLen, err := rules.LoadSpecificCasePhrases(f)
		if err != nil {
			panic(err)
		}
		specificCaseMap = m
		specificCaseMax = maxLen
	})
	return specificCaseMap, specificCaseMax
}

// IrishSpecificCaseRule ports org.languagetool.rules.ga.IrishSpecificCaseRule.
// Java extends Rule directly (not AbstractSpecificCaseRule); match algorithm is the
// specific-case phrase scan with ga.Utils case helpers. Shared AbstractSpecificCaseRule
// base carries metadata / example helpers; Match is the Irish Java twin (Utils +
// 5-arg RuleMatch with empty short).
type IrishSpecificCaseRule struct {
	*rules.AbstractSpecificCaseRule
}

// NewIrishSpecificCaseRule ports IrishSpecificCaseRule(ResourceBundle messages).
func NewIrishSpecificCaseRule(messages map[string]string) *IrishSpecificCaseRule {
	m, maxLen := loadSpecificCase()
	base := &rules.AbstractSpecificCaseRule{
		// Java: phrases from /ga/specific_case.txt
		LcToProper:   m,
		MaxPhraseLen: maxLen,
		// Java IrishSpecificCaseRule.getId / getDescription
		ID:          IrishSpecificCaseRuleID,
		Description: "Checks upper/lower case spelling of some proper nouns",
		// Java match: allWordsUppercase(properSpelling) branch (Utils.startsWithUppercase)
		InitialCapitalMessage: "Más ainmfhocal dílis é, scríobh é i gceannlitreacha.",
		// Java match: else branch
		OtherCapitalizationMessage: "If the term is a proper noun, use the suggested capitalization.",
		// Java: new RuleMatch(this, sentence, from, to, msg) — no short message argument
		// and no getShortMessage override. Do not invent Abstract's "Special capitalization".
		ShortMsg: "",
	}
	// Java: super.setCategory(Categories.CASING.getCategory(messages)); setLocQualityIssueType(Misspelling)
	rules.InitSpecificCaseMeta(base, messages)
	// Java: mbéal Feirste → mBéal Feirste
	base.AddExamplePair(
		rules.Wrong("Rugadh agus tógadh i <marker>mbéal Feirste</marker> é."),
		rules.Fixed("Rugadh agus tógadh i <marker>mBéal Feirste</marker> é."),
	)
	return &IrishSpecificCaseRule{AbstractSpecificCaseRule: base}
}

// Match ports IrishSpecificCaseRule.match (Utils case checks; 5-arg RuleMatch → empty short).
// End position is start + phrase.length() (Java UTF-16), not last-token GetEndPos.
func (r *IrishSpecificCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 0; i < len(tokens); i++ {
		var l []string
		j := 0
		for len(l) < r.MaxPhraseLen && i+j < len(tokens) {
			l = append(l, tokens[i+j].GetToken())
			j++
			phrase := strings.Join(l, " ")
			lcPhrase := strings.ToLower(phrase)
			properSpelling, ok := r.LcToProper[lcPhrase]
			// Java: !Utils.isAllUppercase(phrase) && !phrase.equals(properSpelling)
			if !ok || gatag.IsAllUppercase(phrase) || phrase == properSpelling {
				continue
			}
			// Java: avoid suggesting e.g. "vitamin C" at sentence start (Utils.startsWithUppercase)
			if i > 0 && tokens[i-1].IsSentenceStart() && !gatag.StartsWithUppercase(properSpelling) {
				continue
			}
			var msg string
			if irishAllWordsUppercase(properSpelling) {
				msg = r.InitialCapitalMessage
			} else {
				msg = r.OtherCapitalizationMessage
			}
			from := tokens[i].GetStartPos()
			// Java: tokens[i].getStartPos() + phrase.length()  (UTF-16 code units)
			to := from + tagging.UTF16Len(phrase)
			// Java 5-arg RuleMatch — short message null / empty
			rm := rules.NewRuleMatch(r, sentence, from, to, msg)
			rm.ShortMessage = r.ShortMsg
			rm.SetSuggestedReplacement(properSpelling)
			matches = append(matches, rm)
		}
	}
	return matches
}

// irishAllWordsUppercase ports IrishSpecificCaseRule.allWordsUppercase (Utils.startsWithUppercase).
func irishAllWordsUppercase(s string) bool {
	for _, w := range strings.Split(s, " ") {
		if !gatag.StartsWithUppercase(w) {
			return false
		}
	}
	return true
}
