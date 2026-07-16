package ngrams

// LanguageModel is the Go surface for org.languagetool.languagemodel.LanguageModel.
type LanguageModel interface {
	GetPseudoProbability(tokens []string) Probability
}

// FuncLanguageModel adapts a function to LanguageModel.
type FuncLanguageModel func(tokens []string) Probability

func (f FuncLanguageModel) GetPseudoProbability(tokens []string) Probability {
	return f(tokens)
}

// UniformLanguageModel always returns the given probability (for tests).
func UniformLanguageModel(prob float64, coverage float32) LanguageModel {
	return FuncLanguageModel(func([]string) Probability {
		return NewProbabilitySimple(prob, coverage)
	})
}
