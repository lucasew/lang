package ca

import (
	"embed"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/replace_dnv.txt
var dnvFS embed.FS

var (
	dnvOnce sync.Once
	dnvMap  map[string][]string
)

func loadDNV() map[string][]string {
	dnvOnce.Do(func() {
		f, err := dnvFS.Open("data/replace_dnv.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		out := make(map[string][]string, len(m))
		for k, v := range m {
			out[strings.ToLower(k)] = v
		}
		dnvMap = out
	})
	return dnvMap
}

// SimpleReplaceDNVRule ports org.languagetool.rules.ca.SimpleReplaceDNVRule
// without Catalan synthesizer — surface lemmas + light plural heuristics.
type SimpleReplaceDNVRule struct {
	messages map[string]string
}

func NewSimpleReplaceDNVRule(messages map[string]string) *SimpleReplaceDNVRule {
	_ = loadDNV()
	return &SimpleReplaceDNVRule{messages: messages}
}

func (r *SimpleReplaceDNVRule) GetID() string { return "CA_SIMPLE_REPLACE_DNV" }

func (r *SimpleReplaceDNVRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	m := loadDNV()
	tokens := sentence.GetTokensWithoutWhitespace()
	covered := map[int]bool{}
	var out []*rules.RuleMatch

	// Prefer joined L'/D'/… + word (tokenizer splits apostrophes).
	for i := 0; i+2 < len(tokens); i++ {
		a, b, c := tokens[i], tokens[i+1], tokens[i+2]
		if a.IsSentenceStart() {
			continue
		}
		at, bt, ct := a.GetToken(), b.GetToken(), c.GetToken()
		if len(at) == 1 && (bt == "'" || bt == "’") && !b.IsWhitespaceBefore() && !c.IsWhitespaceBefore() {
			reps, ok := lookupDNV(strings.ToLower(ct), m)
			if !ok {
				continue
			}
			from := c.GetStartPos()
			if covered[from] {
				continue
			}
			covered[from] = true
			out = append(out, r.makeMatch(sentence, from, c.GetEndPos(), ct, reps))
		}
	}

	for _, tok := range tokens {
		if tok.IsSentenceStart() || tok.IsImmunized() {
			continue
		}
		from := tok.GetStartPos()
		if covered[from] {
			continue
		}
		t := tok.GetToken()
		// skip bare article letters and apostrophes
		if t == "'" || t == "’" || (len(t) == 1 && strings.ContainsAny(t, "LlDdNnSs")) {
			continue
		}
		reps, ok := lookupDNV(strings.ToLower(t), m)
		if !ok {
			continue
		}
		covered[from] = true
		out = append(out, r.makeMatch(sentence, from, from+utf16Len(t), t, reps))
	}
	return out
}

func (r *SimpleReplaceDNVRule) makeMatch(sentence *languagetool.AnalyzedSentence, from, to int, surface string, reps []string) *rules.RuleMatch {
	final := caseAdjustAll(surface, reps)
	rm := rules.NewRuleMatch(r, sentence, from, to, "Paraula admesa pel DNV (AVL), però no per altres diccionaris.")
	rm.ShortMessage = "Paraula admesa només pel DNV (AVL)."
	rm.SetSuggestedReplacements(final)
	return rm
}

func lookupDNV(token string, m map[string][]string) ([]string, bool) {
	if r, ok := m[token]; ok {
		return r, true
	}
	if strings.HasSuffix(token, "s") {
		base := strings.TrimSuffix(token, "s")
		if r, ok := m[base]; ok {
			return pluralizeCASuggestions(r), true
		}
		if strings.HasSuffix(token, "es") {
			stem := strings.TrimSuffix(token, "es")
			if r, ok := m[stem+"a"]; ok {
				return pluralizeCASuggestions(r), true
			}
		}
	}
	return nil, false
}

func pluralizeCASuggestions(reps []string) []string {
	var out []string
	for _, s := range reps {
		switch {
		case strings.HasSuffix(s, "c"):
			out = append(out, s+"s", s+"os")
		case strings.HasSuffix(s, "a"):
			out = append(out, strings.TrimSuffix(s, "a")+"es")
		case strings.HasSuffix(s, "ó"):
			out = append(out, strings.TrimSuffix(s, "ó")+"ons")
		case strings.HasSuffix(s, "ió"):
			out = append(out, strings.TrimSuffix(s, "ió")+"ions")
		default:
			out = append(out, s+"s")
		}
	}
	return out
}

func caseAdjustAll(surface string, reps []string) []string {
	out := make([]string, len(reps))
	for i, s := range reps {
		switch {
		case tools.IsAllUppercase(surface):
			out[i] = strings.ToUpper(s)
		case tools.StartsWithUppercase(surface):
			out[i] = tools.UppercaseFirstChar(s)
		default:
			out[i] = s
		}
	}
	return out
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
