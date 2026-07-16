package ar

import (
	"embed"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/verb_trans_to_untrans2.txt
var transVerbFS embed.FS

var (
	transVerbOnce sync.Once
	// lemma (with/without diacritics) → required preposition(s)
	transVerbMap map[string][]string
)

func stripArabicDiacritics(s string) string {
	var b strings.Builder
	for _, r := range s {
		// Arabic combining diacritics U+064B–U+065F, tatweel U+0640
		if (r >= 0x064B && r <= 0x065F) || r == 0x0640 || r == 0x0670 {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func loadTransVerbs() map[string][]string {
	transVerbOnce.Do(func() {
		b, err := transVerbFS.ReadFile("data/verb_trans_to_untrans2.txt")
		if err != nil {
			panic(err)
		}
		transVerbMap = map[string][]string{}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = strings.TrimSpace(line[:i])
			}
			kv := strings.SplitN(line, "=", 2)
			if len(kv) < 2 {
				continue
			}
			lemma := strings.TrimSpace(kv[0])
			preps := strings.Split(strings.TrimSpace(kv[1]), "|")
			for i := range preps {
				preps[i] = strings.TrimSpace(preps[i])
			}
			transVerbMap[lemma] = preps
			transVerbMap[stripArabicDiacritics(lemma)] = preps
		}
	})
	return transVerbMap
}

// ArabicTransVerbRule is a surface stand-in for ArabicTransVerbRule.
// Full fidelity needs Arabic tagger/synthesizer for attached pronouns.
// Surface: if token starts with a listed verb lemma and is longer (clitic object),
// and the next token is not the expected preposition, flag it.
type ArabicTransVerbRule struct {
	Messages map[string]string
	verbs    map[string][]string
}

func NewArabicTransVerbRule(messages map[string]string) *ArabicTransVerbRule {
	return &ArabicTransVerbRule{Messages: messages, verbs: loadTransVerbs()}
}

func (r *ArabicTransVerbRule) GetID() string { return "AR_VERB_TRANSITIVE_IINDIRECT" }

func (r *ArabicTransVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		word := tok.GetToken()
		wordND := stripArabicDiacritics(word)
		var preps []string
		var lemma string
		// longest lemma prefix match
		for lem, p := range r.verbs {
			lemND := stripArabicDiacritics(lem)
			if lemND == "" {
				continue
			}
			if wordND == lemND {
				// unattached bare verb — not the "attached" case
				continue
			}
			if strings.HasPrefix(wordND, lemND) && utf8.RuneCountInString(wordND) > utf8.RuneCountInString(lemND) {
				if lemma == "" || utf8.RuneCountInString(lemND) > utf8.RuneCountInString(stripArabicDiacritics(lemma)) {
					lemma = lem
					preps = p
				}
			}
		}
		if lemma == "" || len(preps) == 0 {
			continue
		}
		// next token should be the preposition; if missing or wrong, flag
		okPrep := false
		if i+1 < len(tokens) {
			next := stripArabicDiacritics(tokens[i+1].GetToken())
			for _, p := range preps {
				if next == stripArabicDiacritics(p) || strings.HasPrefix(next, stripArabicDiacritics(p)) {
					okPrep = true
					break
				}
			}
		}
		if okPrep {
			continue
		}
		sug := lemma + " " + preps[0]
		msg := "قل " + sug + " بدلا من '" + word + "' لأنّ الفعل متعد بحرف."
		from := tok.GetStartPos()
		to := tok.GetEndPos()
		if i+1 < len(tokens) {
			to = tokens[i+1].GetEndPos()
		}
		rm := rules.NewRuleMatch(r, sentence, from, to, msg)
		rm.ShortMessage = "خطأ في الفعل المتعدي بحرف"
		rm.SetSuggestedReplacement(sug)
		matches = append(matches, rm)
	}
	return matches
}
