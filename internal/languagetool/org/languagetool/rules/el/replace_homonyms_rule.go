package el

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var replaceFS embed.FS

// GreekHomonymsReplaceID ports ReplaceHomonymsRule.getId() → "GREEK_HOMONYMS_REPLACE".
const GreekHomonymsReplaceID = "GREEK_HOMONYMS_REPLACE"

// GreekHomonymsFilePath is the official resource path Java loads
// (ReplaceHomonymsRule.getFileNames() → "/el/replace.txt").
const GreekHomonymsFilePath = "/el/replace.txt"

var (
	homonymOnce sync.Once
	homonymBase *rules.AbstractSimpleReplaceRule2
)

// loadHomonyms loads official /el/replace.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadHomonyms() *rules.AbstractSimpleReplaceRule2 {
	homonymOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of ReplaceHomonymsRule + AbstractSimpleReplaceRule2 defaults:
		// - id GREEK_HOMONYMS_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator default ", " (Java getSuggestionsSeparator not overridden)
		// - CaseSensitivy.CI (base default)
		// - Locale EL via LanguageCode (getLocale → new Locale("EL"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.MISC (AbstractSimpleReplaceRule2 ctor), ITSIssueType.Uncategorized (Rule default)
		// - no addExamplePair in Java
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          GreekHomonymsReplaceID,
			Description: "Έλεγχος για λανθασμένη χρήση ομόηχων λέξεων σε μια πρόταση",
			ShortMsg:    "Λανθασμένη χρήση της λέξης",
			// Java getMessage(): "Μήπως εννοούσατε $suggestions?"
			MessageTemplate: "Μήπως εννοούσατε $suggestions?",
			// SuggestionsSeparator left empty → Match uses Java default ", "
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "EL", // Java: new Locale("EL")
			SubRuleSpecificIDs: false,
			Category:           rules.CatMisc.GetCategory(nil),
			// IssueType left empty → GetLocQualityIssueType returns ITSUncategorized
		}
		if err := base.LoadSimpleReplaceRule2Data(f, GreekHomonymsFilePath); err != nil {
			panic(err)
		}
		homonymBase = base
	})
	return homonymBase
}

// ReplaceHomonymsRule ports org.languagetool.rules.el.ReplaceHomonymsRule
// (extends AbstractSimpleReplaceRule2).
type ReplaceHomonymsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewReplaceHomonymsRule ports ReplaceHomonymsRule(ResourceBundle messages, Language language).
// Official data is go:embed of the same replace.txt Java loads from the classpath.
// language is represented via LanguageCode "EL" on the base (getLocale → Locale("EL")).
func NewReplaceHomonymsRule(messages map[string]string) *ReplaceHomonymsRule {
	base := loadHomonyms()
	r := *base
	r.Messages = messages
	// Java AbstractSimpleReplaceRule2 ctor: setCategory(Categories.MISC.getCategory(messages))
	r.Category = rules.CatMisc.GetCategory(messages)
	return &ReplaceHomonymsRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports ReplaceHomonymsRule.getFileNames() → singleton "/el/replace.txt".
func (r *ReplaceHomonymsRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{GreekHomonymsFilePath}
}

// GetShort ports ReplaceHomonymsRule.getShort().
func (r *ReplaceHomonymsRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports ReplaceHomonymsRule.getMessage().
func (r *ReplaceHomonymsRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports AbstractSimpleReplaceRule2.getSuggestionsSeparator()
// (ReplaceHomonymsRule does not override) → ", ".
func (r *ReplaceHomonymsRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports ReplaceHomonymsRule.getLocale() → Locale("EL").
func (r *ReplaceHomonymsRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replace.txt).
func (r *ReplaceHomonymsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
