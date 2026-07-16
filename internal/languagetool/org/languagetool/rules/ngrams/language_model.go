package ngrams

// LanguageModel is the local alias used by ngram rules.
// Canonical package is org/languagetool/languagemodel; this interface stays
// here so ngram rules do not force every caller through that package path.
// Prefer implementing GetPseudoProbability only; Close is optional for map models.
type LanguageModel interface {
	GetPseudoProbability(tokens []string) Probability
}

// FuncLanguageModel adapts a function to LanguageModel.
type FuncLanguageModel func(tokens []string) Probability

func (f FuncLanguageModel) GetPseudoProbability(tokens []string) Probability { return f(tokens) }

// UniformLanguageModel always returns the given probability (for tests).
func UniformLanguageModel(prob float64, coverage float32) LanguageModel {
	return FuncLanguageModel(func([]string) Probability {
		return NewProbabilitySimple(prob, coverage)
	})
}

// AdaptLanguageModel wraps a model that also has Close (languagemodel package).
type closableLM interface {
	GetPseudoProbability(tokens []string) Probability
}

// AsNgramLM returns lm as ngrams.LanguageModel.
func AsNgramLM(lm closableLM) LanguageModel {
	return FuncLanguageModel(lm.GetPseudoProbability)
}
