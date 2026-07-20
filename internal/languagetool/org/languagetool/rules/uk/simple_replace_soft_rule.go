package uk

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_soft.txt
var softFS embed.FS

var (
	softOnce  sync.Once
	softWords map[string][]string
)

func loadSoftWords() map[string][]string {
	softOnce.Do(func() {
		f, err := softFS.Open("data/replace_soft.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		// lower-case keys for case-insensitive lookup (Java isCaseSensitive false)
		out := make(map[string][]string, len(m))
		for k, v := range m {
			out[strings.ToLower(k)] = v
		}
		softWords = out
	})
	return softWords
}

// SimpleReplaceSoftRule ports org.languagetool.rules.uk.SimpleReplaceSoftRule
// (AbstractSimpleReplaceRule + ctx: contexts; Style ITS; no ignoreTaggedWords).
type SimpleReplaceSoftRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceSoftRule(messages map[string]string) *SimpleReplaceSoftRule {
	words := loadSoftWords()
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    words,
		CaseSensitive: false,
		// Java AbstractSimpleReplaceRule default checkLemmas true; Soft does not override.
		CheckLemmas: true,
		// Java Soft does not call setIgnoreTaggedWords.
		IgnoreTaggedWords: false,
		ID:                "UK_SIMPLE_REPLACE_SOFT",
		Description:       "Пошук нерекомендованих слів",
		ShortMsg:          "Нерекомендоване слово",
		IssueType:         rules.ITSStyle,
		TokenException:    softTokenException,
		MessageFn:         softMessage,
	}
	rules.InitSimpleReplaceMeta(base, messages)
	return &SimpleReplaceSoftRule{AbstractSimpleReplaceRule: base}
}

// softTokenException ports SimpleReplaceSoftRule.isTokenException (завидна + super).
// Java: "завидна".equals(atr.getCleanToken()) || super… (super always false).
func softTokenException(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	c := atr.GetCleanToken()
	if c == "" {
		c = atr.GetToken()
	}
	return c == "завидна"
}

// softMessage ports SimpleReplaceSoftRule.getMessage (ctx: + non-ctx suggestions).
func softMessage(tokenStr string, replacements []string) string {
	ctxs, suggestions := splitSoftReplacements(replacements)
	if len(suggestions) == 0 {
		return "«" + tokenStr + "» — нерекомендоване слово."
	}
	if len(ctxs) > 0 {
		return "«" + tokenStr + "» вживається лише в таких контекстах: " + strings.Join(ctxs, ", ") +
			", можливо, ви мали на увазі: " + strings.Join(suggestions, ", ") + "?"
	}
	return "«" + tokenStr + "» — нерекомендоване слово, кращий варіант: " + strings.Join(suggestions, ", ") + "."
}

// Match runs AbstractSimpleReplaceRule.match then retains only non-ctx suggestions
// (Java getMessage: replacements.retainAll(repl.replacements)).
func (r *SimpleReplaceSoftRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	matches := r.AbstractSimpleReplaceRule.Match(sentence)
	for _, m := range matches {
		if m == nil {
			continue
		}
		_, sugs := splitSoftReplacements(m.GetSuggestedReplacements())
		m.SetSuggestedReplacements(sugs)
	}
	return matches
}

func splitSoftReplacements(reps []string) (contexts, suggestions []string) {
	for _, rep := range reps {
		// Java CONTEXT_PREFIX = "ctx:"; startsWith is case-sensitive
		if strings.HasPrefix(rep, "ctx:") {
			rest := strings.TrimSpace(rep[len("ctx:"):])
			for _, c := range strings.Split(rest, ",") {
				c = strings.TrimSpace(c)
				if c != "" {
					contexts = append(contexts, c)
				}
			}
			continue
		}
		suggestions = append(suggestions, rep)
	}
	return contexts, suggestions
}
