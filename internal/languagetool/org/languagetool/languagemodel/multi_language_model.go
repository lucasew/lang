package languagemodel

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

// MultiLanguageModel ports org.languagetool.languagemodel.MultiLanguageModel.
type MultiLanguageModel struct {
	lms []LanguageModel
}

func NewMultiLanguageModel(lms []LanguageModel) *MultiLanguageModel {
	if len(lms) == 0 {
		panic("List of language models is empty")
	}
	return &MultiLanguageModel{lms: append([]LanguageModel(nil), lms...)}
}

func (m *MultiLanguageModel) GetPseudoProbability(context []string) ngrams.Probability {
	var prob float64
	var coverage float32
	var occurrences int64
	for _, lm := range m.lms {
		p := lm.GetPseudoProbability(context)
		prob += p.GetProb()
		coverage += p.GetCoverage()
		occurrences += p.GetOccurrences()
	}
	n := float32(len(m.lms))
	return ngrams.NewProbability(prob, coverage/n, occurrences)
}

func (m *MultiLanguageModel) Close() error {
	var first error
	for _, lm := range m.lms {
		if err := lm.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (m *MultiLanguageModel) String() string {
	return fmt.Sprint(m.lms)
}
