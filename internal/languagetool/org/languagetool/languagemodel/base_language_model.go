package languagemodel

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// CountProvider supplies ngram occurrence counts (storage backend).
type CountProvider interface {
	GetCountToken(token string) int64
	GetCount(tokens []string) int64
	GetTotalTokenCount() int64
}

// BaseLanguageModel ports org.languagetool.languagemodel.BaseLanguageModel algorithms
// independent of storage.
type BaseLanguageModel struct {
	Counts CountProvider
	total  *int64
}

func NewBaseLanguageModel(counts CountProvider) *BaseLanguageModel {
	return &BaseLanguageModel{Counts: counts}
}

// GetPseudoProbabilityStupidBackoff ports BaseLanguageModel.getPseudoProbabilityStupidBackoff.
func (m *BaseLanguageModel) GetPseudoProbabilityStupidBackoff(context []string) ngrams.Probability {
	backoff := append([]string(nil), context...)
	maxCoverage := len(context)
	coverage := maxCoverage
	lambda := 1.0
	const lambdaFactor = 0.4
	for len(backoff) > 0 {
		count := m.tryGetCount(backoff)
		if count != 0 {
			baseCount := m.tryGetCount(backoff[:len(backoff)-1])
			if baseCount == 0 {
				baseCount = 1
			}
			prob := float64(count) / float64(baseCount)
			coverageRate := float32(coverage) / float32(maxCoverage)
			return ngrams.NewProbability(lambda*prob, coverageRate, -1)
		}
		coverage--
		backoff = backoff[:len(backoff)-1]
		lambda *= lambdaFactor
	}
	return ngrams.NewProbabilitySimple(0, 0)
}

func (m *BaseLanguageModel) tryGetCount(context []string) int64 {
	if m == nil || m.Counts == nil {
		return 0
	}
	if len(context) == 0 {
		return 0
	}
	if len(context) == 1 {
		return m.Counts.GetCountToken(context[0])
	}
	return m.Counts.GetCount(context)
}

// GetPseudoProbability ports BaseLanguageModel.getPseudoProbability (chain rule + add-1).
func (m *BaseLanguageModel) GetPseudoProbability(context []string) ngrams.Probability {
	if len(context) == 0 {
		panic("index out of bounds: empty context") // Java IndexOutOfBoundsException
	}
	if m == nil || m.Counts == nil {
		return ngrams.NewProbabilitySimple(0, 0)
	}
	if m.total == nil {
		t := m.Counts.GetTotalTokenCount()
		m.total = &t
	}
	totalTokenCount := *m.total
	maxCoverage := 0
	coverage := 0
	firstWordCount := m.Counts.GetCountToken(context[0])
	maxCoverage++
	if firstWordCount > 0 {
		coverage++
	}
	p := float64(firstWordCount+1) / float64(totalTokenCount+1)
	var totalCount int64
	for i := 2; i <= len(context); i++ {
		sub := context[:i]
		phraseCount := m.Counts.GetCount(sub)
		if len(sub) == 3 {
			totalCount = phraseCount
		}
		thisP := float64(phraseCount+1) / float64(firstWordCount+1)
		maxCoverage++
		if phraseCount > 0 {
			coverage++
		}
		p *= thisP
	}
	return ngrams.NewProbability(p, float32(coverage)/float32(maxCoverage), totalCount)
}

func (m *BaseLanguageModel) Close() error { return nil }
