package suggestions

import (
	"fmt"
	"math"
	"sort"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik/suggestions_ordering"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/symspell/implementation"
)

// LanguageModelHook ports the LanguageModel surface used by
// SuggestionsOrdererFeatureExtractor (getPseudoProbability + BaseLanguageModel.getCount).
type LanguageModelHook interface {
	// PseudoProbability returns LanguageModel.getPseudoProbability(tokens).getProb().
	PseudoProbability(tokens []string) float64
	// Count ports BaseLanguageModel.getCount(String).
	Count(word string) int64
}

// ngramLM adapts LanguageModelHook to ngrams.LanguageModel for 3gram utils.
type ngramLM struct{ h LanguageModelHook }

func (n ngramLM) GetPseudoProbability(tokens []string) ngrams.Probability {
	if n.h == nil {
		return ngrams.NewProbabilitySimple(0, 0)
	}
	return ngrams.NewProbabilitySimple(n.h.PseudoProbability(tokens), 1)
}

// SuggestionsOrdererFeatureExtractor ports
// org.languagetool.rules.spelling.suggestions.SuggestionsOrdererFeatureExtractor.
type SuggestionsOrdererFeatureExtractor struct {
	// LM is the language model (Java languageModel). Nil → isMlAvailable false.
	LM LanguageModelHook
	// Tokenize is the language Google-style word tokenizer for 3gram context
	// (Java LanguageModelUtils.getGoogleStyleWordTokenizer(language)).
	// Nil → 3gram probability is 0 (fail-closed).
	Tokenize ngrams.TokenizerFunc

	// TopN, Score, MistakeProb are set from SuggestionsChanges experiment
	// (Java initParameters) and may be overridden for tests.
	TopN        int
	Score       string
	MistakeProb float64
}

// NewSuggestionsOrdererFeatureExtractor ports the Java constructor
// (initParameters loads experiment topN/score/levenstheinProb when present).
func NewSuggestionsOrdererFeatureExtractor(lm LanguageModelHook) *SuggestionsOrdererFeatureExtractor {
	e := &SuggestionsOrdererFeatureExtractor{
		LM:          lm,
		TopN:        -1,
		Score:       "",
		MistakeProb: 1.0,
	}
	e.initParameters()
	return e
}

// initParameters ports SuggestionsOrdererFeatureExtractor.initParameters.
// Java NPEs if no experiment; Go only applies overrides when experiment is set
// (same defaults as Java fields: topN=-1, score=null/empty, mistakeProb=1.0).
func (e *SuggestionsOrdererFeatureExtractor) initParameters() {
	if e == nil {
		return
	}
	e.TopN = -1
	e.Score = ""
	e.MistakeProb = 1.0
	s := GetSuggestionsChanges()
	if s == nil {
		return
	}
	exp := s.GetCurrentExperiment()
	if exp == nil || exp.Parameters == nil {
		return
	}
	if v, ok := exp.Parameters["topN"]; ok {
		switch n := v.(type) {
		case int:
			e.TopN = n
		case int32:
			e.TopN = int(n)
		case int64:
			e.TopN = int(n)
		case float64: // JSON numbers
			e.TopN = int(n)
		}
	}
	if v, ok := exp.Parameters["score"].(string); ok {
		e.Score = v
	}
	if v, ok := exp.Parameters["levenstheinProb"]; ok {
		switch p := v.(type) {
		case float64:
			e.MistakeProb = p
		case float32:
			e.MistakeProb = float64(p)
		case int:
			e.MistakeProb = float64(p)
		}
	}
}

func (e *SuggestionsOrdererFeatureExtractor) IsMlAvailable() bool {
	return e != nil && e.LM != nil
}

func (e *SuggestionsOrdererFeatureExtractor) OrderSuggestions(suggestions []string, word string, sentence *languagetool.AnalyzedSentence, startPos int) []*rules.SuggestedReplacement {
	ordered, _ := e.ComputeFeatures(suggestions, word, sentence, startPos)
	return ordered
}

// ComputeFeatures ports SuggestionsOrdererFeatureExtractor.computeFeatures.
// Returns ordered candidates (with per-candidate Features) and match-level features
// (Java SortedMap: candidateCount).
func (e *SuggestionsOrdererFeatureExtractor) ComputeFeatures(
	suggestions []string,
	word string,
	sentence *languagetool.AnalyzedSentence,
	startPos int,
) ([]*rules.SuggestedReplacement, map[string]float32) {
	if len(suggestions) == 0 {
		return nil, map[string]float32{}
	}
	// Re-apply experiment overrides (Java fields set at construction; re-read
	// keeps process-wide experiment switches live for tests).
	if s := GetSuggestionsChanges(); s != nil {
		if exp := s.GetCurrentExperiment(); exp != nil && exp.Parameters != nil {
			if v, ok := exp.Parameters["topN"]; ok {
				switch n := v.(type) {
				case int:
					e.TopN = n
				case int64:
					e.TopN = int(n)
				case float64:
					e.TopN = int(n)
				}
			}
			if v, ok := exp.Parameters["score"].(string); ok {
				e.Score = v
			}
			if v, ok := exp.Parameters["levenstheinProb"]; ok {
				switch p := v.(type) {
				case float64:
					e.MistakeProb = p
				case float32:
					e.MistakeProb = float64(p)
				case int:
					e.MistakeProb = float64(p)
				}
			}
		}
	}
	topN := e.TopN
	if topN <= 0 {
		topN = len(suggestions)
	}
	if topN > len(suggestions) {
		topN = len(suggestions)
	}
	topSuggestions := suggestions[:topN]

	// Java: EditDistance levenstheinDistance = new EditDistance(word, DistanceAlgorithm.Damerau);
	ed := implementation.NewEditDistance(word, implementation.Damerau)

	feats := make([]*feature, 0, len(topSuggestions))
	for _, candidate := range topSuggestions {
		var prob1, prob3 float64
		var wordCount int64
		if e.LM != nil {
			prob1 = e.LM.PseudoProbability([]string{candidate})
			wordCount = e.LM.Count(candidate)
			// Java LanguageModelUtils.get3gramProbabilityFor(language, languageModel, startPos, sentence, candidate)
			prob3 = ngrams.Get3gramProbabilityForSentence(ngramLM{e.LM}, startPos, sentence, candidate, e.Tokenize)
		}
		// Java: levenstheinDistance.compare(candidate, 3)
		leven := ed.Compare(candidate, 3)
		if leven < 0 {
			// Java compare returns -1 when > maxDistance; Feature still stores the int as-is.
			// Keep -1 so ranking/features match the wire value (cast to float later).
		}
		jw := jaroWinkler(word, candidate)
		detailed := suggestions_ordering.Compare(word, candidate)
		feats = append(feats, newFeature(prob1, prob3, wordCount, leven, detailed, jw, candidate, e))
	}
	if e.Score != "noop" {
		sort.SliceStable(feats, func(i, j int) bool {
			// Java Feature.compareTo: descending mean probability
			return feats[i].meanProbability() > feats[j].meanProbability()
		})
	}

	matchData := map[string]float32{
		"candidateCount": float32(len(feats)),
	}
	out := make([]*rules.SuggestedReplacement, 0, len(feats))
	for _, f := range feats {
		sr := rules.NewSuggestedReplacement(f.word)
		sr.SetFeatures(f.getData())
		out = append(out, sr)
	}
	return out, matchData
}

// feature ports SuggestionsOrdererFeatureExtractor.Feature.
type feature struct {
	prob1gram            float64
	prob3gram            float64
	wordCount            int64
	levenshteinDistance  int
	detailedDistance     suggestions_ordering.Distance
	jaroWrinklerDistance float64
	word                 string
	score                string
	mistakeProb          float64
}

func newFeature(
	prob1, prob3 float64,
	wordCount int64,
	levenshteinDistance int,
	detailed suggestions_ordering.Distance,
	jaroWinkler float64,
	word string,
	e *SuggestionsOrdererFeatureExtractor,
) *feature {
	score := ""
	mistake := 1.0
	if e != nil {
		score = e.Score
		mistake = e.MistakeProb
	}
	return &feature{
		prob1gram:            prob1,
		prob3gram:            prob3,
		wordCount:            wordCount,
		levenshteinDistance:  levenshteinDistance,
		detailedDistance:     detailed,
		jaroWrinklerDistance: jaroWinkler,
		word:                 word,
		score:                score,
		mistakeProb:          mistake,
	}
}

func factorial(n int) int {
	// Java Feature.factorial
	factor := n
	result := 1
	for factor > 1 {
		result *= factor
		factor--
	}
	return result
}

func binomialCoefficient(n, k int) int {
	// Java: factorial(n) / (factorial(k) * factorial(n - k)) — integer division
	return factorial(n) / (factorial(k) * factorial(n-k))
}

func binomialProbability(p float64, n, k int) float64 {
	return float64(binomialCoefficient(n, k)) * math.Pow(p, float64(k)) * math.Pow(1-p, float64(n-k))
}

// meanProbability ports Feature.getMeanProbability (score-dependent).
func (f *feature) meanProbability() float64 {
	// Java: double ngramProb = Math.log(prob1gram) + Math.log(prob3gram);
	ngramProb := math.Log(f.prob1gram) + math.Log(f.prob3gram)
	switch f.score {
	case "ngrams+levensthein":
		misspellingProb := math.Pow(f.mistakeProb, float64(f.levenshteinDistance))
		return ngramProb + math.Log(misspellingProb)
	case "ngrams":
		return ngramProb
	case "ngrams+binomialLevensthein":
		// Java: binomialProbability(mistakeProb, word.length(), levenshteinDistance)
		// word.length() is UTF-16
		misspellingProb := binomialProbability(f.mistakeProb, javaStringLenUTF16(f.word), f.levenshteinDistance)
		return ngramProb + math.Log(misspellingProb)
	case "noop":
		return 0
	default:
		// Java: throw new RuntimeException("Unknown scoring method: " + score);
		panic(fmt.Sprintf("Unknown scoring method: %s", f.score))
	}
}

// getData ports Feature.getData SortedMap keys/values.
func (f *feature) getData() map[string]float32 {
	return map[string]float32{
		"prob1gram":    float32(f.prob1gram),
		"prob3gram":    float32(f.prob3gram),
		"wordCount":    float32(f.wordCount),
		"levensthein":  float32(f.levenshteinDistance),
		"jaroWrinkler": float32(f.jaroWrinklerDistance),
		"inserts":      float32(f.detailedDistance.Inserts),
		"deletes":      float32(f.detailedDistance.Deletes),
		"replaces":     float32(f.detailedDistance.Replaces),
		"transposes":   float32(f.detailedDistance.Transposes),
		"wordLength":   float32(javaStringLenUTF16(f.word)),
	}
}

func javaStringLenUTF16(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// jaroWinkler ports Apache Commons Text JaroWinklerDistance (similarity in [0,1]).
// CharSequence indexing uses UTF-16 code units (Java charAt).
func jaroWinkler(s1, s2 string) float64 {
	a := utf16.Encode([]rune(s1))
	b := utf16.Encode([]rune(s2))
	if len(a) == 0 && len(b) == 0 {
		return 1
	}
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	// Commons Text default: matches within floor(max/2)-1
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
	// Commons default threshold 0.7: only boost when jaro >= 0.7
	const threshold = 0.7
	const prefixScale = 0.1
	if jaro < threshold {
		return jaro
	}
	prefix := 0
	for i := 0; i < len(a) && i < len(b) && i < 4; i++ {
		if a[i] == b[i] {
			prefix++
		} else {
			break
		}
	}
	return jaro + float64(prefix)*prefixScale*(1-jaro)
}

var _ SuggestionsOrderer = (*SuggestionsOrdererFeatureExtractor)(nil)
