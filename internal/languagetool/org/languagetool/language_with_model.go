package languagetool

import "sync"

// PseudoProbability is a minimal score surface (avoids importing ngrams into this package).
type PseudoProbability interface {
	GetProb() float64
	GetCoverage() float32
}

// NgramModel is the minimal LM interface for LanguageWithModel.
type NgramModel interface {
	GetPseudoProbability(context []string) PseudoProbability
}

// LanguageWithModel ports org.languagetool.LanguageWithModel as a holder
// for a lazily initialized ngram language model.
type LanguageWithModel struct {
	ShortCode string
	Name      string
	mu        sync.Mutex
	model     NgramModel
	// InitModel creates a model for indexDir when missing.
	InitModel func(indexDir string) (NgramModel, error)
}

func NewLanguageWithModel(shortCode, name string) *LanguageWithModel {
	return &LanguageWithModel{ShortCode: shortCode, Name: name}
}

func (l *LanguageWithModel) GetShortCode() string { return l.ShortCode }
func (l *LanguageWithModel) GetName() string      { return l.Name }

func (l *LanguageWithModel) GetLanguageModel(indexDir string) (NgramModel, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.model != nil {
		return l.model, nil
	}
	if l.InitModel == nil {
		return nil, nil
	}
	m, err := l.InitModel(indexDir)
	if err != nil {
		return nil, err
	}
	l.model = m
	return l.model, nil
}

func (l *LanguageWithModel) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.model = nil
	return nil
}
