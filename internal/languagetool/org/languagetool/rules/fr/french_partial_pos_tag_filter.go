package fr

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FrenchPartialPosTagFilter ports org.languagetool.rules.fr.FrenchPartialPosTagFilter
// (PartialPosTagFilter that tags and disambiguates a single token in Java).
//
// Java: French.getInstance().getTagger() then getDisambiguator() on a one-token
// AnalyzedSentence. Without both process-wide hooks, Accept fails closed — no invent.
type FrenchPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewFrenchPartialPosTagFilter(tag func(string) []string) *FrenchPartialPosTagFilter {
	if tag == nil {
		tag = frenchPartialTagAndDisambiguatePOS
	}
	return &FrenchPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

func frenchPartialTagAndDisambiguatePOS(partial string) []string {
	tag := getDefaultFrenchPartialPosTagger()
	d := getFilterDisambiguator()
	if tag == nil || d == nil {
		return nil
	}
	posList := tag(partial)
	var readings []*languagetool.AnalyzedToken
	if len(posList) == 0 {
		readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(partial, nil, nil)}
	} else {
		readings = make([]*languagetool.AnalyzedToken, 0, len(posList))
		for _, p := range posList {
			pp := p
			readings = append(readings, languagetool.NewAnalyzedToken(partial, &pp, nil))
		}
	}
	atr := languagetool.NewAnalyzedTokenReadingsList(readings, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{atr})
	out := d.Disambiguate(sent)
	if out == nil {
		return collectPOSTagsFR(atr)
	}
	var posTags []string
	for _, t := range out.GetTokens() {
		if t == nil || t.GetToken() == "" {
			continue
		}
		for _, r := range t.GetReadings() {
			if r == nil {
				continue
			}
			if p := r.GetPOSTag(); p != nil && *p != "" &&
				*p != languagetool.SentenceStartTagName &&
				*p != languagetool.SentenceEndTagName &&
				*p != languagetool.ParagraphEndTagName {
				posTags = append(posTags, *p)
			}
		}
	}
	return posTags
}

func collectPOSTagsFR(atr *languagetool.AnalyzedTokenReadings) []string {
	if atr == nil {
		return nil
	}
	var out []string
	for _, r := range atr.GetReadings() {
		if r == nil {
			continue
		}
		if p := r.GetPOSTag(); p != nil && *p != "" {
			out = append(out, *p)
		}
	}
	return out
}

var (
	frPartialTagMu sync.RWMutex
	frPartialTag   func(string) []string

	frFilterDisambigMu sync.RWMutex
	frFilterDisambig   filterDisambiguator
)

type filterDisambiguator interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// SetDefaultFrenchPartialPosTagger wires French tagger POS for filter hooks.
// Java: French.getInstance().getTagger().
func SetDefaultFrenchPartialPosTagger(tag func(string) []string) {
	frPartialTagMu.Lock()
	defer frPartialTagMu.Unlock()
	frPartialTag = tag
}

// ClearDefaultFrenchPartialPosTagger clears the process-wide tagger (tests).
func ClearDefaultFrenchPartialPosTagger() {
	SetDefaultFrenchPartialPosTagger(nil)
}

func getDefaultFrenchPartialPosTagger() func(string) []string {
	frPartialTagMu.RLock()
	defer frPartialTagMu.RUnlock()
	return frPartialTag
}

// WireFrenchFilterDisambiguator installs the FR hybrid for PartialPosTagFilter probes.
func WireFrenchFilterDisambiguator(d filterDisambiguator) {
	frFilterDisambigMu.Lock()
	defer frFilterDisambigMu.Unlock()
	frFilterDisambig = d
}

// ClearFrenchFilterDisambiguator clears the process-wide filter disambiguator (tests).
func ClearFrenchFilterDisambiguator() {
	WireFrenchFilterDisambiguator(nil)
}

func getFilterDisambiguator() filterDisambiguator {
	frFilterDisambigMu.RLock()
	defer frFilterDisambigMu.RUnlock()
	return frFilterDisambig
}

// WireFrenchFilterTaggerFromTagWord installs POS list hook from lt.TagWord.
func WireFrenchFilterTaggerFromTagWord(tw func(token string) []languagetool.TokenTag) {
	if tw == nil {
		SetDefaultFrenchPartialPosTagger(nil)
		return
	}
	SetDefaultFrenchPartialPosTagger(func(token string) []string {
		tags := tw(token)
		if len(tags) == 0 {
			return nil
		}
		out := make([]string, 0, len(tags))
		for _, t := range tags {
			if t.POS != "" {
				out = append(out, t.POS)
			}
		}
		return out
	})
}
