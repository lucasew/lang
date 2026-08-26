package languagemodel

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

// MapLanguageModel is an in-memory CountProvider for tests and tiny models.
// Multi-token ngram keys are joined with \x1f.
type MapLanguageModel struct {
	Uni   map[string]int64
	N     map[string]int64
	Total int64
	base  *BaseLanguageModel
}

func NewMapLanguageModel() *MapLanguageModel {
	m := &MapLanguageModel{
		Uni: map[string]int64{},
		N:   map[string]int64{},
	}
	m.base = NewBaseLanguageModel(m)
	return m
}

func ngramKey(tokens []string) string {
	return strings.Join(tokens, "\x1f")
}

func (m *MapLanguageModel) GetCountToken(token string) int64 { return m.Uni[token] }

func (m *MapLanguageModel) GetCount(tokens []string) int64 {
	if len(tokens) == 1 {
		return m.GetCountToken(tokens[0])
	}
	return m.N[ngramKey(tokens)]
}

func (m *MapLanguageModel) GetTotalTokenCount() int64 {
	if m.Total > 0 {
		return m.Total
	}
	var s int64
	for _, v := range m.Uni {
		s += v
	}
	return s
}

func (m *MapLanguageModel) Add(tokens []string, count int64) {
	if len(tokens) == 1 {
		m.Uni[tokens[0]] += count
		return
	}
	m.N[ngramKey(tokens)] += count
}

func (m *MapLanguageModel) GetPseudoProbability(context []string) ngrams.Probability {
	return m.base.GetPseudoProbability(context)
}

func (m *MapLanguageModel) Close() error { return nil }
