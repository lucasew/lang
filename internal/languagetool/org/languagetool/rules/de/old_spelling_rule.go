package de

import (
	"bufio"
	"embed"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/alt_neu.csv
var altNeuFS embed.FS

var oldSpellingExceptions = []string{
	"Schloß Holte", "Schloß Neuhaus", "Schloß Ricklingen", "Schloß-Nauses",
	"Schloß Rötteln", "Klinikum Schloß Winnenden", "Grazer Schloßberg",
	"Höchster Schloß", "Bell Telephone", "Telephone Company", "American Telephone",
	"England Telephone", "Mobile Telephone", "Cordless Telephone", "Telephone Line",
	"World Telephone", "Tip Top", "Hans Joachim Blaß", "kurz fassen",
}

// Short German inflection suffixes used as a stand-in for the Java synthesizer.
// Kept tight (genitive/plural-ish only) to avoid false substring hits like Photo+n → Photon.
var deOldInflSuffixes = []string{
	"", "s", "es",
}

var (
	oldSpellingOnce sync.Once
	oldSpellingKeys []string // longest first
	oldSpellingMap  map[string]string
)

func loadOldSpelling() (map[string]string, []string) {
	oldSpellingOnce.Do(func() {
		f, err := altNeuFS.Open("data/alt_neu.csv")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m := map[string]string{}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			parts := strings.Split(line, ";")
			if len(parts) < 2 {
				continue
			}
			old := strings.TrimSpace(parts[0])
			neu := strings.TrimSpace(parts[1])
			if old == "" || neu == "" || old == neu {
				continue
			}
			if _, ok := m[old]; !ok {
				m[old] = neu
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return len([]rune(keys[i])) > len([]rune(keys[j]))
		})
		oldSpellingMap = m
		oldSpellingKeys = keys
	})
	return oldSpellingMap, oldSpellingKeys
}

// OldSpellingRule ports org.languagetool.rules.de.OldSpellingRule without Aho-Corasick
// or German synthesizer (listed CSV forms + short inflection suffixes).
type OldSpellingRule struct {
	messages map[string]string
	austrian bool
}

func NewOldSpellingRule(messages map[string]string) *OldSpellingRule {
	_, _ = loadOldSpelling()
	return &OldSpellingRule{messages: messages}
}

// NewOldSpellingRuleAT is the Austrian variant (Geschoß remains acceptable).
func NewOldSpellingRuleAT(messages map[string]string) *OldSpellingRule {
	r := NewOldSpellingRule(messages)
	r.austrian = true
	return r
}

func (r *OldSpellingRule) GetID() string { return "OLD_SPELLING_RULE" }

func (r *OldSpellingRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	m, keys := loadOldSpelling()
	text := sentence.GetText()
	tokens := sentence.GetTokensWithoutWhitespace()
	covered := map[int]bool{}
	var out []*rules.RuleMatch

	for i := 0; i < len(tokens); i++ {
		if tokens[i].IsSentenceStart() {
			continue
		}
		fromPos := tokens[i].GetStartPos()
		if covered[fromPos] {
			continue
		}

		// Build phrases of 1..4 tokens (handles "Hot pants", "Corpus delicti", "naß machen").
		var b strings.Builder
		bestEnd := -1
		var bestNeu string
		var bestCovered string
		for j := i; j < len(tokens) && j-i < 4; j++ {
			if tokens[j].IsSentenceStart() {
				break
			}
			if j > i {
				if tokens[j].IsWhitespaceBefore() {
					b.WriteByte(' ')
				}
			}
			b.WriteString(tokens[j].GetToken())
			phrase := b.String()
			if neu, ok := lookupOldSpelling(phrase, m, keys); ok {
				bestEnd = j
				bestNeu = neu
				bestCovered = phrase
			}
		}
		if bestEnd < 0 {
			continue
		}

		from := tokens[i].GetStartPos()
		to := tokens[bestEnd].GetEndPos()
		if r.shouldIgnore(text, from, to, bestCovered) {
			// still consume? no — allow shorter match? skip this start
			continue
		}
		// Austrian Geschoß
		if r.austrian && strings.Contains(strings.ToLower(bestCovered), "geschoß") {
			continue
		}

		// Mark covered start positions for joined tokens
		for k := i; k <= bestEnd; k++ {
			covered[tokens[k].GetStartPos()] = true
		}

		msg := "Diese Schreibweise war nur in der alten Rechtschreibung korrekt."
		suggs := strings.Split(bestNeu, "|")
		// ss/ß tip
		if len(suggs) > 0 {
			ssForm := strings.Replace(suggs[0], "ss", "ß", 1)
			if ssForm == bestCovered {
				msg += " Das Wort wird mit 'ss' geschrieben, wenn davor eine kurz gesprochene Silbe steht."
			}
		}
		rm := rules.NewRuleMatch(r, sentence, from, to, msg)
		rm.ShortMessage = "Alte Rechtschreibung"
		rm.SetSuggestedReplacements(suggs)
		out = append(out, rm)
		i = bestEnd
	}
	return out
}

func lookupOldSpelling(phrase string, m map[string]string, keys []string) (string, bool) {
	if neu, ok := m[phrase]; ok {
		return neu, true
	}
	// Sentence-start capitalization of a lowercase CSV entry (Läßt ← läßt).
	runes := []rune(phrase)
	if len(runes) > 0 && unicode.IsUpper(runes[0]) {
		runes[0] = unicode.ToLower(runes[0])
		low := string(runes)
		if neu, ok := m[low]; ok {
			return capitalizeSuggestions(neu), true
		}
		if neu, ok := matchOldWithInfl(low, keys, m); ok {
			return capitalizeSuggestions(neu), true
		}
	}
	if neu, ok := matchOldWithInfl(phrase, keys, m); ok {
		return neu, true
	}
	return "", false
}

func matchOldWithInfl(token string, keys []string, m map[string]string) (string, bool) {
	if strings.Contains(token, " ") {
		return "", false
	}
	for _, key := range keys {
		if !strings.HasPrefix(token, key) {
			continue
		}
		rest := token[len(key):]
		for _, suf := range deOldInflSuffixes {
			if rest != suf {
				continue
			}
			neu := m[key]
			parts := strings.Split(neu, "|")
			for i, p := range parts {
				parts[i] = p + suf
			}
			return strings.Join(parts, "|"), true
		}
	}
	return "", false
}

func capitalizeSuggestions(neu string) string {
	parts := strings.Split(neu, "|")
	for i, p := range parts {
		parts[i] = tools.UppercaseFirstChar(p)
	}
	return strings.Join(parts, "|")
}

func (r *OldSpellingRule) shouldIgnore(text string, from, to int, covered string) bool {
	// EXCEPTION phrases containing the hit
	windowFrom := from - 20
	if windowFrom < 0 {
		windowFrom = 0
	}
	windowTo := to + 30
	if tl := utf16Len(text); windowTo > tl {
		windowTo = tl
	}
	window := strings.ToLower(substringByUTF16(text, windowFrom, windowTo))
	for _, ex := range oldSpellingExceptions {
		if strings.Contains(window, strings.ToLower(ex)) {
			return true
		}
	}
	// Title + name: "Herr Naß", "Dr. Naß"
	prev := strings.TrimSpace(substringByUTF16(text, max0(from-8), from))
	for _, title := range []string{"Herr", "Frau", "Hr.", "Fr.", "Dr.", "Prof."} {
		if strings.HasSuffix(prev, title) {
			return true
		}
	}
	_ = covered
	return false
}

func substringByUTF16(s string, from, to int) string {
	var b strings.Builder
	pos := 0
	for _, r := range s {
		w := len(utf16.Encode([]rune{r}))
		if pos >= to {
			break
		}
		if pos+w > from {
			b.WriteRune(r)
		}
		pos += w
	}
	return b.String()
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}

func max0(a int) int {
	if a < 0 {
		return 0
	}
	return a
}
