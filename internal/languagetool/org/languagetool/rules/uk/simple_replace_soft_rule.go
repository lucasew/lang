package uk

import (
	"embed"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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
		// lower-case keys for case-insensitive lookup
		out := make(map[string][]string, len(m))
		for k, v := range m {
			out[strings.ToLower(k)] = v
		}
		softWords = out
	})
	return softWords
}

// SimpleReplaceSoftRule ports org.languagetool.rules.uk.SimpleReplaceSoftRule
// (surface match; no POS ignoreTaggedWords without a tagger).
type SimpleReplaceSoftRule struct {
	messages map[string]string
}

func NewSimpleReplaceSoftRule(messages map[string]string) *SimpleReplaceSoftRule {
	_ = loadSoftWords()
	return &SimpleReplaceSoftRule{messages: messages}
}

func (r *SimpleReplaceSoftRule) GetID() string { return "UK_SIMPLE_REPLACE_SOFT" }

func (r *SimpleReplaceSoftRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	words := loadSoftWords()
	var out []*rules.RuleMatch
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if tok.IsSentenceStart() || tok.IsImmunized() {
			continue
		}
		t := tok.GetToken()
		if strings.EqualFold(t, "завидна") {
			continue
		}
		reps, ok := words[strings.ToLower(t)]
		if !ok || len(reps) == 0 {
			continue
		}
		ctxs, suggestions := splitSoftReplacements(reps)
		if len(suggestions) == 0 {
			continue
		}
		// case-adjust suggestions
		final := make([]string, 0, len(suggestions))
		for _, s := range suggestions {
			if tools.IsAllUppercase(t) {
				final = append(final, strings.ToUpper(s))
			} else if tools.StartsWithUppercase(t) {
				final = append(final, tools.UppercaseFirstChar(s))
			} else {
				final = append(final, s)
			}
		}
		msg := "«" + t + "» — нерекомендоване слово, кращий варіант: " + strings.Join(final, ", ") + "."
		if len(ctxs) > 0 {
			msg = "«" + t + "» вживається лише в таких контекстах: " + strings.Join(ctxs, ", ") +
				", можливо, ви мали на увазі: " + strings.Join(final, ", ") + "?"
		}
		from := tok.GetStartPos()
		to := from + utf16Len(t)
		rm := rules.NewRuleMatch(r, sentence, from, to, msg)
		rm.ShortMessage = "Нерекомендоване слово"
		rm.SetSuggestedReplacements(final)
		out = append(out, rm)
	}
	return out
}

func splitSoftReplacements(reps []string) (contexts, suggestions []string) {
	for _, rep := range reps {
		if strings.HasPrefix(rep, "ctx:") {
			rest := strings.TrimSpace(strings.TrimPrefix(rep, "ctx:"))
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

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
