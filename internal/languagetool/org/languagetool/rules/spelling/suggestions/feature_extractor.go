package suggestions

import (
	"sort"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FeatureVector holds per-candidate ranking features
// (ports SuggestionsOrdererFeatureExtractor.Feature subset without LM).
type FeatureVector struct {
	Candidate   string
	Levenshtein int
	JaroWinkler float64
	UnigramProb float64
	TrigramProb float64
	WordCount   int64
}

// LanguageModelHook is optional; nil skips probability features.
type LanguageModelHook interface {
	PseudoProbability(tokens []string) float64
	Count(word string) int64
}

// SuggestionsOrdererFeatureExtractor ports
// org.languagetool.rules.spelling.suggestions.SuggestionsOrdererFeatureExtractor
// with pluggable LM (full n-gram / detailed Damerau deferred).
type SuggestionsOrdererFeatureExtractor struct {
	LM   LanguageModelHook
	TopN int
	// Score is "noop" to keep input order; anything else sorts by Levenshtein then JW.
	Score string
}

func NewSuggestionsOrdererFeatureExtractor(lm LanguageModelHook) *SuggestionsOrdererFeatureExtractor {
	return &SuggestionsOrdererFeatureExtractor{LM: lm, TopN: -1, Score: "levenshtein"}
}

func (e *SuggestionsOrdererFeatureExtractor) IsMlAvailable() bool {
	return e != nil && e.LM != nil
}

func (e *SuggestionsOrdererFeatureExtractor) OrderSuggestions(suggestions []string, word string, sentence *languagetool.AnalyzedSentence, startPos int) []*rules.SuggestedReplacement {
	ordered, _ := e.ComputeFeatures(suggestions, word, sentence, startPos)
	return ordered
}

// ComputeFeatures returns ordered candidates and aggregate feature map for the match.
func (e *SuggestionsOrdererFeatureExtractor) ComputeFeatures(suggestions []string, word string, _ *languagetool.AnalyzedSentence, _ int) ([]*rules.SuggestedReplacement, map[string]float32) {
	if len(suggestions) == 0 {
		return nil, map[string]float32{}
	}
	topN := e.TopN
	if topN <= 0 || topN > len(suggestions) {
		topN = len(suggestions)
	}
	// apply experiment overrides when running
	if s := GetSuggestionsChanges(); s != nil {
		if exp := s.GetCurrentExperiment(); exp != nil && exp.Parameters != nil {
			if v, ok := exp.Parameters["topN"].(int); ok && v > 0 {
				topN = v
				if topN > len(suggestions) {
					topN = len(suggestions)
				}
			}
			if v, ok := exp.Parameters["score"].(string); ok {
				e.Score = v
			}
		}
	}
	feats := make([]FeatureVector, 0, topN)
	for _, cand := range suggestions[:topN] {
		fv := FeatureVector{
			Candidate:   cand,
			Levenshtein: lev(word, cand),
			JaroWinkler: jaroWinkler(word, cand),
		}
		if e.LM != nil {
			fv.UnigramProb = e.LM.PseudoProbability([]string{cand})
			fv.WordCount = e.LM.Count(cand)
		}
		feats = append(feats, fv)
	}
	if e.Score != "noop" {
		sort.SliceStable(feats, func(i, j int) bool {
			if feats[i].Levenshtein != feats[j].Levenshtein {
				return feats[i].Levenshtein < feats[j].Levenshtein
			}
			return feats[i].JaroWinkler > feats[j].JaroWinkler
		})
	}
	out := make([]*rules.SuggestedReplacement, 0, len(feats))
	agg := map[string]float32{}
	for i, f := range feats {
		sr := rules.NewSuggestedReplacement(f.Candidate)
		// rough confidence: higher JW / lower edit distance
		c := float32(f.JaroWinkler)
		if f.Levenshtein > 0 {
			c = c / float32(1+f.Levenshtein)
		}
		sr.Confidence = &c
		out = append(out, sr)
		if i == 0 {
			agg["top_levenshtein"] = float32(f.Levenshtein)
			agg["top_jaro_winkler"] = float32(f.JaroWinkler)
			agg["top_unigram"] = float32(f.UnigramProb)
		}
	}
	return out, agg
}

// jaroWinkler is a compact JW similarity in [0,1].
func jaroWinkler(s1, s2 string) float64 {
	a, b := []rune(s1), []rune(s2)
	if len(a) == 0 && len(b) == 0 {
		return 1
	}
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	matchDist := len(a)
	if len(b) > matchDist {
		matchDist = len(b)
	}
	matchDist = matchDist/2 - 1
	if matchDist < 0 {
		matchDist = 0
	}
	aM := make([]bool, len(a))
	bM := make([]bool, len(b))
	matches := 0
	for i := range a {
		start := i - matchDist
		if start < 0 {
			start = 0
		}
		end := i + matchDist + 1
		if end > len(b) {
			end = len(b)
		}
		for j := start; j < end; j++ {
			if bM[j] || a[i] != b[j] {
				continue
			}
			aM[i] = true
			bM[j] = true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0
	}
	// transpositions
	k := 0
	trans := 0
	for i := range a {
		if !aM[i] {
			continue
		}
		for !bM[k] {
			k++
		}
		if a[i] != b[k] {
			trans++
		}
		k++
	}
	m := float64(matches)
	jaro := (m/float64(len(a)) + m/float64(len(b)) + (m-float64(trans)/2)/m) / 3
	// winkler prefix
	prefix := 0
	for i := 0; i < len(a) && i < len(b) && i < 4; i++ {
		if a[i] == b[i] {
			prefix++
		} else {
			break
		}
	}
	return jaro + float64(prefix)*0.1*(1-jaro)
}

var _ SuggestionsOrderer = (*SuggestionsOrdererFeatureExtractor)(nil)
