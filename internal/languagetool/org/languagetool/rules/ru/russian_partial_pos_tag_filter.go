package ru

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianPartialPosTagFilter ports org.languagetool.rules.ru.RussianPartialPosTagFilter
// (PartialPosTagFilter that tags and disambiguates a single token in Java).
//
// Java: Russian.getInstance().getTagger() then getDisambiguator() on a one-token
// AnalyzedSentence. Without both process-wide hooks, Accept fails closed — no invent.
type RussianPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewRussianPartialPosTagFilter(tag func(string) []string) *RussianPartialPosTagFilter {
	if tag == nil {
		tag = russianPartialTagAndDisambiguatePOS
	}
	return &RussianPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

// russianPartialTagAndDisambiguatePOS ports RussianPartialPosTagFilter.tag.
func russianPartialTagAndDisambiguatePOS(partial string) []string {
	tag := getDefaultRussianPartialPosTagger()
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
		return collectPOSTagsRU(atr)
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

func collectPOSTagsRU(atr *languagetool.AnalyzedTokenReadings) []string {
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

// NoDisambiguationRussianPartialPosTagFilter ports
// org.languagetool.rules.ru.NoDisambiguationRussianPartialPosTagFilter
// (PartialPosTagFilter + Russian tagger only, no disambiguator).
// When tag is nil, uses the process-wide filter tagger from SetDefaultRussianPartialPosTagger.
// Without a wired tagger, Accept fails closed (do not invent POS).
type NoDisambiguationRussianPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewNoDisambiguationRussianPartialPosTagFilter(tag func(string) []string) *NoDisambiguationRussianPartialPosTagFilter {
	if tag == nil {
		tag = russianNoDisambigTagPOS
	}
	return &NoDisambiguationRussianPartialPosTagFilter{
		PartialPosTagFilter: rules.NewPartialPosTagFilter(tag),
	}
}

var (
	ruPartialTagMu sync.RWMutex
	ruPartialTag   func(string) []string

	ruFilterDisambigMu sync.RWMutex
	ruFilterDisambig   filterDisambiguator
)

type filterDisambiguator interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// SetDefaultRussianPartialPosTagger wires Russian tagger POS for NoDisambiguation filters.
// Java: Languages.getLanguageForShortCode("ru").getTagger().
func SetDefaultRussianPartialPosTagger(tag func(string) []string) {
	ruPartialTagMu.Lock()
	defer ruPartialTagMu.Unlock()
	ruPartialTag = tag
}

// ClearDefaultRussianPartialPosTagger clears the process-wide tagger (tests).
func ClearDefaultRussianPartialPosTagger() {
	SetDefaultRussianPartialPosTagger(nil)
}

func getDefaultRussianPartialPosTagger() func(string) []string {
	ruPartialTagMu.RLock()
	defer ruPartialTagMu.RUnlock()
	return ruPartialTag
}

// WireRussianFilterDisambiguator installs the RU hybrid for PartialPosTagFilter probes.
// Call from RegisterHybridDisambiguator when the hybrid is ready.
func WireRussianFilterDisambiguator(d filterDisambiguator) {
	ruFilterDisambigMu.Lock()
	defer ruFilterDisambigMu.Unlock()
	ruFilterDisambig = d
}

// ClearRussianFilterDisambiguator clears the process-wide filter disambiguator (tests).
func ClearRussianFilterDisambiguator() {
	WireRussianFilterDisambiguator(nil)
}

func getFilterDisambiguator() filterDisambiguator {
	ruFilterDisambigMu.RLock()
	defer ruFilterDisambigMu.RUnlock()
	return ruFilterDisambig
}

func russianNoDisambigTagPOS(partial string) []string {
	tag := getDefaultRussianPartialPosTagger()
	if tag == nil {
		return nil
	}
	return tag(partial)
}

// WireRussianFilterTaggerFromTagWord installs POS list hook from lt.TagWord.
func WireRussianFilterTaggerFromTagWord(tw func(token string) []languagetool.TokenTag) {
	if tw == nil {
		SetDefaultRussianPartialPosTagger(nil)
		return
	}
	SetDefaultRussianPartialPosTagger(func(token string) []string {
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
