package ca

import (
	"io"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ConfusionPairsDataLoader ports org.languagetool.rules.ca.ConfusionPairsDataLoader.
type ConfusionPairsDataLoader struct{}

func (ConfusionPairsDataLoader) LoadWords(r io.Reader, path string) (map[string]*languagetool.AnalyzedTokenReadings, error) {
	return rules.NewAccentuationDataLoader(true).LoadWords(r, path)
}
