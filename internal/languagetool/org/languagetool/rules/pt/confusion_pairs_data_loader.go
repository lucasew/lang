package pt

import (
	"io"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ConfusionPairsDataLoader ports org.languagetool.rules.pt.ConfusionPairsDataLoader.
// File lines: unaccented;accented;POS (3 semicolon-separated fields).
type ConfusionPairsDataLoader struct{}

func (ConfusionPairsDataLoader) LoadWords(r io.Reader, path string) (map[string]*languagetool.AnalyzedTokenReadings, error) {
	// false = do not reverse columns (parts[0] key, parts[1] surface, parts[2] POS)
	return rules.NewAccentuationDataLoader(false).LoadWords(r, path)
}
