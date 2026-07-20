package languagetool

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// PseudoProbability is a minimal score surface (avoids importing ngrams into this package).
type PseudoProbability interface {
	GetProb() float64
	GetCoverage() float32
}

// NgramModel is the minimal LM interface for LanguageWithModel.
// Java LanguageModel is AutoCloseable; Close is optional here.
type NgramModel interface {
	GetPseudoProbability(context []string) PseudoProbability
}

// ClosableNgramModel optionally closes (Java LanguageModel.close).
type ClosableNgramModel interface {
	NgramModel
	Close() error
}

// LanguageWithModel ports org.languagetool.LanguageWithModel as a holder
// for a lazily initialized ngram language model.
type LanguageWithModel struct {
	ShortCode string
	Name      string
	mu        sync.Mutex
	model     NgramModel
	// noLmWarningPrinted ports AtomicBoolean noLmWarningPrinted.
	noLmWarningPrinted atomic.Bool
	// InitModel creates a model for topIndexDir (indexDir/shortCode) when present.
	// When nil, default init only checks directory existence and warns once.
	InitModel func(topIndexDir string) (NgramModel, error)
}

func NewLanguageWithModel(shortCode, name string) *LanguageWithModel {
	return &LanguageWithModel{ShortCode: shortCode, Name: name}
}

func (l *LanguageWithModel) GetShortCode() string { return l.ShortCode }
func (l *LanguageWithModel) GetName() string      { return l.Name }

// GetLanguageModel ports getLanguageModel(File indexDir) with double-checked locking.
func (l *LanguageWithModel) GetLanguageModel(indexDir string) (NgramModel, error) {
	if l == nil {
		return nil, nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.model != nil {
		return l.model, nil
	}
	m, err := l.initLanguageModel(indexDir)
	if err != nil {
		return nil, err
	}
	l.model = m
	return l.model, nil
}

// initLanguageModel ports protected initLanguageModel(File indexDir, LanguageModel languageModel).
func (l *LanguageWithModel) initLanguageModel(indexDir string) (NgramModel, error) {
	topIndexDir := filepath.Join(indexDir, l.ShortCode)
	if st, err := os.Stat(topIndexDir); err == nil && st.IsDir() {
		if l.InitModel != nil {
			return l.InitModel(topIndexDir)
		}
		// Without Lucene twin, return nil model when dir exists but no InitModel
		// (Java would construct LuceneLanguageModel).
		return nil, nil
	}
	if l.noLmWarningPrinted.CompareAndSwap(false, true) {
		// Java: System.err.println("WARN: ngram index dir " + topIndexDir + " not found for " + getName());
		fmt.Fprintf(os.Stderr, "WARN: ngram index dir %s not found for %s\n", topIndexDir, l.Name)
	}
	return nil, nil
}

// Close ports close() — closes model if ClosableNgramModel.
func (l *LanguageWithModel) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.model != nil {
		if c, ok := l.model.(ClosableNgramModel); ok {
			err := c.Close()
			l.model = nil
			return err
		}
		l.model = nil
	}
	return nil
}
