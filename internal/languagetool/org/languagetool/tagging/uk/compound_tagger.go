package uk

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// CompoundTagger ports tagging.uk.CompoundTagger: tags hyphenated compounds via parts.
type CompoundTagger struct {
	Inner *UkrainianTagger
	Debug *CompoundDebugLogger
}

func NewCompoundTagger(inner *UkrainianTagger) *CompoundTagger {
	return &CompoundTagger{Inner: inner, Debug: NewCompoundDebugLogger(false)}
}

func (t *CompoundTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil || t.Inner == nil {
		return nil
	}
	base := t.Inner.Tag(sentenceTokens)
	for i, word := range sentenceTokens {
		if i >= len(base) || !strings.Contains(word, "-") {
			continue
		}
		if base[i] != nil && base[i].IsTagged() {
			continue
		}
		parts := strings.Split(word, "-")
		if len(parts) < 2 {
			continue
		}
		if t.Debug != nil {
			t.Debug.Log("compound", word)
		}
		last := t.Inner.Tag([]string{parts[len(parts)-1]})
		if len(last) > 0 && last[0] != nil {
			base[i] = last[0]
		}
	}
	return base
}
