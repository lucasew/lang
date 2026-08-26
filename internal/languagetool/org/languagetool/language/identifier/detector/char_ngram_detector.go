package detector

import (
	"math"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// CharNGramDetector is a lightweight stand-in for Java NGramDetector zip models
// (full BPE vocab encode deferred). Character n-gram relative frequencies only.
type CharNGramDetector struct {
	// N is n-gram order (default 3).
	N int
	// Profiles: lang code → ngram → probability
	Profiles map[string]map[string]float64
}

func NewCharNGramDetector(n int) *CharNGramDetector {
	if n <= 0 {
		n = 3
	}
	return &CharNGramDetector{N: n, Profiles: map[string]map[string]float64{}}
}

// TrainFromText builds a profile for lang from sample text.
func (d *CharNGramDetector) TrainFromText(lang, text string) {
	if d.Profiles == nil {
		d.Profiles = map[string]map[string]float64{}
	}
	counts := countNGrams(normalizeNGramText(text), d.N)
	total := 0
	for _, c := range counts {
		total += c
	}
	if total == 0 {
		return
	}
	prof := map[string]float64{}
	for g, c := range counts {
		prof[g] = float64(c) / float64(total)
	}
	d.Profiles[lang] = prof
}

// DetectLanguages returns lang → relative score in [0,1] (softmax of log probs).
func (d *CharNGramDetector) DetectLanguages(text string) map[string]float64 {
	if d == nil || len(d.Profiles) == 0 {
		return nil
	}
	norm := normalizeNGramText(text)
	grams := countNGrams(norm, d.N)
	if len(grams) == 0 {
		return nil
	}
	const eps = 1e-6
	logs := map[string]float64{}
	var maxLog float64 = math.Inf(-1)
	for lang, prof := range d.Profiles {
		var s float64
		for g, c := range grams {
			p := prof[g]
			if p <= 0 {
				p = eps
			}
			s += float64(c) * math.Log(p)
		}
		logs[lang] = s
		if s > maxLog {
			maxLog = s
		}
	}
	// softmax
	var sum float64
	out := map[string]float64{}
	for lang, s := range logs {
		v := math.Exp(s - maxLog)
		out[lang] = v
		sum += v
	}
	if sum > 0 {
		for lang := range out {
			out[lang] /= sum
		}
	}
	return out
}

// normalizeNGramText mirrors the start of Java NGramDetector.encode:
// Normalizer.normalize(text, NFKC).toLowerCase() then keeps letters/spaces for
// the incomplete char-ngram path (full <NUM>/<KO>/… BPE replacements deferred).
func normalizeNGramText(s string) string {
	// Java: Normalizer.normalize(text, Normalizer.Form.NFKC).toLowerCase()
	s = norm.NFKC.String(s)
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		// Keep letters; Unicode White_Space as space class (NFKC path before BPE).
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	// Collapse runs of ASCII whitespace to a single ' ' (not strings.Fields invent).
	return collapseASCIIWhitespaceRuns(b.String())
}

// collapseASCIIWhitespaceRuns maps any run of [ \t\n\v\f\r] to one ' '.
// Other Unicode spaces that passed IsSpace stay as single chars (not collapsed as Fields would).
func collapseASCIIWhitespaceRuns(s string) string {
	var out strings.Builder
	out.Grow(len(s))
	inWS := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r' {
			if !inWS {
				out.WriteByte(' ')
				inWS = true
			}
			continue
		}
		inWS = false
		// multi-byte UTF-8: write full rune from this index
		// since we only special-case ASCII WS bytes, copy remaining as-is per byte is OK
		// for multi-byte sequences that are not ASCII WS.
		out.WriteByte(c)
	}
	// trim leading/trailing single spaces produced by collapse
	res := out.String()
	for len(res) > 0 && res[0] == ' ' {
		res = res[1:]
	}
	for len(res) > 0 && res[len(res)-1] == ' ' {
		res = res[:len(res)-1]
	}
	return res
}

func countNGrams(s string, n int) map[string]int {
	if n <= 0 {
		n = 3
	}
	// pad
	pad := strings.Repeat(" ", n-1)
	s = pad + s + pad
	rs := []rune(s)
	out := map[string]int{}
	for i := 0; i+n <= len(rs); i++ {
		g := string(rs[i : i+n])
		out[g]++
	}
	return out
}
