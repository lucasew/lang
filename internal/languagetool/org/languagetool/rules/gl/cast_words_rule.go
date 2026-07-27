package gl

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/spanish.txt
var spanishFS embed.FS

// CastWordsRuleID ports CastWordsRule.GL_CAST_WORDS_RULE → "GL_CAST_WORDS".
const CastWordsRuleID = "GL_CAST_WORDS"

// CastWordsFilePath is the official classpath resource Java loads
// (static loadFromPath("/gl/spanish.txt")).
const CastWordsFilePath = "/gl/spanish.txt"

var (
	spanishOnce sync.Once
	spanishMap  map[string][]string
)

// loadSpanish ports static wrongWords = loadFromPath("/gl/spanish.txt").
// Official resource is go:embed of the same spanish.txt Java loads (byte-identical).
func loadSpanish() map[string][]string {
	spanishOnce.Do(func() {
		f, err := spanishFS.Open("data/spanish.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		spanishMap = m
	})
	return spanishMap
}

// CastWordsRule ports org.languagetool.rules.gl.CastWordsRule
// (extends AbstractSimpleReplaceRule, not AbstractSimpleReplaceRule2).
//
// Java field twin summary:
//   - id GL_CAST_WORDS
//   - static wrongWords from /gl/spanish.txt
//   - Categories.MISC via AbstractSimpleReplaceRule super (no setCategory override)
//   - no setLocQualityIssueType → base ITSIssueType default (Uncategorized)
//   - isCaseSensitive() false
//   - isCheckLemmas default true (ASR field; CastWords does not call setCheckLemmas(false))
//   - getLocale() → Locale("gl")
//   - description / short / message hardcoded Galician strings
//   - no addExamplePair, no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl, no setDefaultOff
type CastWordsRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewCastWordsRule ports CastWordsRule(ResourceBundle messages, Language language).
// language is represented via LanguageCode "gl" (getLocale → Locale("gl")).
func NewCastWordsRule(messages map[string]string) *CastWordsRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadSpanish(),
		CaseSensitive: false, // Java isCaseSensitive() false
		CheckLemmas:   true,  // Java AbstractSimpleReplaceRule default true
		ID:            CastWordsRuleID,
		LanguageCode:  "gl", // Java: new Locale("gl")
		Description:   "Corrección de erros léxicos (castelanismos).",
		ShortMsg:      "Castelanismos léxicos",
		// Java AbstractSimpleReplaceRule ctor: Categories.MISC.getCategory(messages)
		Category: rules.CatMisc.GetCategory(messages),
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java: "'" + tokenStr + "' é un castelanismo. Empregue no seu sitio: "
			//       + String.join(", ", replacements) + "."
			return "'" + tokenStr + "' é un castelanismo. Empregue no seu sitio: " +
				strings.Join(replacements, ", ") + "."
		},
	}
	return &CastWordsRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports CastWordsRule.getShort() → "Castelanismos léxicos".
func (r *CastWordsRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports CastWordsRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real suggestions.
func (r *CastWordsRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports CastWordsRule.getLocale() → Locale("gl").
func (r *CastWordsRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports CastWordsRule.getWrongWords() → static map.
func (r *CastWordsRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// IsDefaultOff ports Rule.isDefaultOff — Java never calls setDefaultOff.
func (r *CastWordsRule) IsDefaultOff() bool {
	return false
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *CastWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
