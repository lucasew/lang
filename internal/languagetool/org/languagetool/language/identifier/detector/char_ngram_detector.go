package detector

import (
	"math"
	"strings"
	"unicode"
)

// CharNGramDetector is a lightweight stand-in for Java NGramDetector zip models.
// Stores per-language character n-gram relative frequencies and scores text by log-likelihood.
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

func normalizeNGramText(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
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
