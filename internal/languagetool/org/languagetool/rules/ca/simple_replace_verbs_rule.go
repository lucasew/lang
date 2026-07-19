package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_verbs.txt
var verbsFS embed.FS

var (
	verbsOnce sync.Once
	verbsMap  map[string][]string
)

// pVerb ports VerbSynthesizer.pVerb used by SimpleReplaceVerbsRule (Pattern "V.*").
// incorrectVerbChunk is defined in catalan_suppress_misspelled_suggestions_filter.go.
const pVerb = `V.*`

func loadVerbs() map[string][]string {
	verbsOnce.Do(func() {
		f, err := verbsFS.Open("data/replace_verbs.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		verbsMap = m
	})
	return verbsMap
}

// SimpleReplaceVerbsRule ports org.languagetool.rules.ca.SimpleReplaceVerbsRule.
// Match requires chunk _incorrect_verb_ and a V.* reading with lemma in replace_verbs.txt;
// suggestions go through AdjustVerbSuggestionsFilter (actions=None).
// Without chunk/POS/synth, fail closed (no surface invent of conjugations).
type SimpleReplaceVerbsRule struct {
	*rules.AbstractSimpleReplaceRule
	// Filter is the AdjustVerbSuggestionsFilter used after createRuleMatch.
	// When nil, a default filter is used (needs Synthesize set for non-nil Accept).
	Filter *AdjustVerbSuggestionsFilter
}

func NewSimpleReplaceVerbsRule(messages map[string]string) *SimpleReplaceVerbsRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:          messages,
		WrongWords:        loadVerbs(),
		CaseSensitive:     false,
		CheckLemmas:       false,
		IgnoreTaggedWords: true, // Java setIgnoreTaggedWords()
		ID:                "CA_SIMPLE_REPLACE_VERBS",
		Description:       "Verb incorrecte: $match",
		ShortMsg:          "Verb incorrecte",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Verb incorrecte."
		},
	}
	return &SimpleReplaceVerbsRule{AbstractSimpleReplaceRule: base}
}

// Match ports SimpleReplaceVerbsRule.match (chunk + lemma path, not surface AbstractSimpleReplace).
func (r *SimpleReplaceVerbsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	wrong := loadVerbs()
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	// Java starts at index 1 (skip SENT_START)
	for index := 1; index < len(tokens); index++ {
		tok := tokens[index]
		if tok == nil {
			continue
		}
		// Java: continue unless chunk contains _incorrect_verb_ AND not N*/A*
		if !hasChunk(tok, incorrectVerbChunk) {
			continue
		}
		if tok.HasPosTagStartingWith("N") || tok.HasPosTagStartingWith("A") {
			continue
		}
		at := tok.ReadingWithTagRegex(pVerb)
		if at == nil || at.GetLemma() == nil {
			continue
		}
		lemma := *at.GetLemma()
		replacementInfinitives, ok := wrong[lemma]
		if !ok || len(replacementInfinitives) == 0 {
			continue
		}
		// createRuleMatch with infinitives as suggestion seeds (Java)
		msg := "Verb incorrecte."
		if r.MessageFn != nil {
			msg = r.MessageFn(lemma, replacementInfinitives)
		}
		potential := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
		potential.ShortMessage = "Verb incorrecte"
		// copy replacements (filter mutates list in Java for ctx; seeds are lemmas)
		seeds := append([]string(nil), replacementInfinitives...)
		potential.SetSuggestedReplacements(seeds)

		filter := r.Filter
		if filter == nil {
			filter = NewAdjustVerbSuggestionsFilter()
		}
		// Java: argumentsMap = Map.of("actions", "None")
		finalMatch := filter.AcceptRuleMatch(potential, map[string]string{"actions": "None"}, 0, nil, nil)
		if finalMatch != nil {
			ruleMatches = append(ruleMatches, finalMatch)
		}
	}
	return ruleMatches
}

func hasChunk(tok *languagetool.AnalyzedTokenReadings, chunk string) bool {
	for _, c := range tok.GetChunkTags() {
		if c == chunk {
			return true
		}
	}
	return false
}
