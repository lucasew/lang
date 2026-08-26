package ga

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DativePluralStandardReplaceRuleID ports DativePluralStandardReplaceRule.getId()
// → "GA_DATIVE_PLURALS_STD".
const DativePluralStandardReplaceRuleID = "GA_DATIVE_PLURALS_STD"

var (
	dativeWrongOnce sync.Once
	dativeWrongMap  map[string][]string
)

// loadDativeWrongWords ports static:
//   replacements = DativePluralsData.getSimpleReplacements();
//   wrongWords = listify(replacements);
func loadDativeWrongWords() map[string][]string {
	dativeWrongOnce.Do(func() {
		dativeWrongMap = listifyDative(GetSimpleReplacements())
	})
	return dativeWrongMap
}

// DativePluralStandardReplaceRule ports org.languagetool.rules.ga.DativePluralStandardReplaceRule
// (extends AbstractSimpleReplaceRule).
//
// Java field twin summary:
//   - id GA_DATIVE_PLURALS_STD
//   - static wrongWords from listify(DativePluralsData.getSimpleReplacements())
//   - Categories.TYPOS (overrides ASR ctor MISC)
//   - ITSIssueType.Grammar (not Misspelling)
//   - isCaseSensitive() not overridden → ASR default true
//   - setCheckLemmas(false)
//   - GA_LOCALE field exists (Locale("GA")) but getLocale() is not overridden
//     (ASR default Locale.getDefault()); stored as LanguageCode for field twin
//   - description / short / message hardcoded Irish strings
//   - example: mnáibh → mná
//   - no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl
type DativePluralStandardReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewDativePluralStandardReplaceRule ports
// DativePluralStandardReplaceRule(ResourceBundle messages, Language language).
func NewDativePluralStandardReplaceRule(messages map[string]string) *DativePluralStandardReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadDativeWrongWords(),
		CaseSensitive: true,  // Java: isCaseSensitive not overridden → ASR default true
		CheckLemmas:   false, // Java setCheckLemmas(false)
		ID:            DativePluralStandardReplaceRuleID,
		LanguageCode:  "GA", // Java field GA_LOCALE = new Locale("GA") (unused by getLocale)
		Description:   "Tuiseal tabharthach iolra",
		ShortMsg:      "Tabharthach iolra",
		Category:      rules.CatTypos.GetCategory(messages),
		IssueType:     rules.ITSGrammar, // Java: ITSIssueType.Grammar
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// return "Is litriú réamhchaighdeánach (tabharthach iolra) é \"" + tokenStr
			//   + "\"; an \"" + String.join(", ", replacements) + "\" a bhí i gceist agat?";
			return "Is litriú réamhchaighdeánach (tabharthach iolra) é \"" + tokenStr + "\"; an \"" +
				strings.Join(replacements, ", ") + "\" a bhí i gceist agat?"
		},
	}
	// Java: mnáibh → mná
	base.AddExamplePair(
		rules.Wrong("Dá mba thruamhéalach é cás an sclábhaí fir, ba mheasa fós do na <marker>mnáibh</marker> a gcás siúd."),
		rules.Fixed("Dá mba thruamhéalach é cás an sclábhaí fir, ba mheasa fós do na <marker>mná</marker> a gcás siúd."),
	)
	return &DativePluralStandardReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports DativePluralStandardReplaceRule.getShort() → "Tabharthach iolra".
func (r *DativePluralStandardReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports getMessage with empty token/replacements for metadata probes.
// Match-time messages go through MessageFn with real suggestions.
func (r *DativePluralStandardReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode returns the Java GA_LOCALE field value (Locale("GA")).
// Note: Java does not override getLocale(); ASR default remains Locale.getDefault().
func (r *DativePluralStandardReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports getWrongWords() → static map.
func (r *DativePluralStandardReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *DativePluralStandardReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
