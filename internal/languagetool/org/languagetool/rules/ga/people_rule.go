package ga

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/names.txt
var namesFS embed.FS

// PeopleRuleID ports PeopleRule.getId() → "GA_PEOPLE".
const PeopleRuleID = "GA_PEOPLE"

// PeopleFilePath is the official classpath resource Java loads
// (static loadFromPath("/ga/names.txt")).
const PeopleFilePath = "/ga/names.txt"

var (
	namesOnce sync.Once
	namesMap  map[string][]string
)

// loadNames ports static wrongWords = loadFromPath("/ga/names.txt").
// Official resource is go:embed of the same names.txt Java loads (byte-identical).
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

// PeopleRule ports org.languagetool.rules.ga.PeopleRule
// (extends AbstractSimpleReplaceRule; English personal names → Irish forms).
//
// Java field twin summary:
//   - id GA_PEOPLE
//   - static wrongWords from /ga/names.txt
//   - Categories.TYPOS (overrides ASR ctor MISC)
//   - ITSIssueType.Misspelling
//   - isCaseSensitive() false
//   - setCheckLemmas(false)
//   - getLocale() → Locale("GA")
//   - description / short / message hardcoded Irish strings
//   - example: Damocles → Dámaicléas
//   - no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl
type PeopleRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewPeopleRule ports PeopleRule(ResourceBundle messages, Language language).
// language is Irish; getLocale is Locale("GA") — stored as LanguageCode.
func NewPeopleRule(messages map[string]string) *PeopleRule {
	// Field twin of PeopleRule + AbstractSimpleReplaceRule:
	// - id GA_PEOPLE
	// - static wrongWords from /ga/names.txt
	// - isCaseSensitive() false
	// - setCheckLemmas(false)
	// - getLocale → Locale("GA") via LanguageCode
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - description / short / message hardcoded GA strings
	// - example pair Damocles → Dámaicléas
	// - no setIgnoreTaggedWords, no setUrl, no useSubRuleSpecificIds
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadNames(),
		CaseSensitive: false, // Java isCaseSensitive() false
		CheckLemmas:   false, // Java setCheckLemmas(false)
		ID:            PeopleRuleID,
		LanguageCode:  "GA", // Java: new Locale("GA")
		Description:   "Ainm Béarla, m.sh., 'Damocles' in áit 'Dámaicléas'.",
		ShortMsg:      "Ainm Béarla",
		Category:      rules.CatTypos.GetCategory(messages),
		IssueType:     rules.ITSMisspelling,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// return "Ainm Béarla. \""
			//     + String.join(", ", replacements) + "\".";
			// tokenStr is unused (matches Java getMessage signature).
			return "Ainm Béarla. \"" + strings.Join(replacements, ", ") + "\"."
		},
	}
	// Java: Damocles → Dámaicléas
	base.AddExamplePair(
		rules.Wrong("Bhí sí cosúil le claíomh <marker>Damocles</marker> ar crochadh sa spéir."),
		rules.Fixed("Bhí sí cosúil le claíomh <marker>Dámaicléas</marker> ar crochadh sa spéir."),
	)
	return &PeopleRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports PeopleRule.getShort() → "Ainm Béarla".
func (r *PeopleRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PeopleRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real suggestions.
func (r *PeopleRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports PeopleRule.getLocale() → Locale("GA").
func (r *PeopleRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports PeopleRule.getWrongWords() → static map.
func (r *PeopleRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *PeopleRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
