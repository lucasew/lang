package languagemodel

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"

// Google sentence markers port LanguageModel.GOOGLE_SENTENCE_*.
const (
	GoogleSentenceStart = "_START_"
	GoogleSentenceEnd   = "_END_"
)

// LanguageModel ports org.languagetool.languagemodel.LanguageModel.
type LanguageModel interface {
	GetPseudoProbability(context []string) ngrams.Probability
	// Close releases resources (no-op for in-memory models).
	Close() error
}

// FuncLanguageModel adapts a function (Close is no-op).
type FuncLanguageModel func(context []string) ngrams.Probability

func (f FuncLanguageModel) GetPseudoProbability(context []string) ngrams.Probability {
	return f(context)
}
func (f FuncLanguageModel) Close() error { return nil }

// UniformLanguageModel returns a constant probability for tests.
func UniformLanguageModel(prob float64, coverage float32) LanguageModel {
	return FuncLanguageModel(func([]string) ngrams.Probability {
		return ngrams.NewProbabilitySimple(prob, coverage)
	})
}
