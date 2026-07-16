package pt

import (
	"io"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseAccentuationDataLoader ports org.languagetool.rules.pt.PortugueseAccentuationDataLoader.
type PortugueseAccentuationDataLoader struct{}

func (PortugueseAccentuationDataLoader) LoadWords(r io.Reader, path string) (map[string]*languagetool.AnalyzedTokenReadings, error) {
	return rules.NewAccentuationDataLoader(false).LoadWords(r, path)
}
