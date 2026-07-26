package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official weasel list Java loads via PortugueseWeaselWordsRule:
//   getFileNames() → singleton "/pt/weaselwords.txt" (fixed; no locale path variants).
//
//go:embed data/weaselwords.txt
var weaselFS embed.FS

// PortugueseWeaselWordReplaceID ports PortugueseWeaselWordsRule.PT_WEASELWORD_REPLACE.
const PortugueseWeaselWordReplaceID = "PT_WEASELWORD_REPLACE"

// PortugueseWeaselWordsFilePath ports PortugueseWeaselWordsRule.FILE_NAME → "/pt/weaselwords.txt".
const PortugueseWeaselWordsFilePath = "/pt/weaselwords.txt"

// PortugueseWeaselWordsURL ports PortugueseWeaselWordsRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Weasel_word").
const PortugueseWeaselWordsURL = "https://pt.wikipedia.org/wiki/Weasel_word"

var (
	weaselOnce sync.Once
	weaselBase *rules.AbstractSimpleReplaceRule2
)

// loadWeasel loads official weaselwords.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadWeasel() *rules.AbstractSimpleReplaceRule2 {
	weaselOnce.Do(func() {
		f, err := weaselFS.Open("data/weaselwords.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of PortugueseWeaselWordsRule + AbstractSimpleReplaceRule2 defaults:
		// - id PT_WEASELWORD_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " ou "
		// - CaseSensitivy.CI (base default)
		// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
		// - does NOT call useSubRuleSpecificIds() (Java leaves default false)
		// - Categories.STYLE, ITSIssueType.Style
		// - URL pt.wikipedia.org Weasel_word
		// - example Diz-se → XYZ (rhetorical demo; fixed sentence longer)
		// Java getMessage() has NO $suggestions placeholder
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   PortugueseWeaselWordReplaceID,
			Description:          "Escrita avançada: Expressões evasivas",
			ShortMsg:             "Expressão evasiva",
			MessageTemplate:      "'$match' é uma expressão ambígua e evasiva. Reconsidere o seu uso, de acordo com o objetivo do seu texto.",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   false, // Java never calls useSubRuleSpecificIds()
			Category:             rules.CatStyle.GetCategory(nil),
			IssueType:            rules.ITSStyle,
			URL:                  PortugueseWeaselWordsURL,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, PortugueseWeaselWordsFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>Diz-se</marker> que programas gratuitos não têm qualidade."),
		//                Example.fixed("<marker>XYZ</marker> dizem que programas gratuitos não têm qualidade. Por isso vendem programas pagos."));
		base.AddExamplePair(
			rules.Wrong("<marker>Diz-se</marker> que programas gratuitos não têm qualidade."),
			rules.Fixed("<marker>XYZ</marker> dizem que programas gratuitos não têm qualidade. Por isso vendem programas pagos."),
		)
		weaselBase = base
	})
	return weaselBase
}

// PortugueseWeaselWordsRule ports org.languagetool.rules.pt.PortugueseWeaselWordsRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /pt/weaselwords.txt.
// Suggestions in the file are often rhetorical prompts rather than drop-in replacements.
type PortugueseWeaselWordsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewPortugueseWeaselWordsRule ports PortugueseWeaselWordsRule(ResourceBundle messages).
// Java: super(messages, Portuguese.getInstance()); setCategory(STYLE); setLocQualityIssueType(Style).
// language is represented via LanguageCode "pt" on the base (getLocale → Locale("pt")).
func NewPortugueseWeaselWordsRule(messages map[string]string) *PortugueseWeaselWordsRule {
	base := loadWeasel()
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	r.URL = PortugueseWeaselWordsURL
	return &PortugueseWeaselWordsRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports PortugueseWeaselWordsRule.getFileNames()
// → Collections.singletonList("/pt/weaselwords.txt").
func (r *PortugueseWeaselWordsRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{PortugueseWeaselWordsFilePath}
}

// GetShort ports PortugueseWeaselWordsRule.getShort().
func (r *PortugueseWeaselWordsRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseWeaselWordsRule.getMessage().
func (r *PortugueseWeaselWordsRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports PortugueseWeaselWordsRule.getSuggestionsSeparator() → " ou ".
func (r *PortugueseWeaselWordsRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports PortugueseWeaselWordsRule.getLocale() → Locale("pt").
func (r *PortugueseWeaselWordsRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from weaselwords.txt).
func (r *PortugueseWeaselWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
