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

// loadSoftWords ports static WRONG_WORDS = loadFromPath("/uk/replace_soft.txt").
// Official resource is go:embed of the same replace_soft.txt Java loads (byte-identical).
// Keys kept as in file (Java SimpleReplaceDataLoader).
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
		softWords = m
	})
	return softWords
}

// SimpleReplaceSoftRule ports org.languagetool.rules.uk.SimpleReplaceSoftRule
// (AbstractSimpleReplaceRule + ctx: contexts; Style ITS; no ignoreTaggedWords).
type SimpleReplaceSoftRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewSimpleReplaceSoftRule ports SimpleReplaceSoftRule(ResourceBundle, Language).
func NewSimpleReplaceSoftRule(messages map[string]string) *SimpleReplaceSoftRule {
	words := loadSoftWords()
	base := &rules.AbstractSimpleReplaceRule{
		Messages:          messages,
		WrongWords:        words,
		CaseSensitive:     false, // Java isCaseSensitive() false
		CheckLemmas:       true,  // Java default; Soft does not override
		IgnoreTaggedWords: false, // Java Soft does not call setIgnoreTaggedWords
		ID:                "UK_SIMPLE_REPLACE_SOFT",
		Description:       "Пошук нерекомендованих слів",
		ShortMsg:          "Нерекомендоване слово",
		IssueType:         rules.ITSStyle,
		TokenException:    softTokenException,
		// MessageFn unused: Soft getMessage mutates replacements (retainAll) before
		// title-case — handled in createRuleMatchSoft (Java createRuleMatch order).
	}
	// Java AbstractSimpleReplaceRule ctor: Categories.MISC
	rules.InitSimpleReplaceMeta(base, messages)
	return &SimpleReplaceSoftRule{AbstractSimpleReplaceRule: base}
}

// softTokenException ports SimpleReplaceSoftRule.isTokenException:
// "завидна".equals(atr.getCleanToken()) || super.isTokenException(atr) (super always false).
func softTokenException(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	return atr.GetCleanToken() == "завидна"
}

// Match ports AbstractSimpleReplaceRule.match with Soft createRuleMatch (ctx retainAll).
func (r *SimpleReplaceSoftRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil || sentence == nil {
		return nil
	}
	var ruleMatches []*rules.RuleMatch
	for _, tokenReadings := range sentence.GetTokensWithoutWhitespace() {
		if tokenReadings == nil || tokenReadings.IsSentenceStart() || tokenReadings.IsImmunized() {
			continue
		}
		if r.TokenException != nil && r.TokenException(tokenReadings) {
			continue
		}
		if tokenReadings.IsIgnoredBySpeller() {
			continue
		}
		// Soft: ignoreTaggedWords is false — no tagged skip.
		// Dictionary lookup via base FindMatches path, then Soft createRuleMatch.
		// Base FindMatches already calls createRuleMatch with MessageFn — we reimplement
		// find to apply Soft getMessage retainAll before title-case.
		matches := r.findMatchesSoft(tokenReadings, sentence)
		ruleMatches = append(ruleMatches, matches...)
	}
	return ruleMatches
}

// findMatchesSoft ports AbstractSimpleReplaceRule.findMatches (same algorithm as base)
// then Soft createRuleMatchSoft for getMessage retainAll.
func (r *SimpleReplaceSoftRule) findMatchesSoft(
	tokenReadings *languagetool.AnalyzedTokenReadings,
	sentence *languagetool.AnalyzedSentence,
) []*rules.RuleMatch {
	originalTokenStr := tokenReadings.GetToken()
	tokenString := r.cleanupSoft(originalTokenStr)
	isAllUppercase := tools.IsAllUppercase(originalTokenStr)

	// try original word, then lower-case (Java getWrongWords)
	possibleReplacements := r.WrongWords[originalTokenStr]
	if possibleReplacements == nil {
		possibleReplacements = r.WrongWords[tokenString]
	}

	// Lemma path: same as AbstractSimpleReplaceRule.findMatches (cleanup lemma keys;
	// no synthesizer on Soft → add replacements directly; distinct).
	if possibleReplacements == nil && r.CheckLemmas {
		var found []string
		seen := map[string]bool{}
		for _, at := range tokenReadings.GetReadings() {
			if at.GetLemma() == nil {
				continue
			}
			lemma := r.cleanupSoft(*at.GetLemma())
			reps, ok := r.WrongWords[lemma]
			if !ok {
				continue
			}
			for _, rep := range reps {
				if !seen[rep] {
					seen[rep] = true
					found = append(found, rep)
				}
			}
		}
		if len(found) > 0 {
			possibleReplacements = found
		}
	}

	if len(possibleReplacements) == 0 {
		return nil
	}

	var replacements []string
	if isAllUppercase {
		for _, s := range possibleReplacements {
			replacements = append(replacements, strings.ToUpper(s))
		}
	} else {
		replacements = append([]string(nil), possibleReplacements...)
	}
	// Java: replacements.remove(originalTokenStr)
	filtered := replacements[:0]
	for _, rep := range replacements {
		if rep != originalTokenStr {
			filtered = append(filtered, rep)
		}
	}
	replacements = filtered
	if len(replacements) == 0 {
		return nil
	}
	return []*rules.RuleMatch{r.createRuleMatchSoft(tokenReadings, replacements, sentence)}
}

func (r *SimpleReplaceSoftRule) cleanupSoft(word string) string {
	if r != nil && r.CaseSensitive {
		return word
	}
	return strings.ToLower(word)
}

// createRuleMatchSoft ports AbstractSimpleReplaceRule.createRuleMatch + Soft getMessage.
// Order: getMessage (retainAll) → title-case remaining suggestions → setSuggestedReplacements.
func (r *SimpleReplaceSoftRule) createRuleMatchSoft(
	tokenReadings *languagetool.AnalyzedTokenReadings,
	replacements []string,
	sentence *languagetool.AnalyzedSentence,
) *rules.RuleMatch {
	tokenString := tokenReadings.GetToken()
	pos := tokenReadings.GetStartPos()
	end := pos + softUTF16Len(tokenString)

	msg, replacements := softMessageRetainAll(tokenString, replacements)

	short := r.ShortMsg
	if short == "" {
		short = "Нерекомендоване слово"
	}
	rm := rules.NewRuleMatch(r, sentence, pos, end, msg)
	rm.ShortMessage = short

	if !r.CaseSensitive && tools.StartsWithUppercase(tokenString) {
		for i, rep := range replacements {
			replacements[i] = tools.UppercaseFirstChar(rep)
		}
	}
	rm.SetSuggestedReplacements(replacements)
	return rm
}

// softMessageRetainAll ports SimpleReplaceSoftRule.getMessage + findContext + retainAll.
func softMessageRetainAll(tokenStr string, replacements []string) (string, []string) {
	ctxs, suggestions := splitSoftReplacements(replacements)
	// Java: replacements.retainAll(repl.replacements)
	replaceText := strings.Join(suggestions, ", ")
	if len(ctxs) > 0 {
		msg := "«" + tokenStr + "» вживається лише в таких контекстах: " +
			strings.Join(ctxs, ", ") +
			", можливо, ви мали на увазі: " + replaceText + "?"
		return msg, suggestions
	}
	msg := "«" + tokenStr + "» — нерекомендоване слово, кращий варіант: " + replaceText + "."
	return msg, suggestions
}

// splitSoftReplacements ports ContextRepl.findContext (CONTEXT_PREFIX = "ctx:").
func splitSoftReplacements(reps []string) (contexts, suggestions []string) {
	for _, rep := range reps {
		// Java: replacement.startsWith("ctx:") — case-sensitive
		if strings.HasPrefix(rep, "ctx:") {
			// Java: replacement.replace(CONTEXT_PREFIX, "").trim().split(", *")
			rest := tools.JavaStringTrim(strings.ReplaceAll(rep, "ctx:", ""))
			for _, c := range softSplitCommaSpace(rest) {
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

// softSplitCommaSpace ports Java String.split(", *").
func softSplitCommaSpace(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, tools.JavaStringTrim(p))
	}
	return out
}

func softUTF16Len(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
