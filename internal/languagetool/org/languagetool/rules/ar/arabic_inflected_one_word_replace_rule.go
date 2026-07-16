package ar

import (
	"embed"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/inflected_one_word.txt
var inflectedOneWordFS embed.FS

var (
	inflectedOnce sync.Once
	inflectedMap  map[string]rules.SuggestionWithMessage
)

func loadInflectedOneWord() map[string]rules.SuggestionWithMessage {
	inflectedOnce.Do(func() {
		inflectedMap = parseARInflectedMap()
	})
	return inflectedMap
}

func parseARInflectedMap() map[string]rules.SuggestionWithMessage {
	b, err := inflectedOneWordFS.ReadFile("data/inflected_one_word.txt")
	if err != nil {
		panic(err)
	}
	out := map[string]rules.SuggestionWithMessage{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		msg := ""
		if len(parts) == 2 {
			msg = parts[1]
		}
		kv := strings.SplitN(parts[0], "=", 2)
		if len(kv) < 2 {
			continue
		}
		wrong := strings.TrimSpace(kv[0])
		sug := strings.TrimSpace(kv[1])
		out[wrong] = rules.SuggestionWithMessage{Suggestion: sug, Message: msg}
	}
	return out
}

// ArabicInflectedOneWordReplaceRule is a surface stand-in for
// org.languagetool.rules.ar.ArabicInflectedOneWordReplaceRule.
// Without the Arabic tagger/synthesizer, matches lemmas and common
// proclitic/enclitic surface variants (و/ب/ال… + plural endings).
type ArabicInflectedOneWordReplaceRule struct {
	Messages map[string]string
	words    map[string]rules.SuggestionWithMessage
}

func NewArabicInflectedOneWordReplaceRule(messages map[string]string) *ArabicInflectedOneWordReplaceRule {
	return &ArabicInflectedOneWordReplaceRule{
		Messages: messages,
		words:    loadInflectedOneWord(),
	}
}

func (r *ArabicInflectedOneWordReplaceRule) GetID() string { return "AR_INFLECTED_ONE_WORD" }

func arSurfaceStems(token string) []string {
	cands := map[string]struct{}{token: {}}
	prefixes := []string{"وال", "بال", "كال", "فال", "لل", "ال", "و", "ب", "ك", "ف", "ل", "أ"}
	work := []string{token}
	for len(work) > 0 {
		w := work[0]
		work = work[1:]
		for _, p := range prefixes {
			if strings.HasPrefix(w, p) && utf8.RuneCountInString(w) > utf8.RuneCountInString(p)+1 {
				rest := strings.TrimPrefix(w, p)
				if _, ok := cands[rest]; !ok {
					cands[rest] = struct{}{}
					work = append(work, rest)
				}
			}
		}
	}
	suffixes := []string{"هما", "كما", "هم", "هن", "كم", "كن", "نا", "ها", "ه", "ك", "ي", "ا", "ان", "ين", "ون", "ات", "ة", "ًا", "ٌ", "ٍ", "َ", "ُ", "ِ", "ْ", "ّ"}
	more := make([]string, 0, len(cands))
	for c := range cands {
		more = append(more, c)
	}
	for _, c := range more {
		w := c
		for {
			stripped := false
			for _, s := range suffixes {
				if strings.HasSuffix(w, s) && utf8.RuneCountInString(w) > utf8.RuneCountInString(s)+1 {
					w = strings.TrimSuffix(w, s)
					cands[w] = struct{}{}
					stripped = true
					break
				}
			}
			if !stripped {
				break
			}
		}
	}
	out := make([]string, 0, len(cands))
	for c := range cands {
		out = append(out, c)
	}
	return out
}

func (r *ArabicInflectedOneWordReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		word := tok.GetToken()
		var hit *rules.SuggestionWithMessage
		if swm, ok := r.words[word]; ok {
			cp := swm
			hit = &cp
		} else {
			for _, stem := range arSurfaceStems(word) {
				if swm, ok := r.words[stem]; ok {
					cp := swm
					hit = &cp
					break
				}
			}
		}
		if hit == nil {
			continue
		}
		msg := " لا تقل '" + word + "' بل قل: " + hit.Suggestion
		if hit.Message != "" {
			msg = hit.Message
		}
		rm := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
		rm.ShortMessage = "خطأ، يفضل أن  يقال:"
		rm.SetSuggestedReplacement(hit.Suggestion)
		matches = append(matches, rm)
	}
	return matches
}
