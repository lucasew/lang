package ga

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// IrishPartialPosTagFilter ports org.languagetool.rules.ga.IrishPartialPosTagFilter
// (tags + disambiguates a single token). Without both process-wide hooks, fail-closed.
type IrishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewIrishPartialPosTagFilter(tag func(string) []string) *IrishPartialPosTagFilter {
	if tag == nil {
		tag = irishPartialTagAndDisambiguatePOS
	}
	return &IrishPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

func irishPartialTagAndDisambiguatePOS(partial string) []string {
	tag := getDefaultIrishPartialPosTagger()
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
		return collectPOSTagsGA(atr)
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

func collectPOSTagsGA(atr *languagetool.AnalyzedTokenReadings) []string {
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

// NoDisambiguationIrishPartialPosTagFilter ports
// org.languagetool.rules.ga.NoDisambiguationIrishPartialPosTagFilter
// (tagger only). When tag is nil, uses process-wide SetDefaultIrishPartialPosTagger.
type NoDisambiguationIrishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewNoDisambiguationIrishPartialPosTagFilter(tag func(string) []string) *NoDisambiguationIrishPartialPosTagFilter {
	if tag == nil {
		tag = irishNoDisambigTagPOS
	}
	return &NoDisambiguationIrishPartialPosTagFilter{
		PartialPosTagFilter: rules.NewPartialPosTagFilter(tag),
	}
}

var (
	gaPartialTagMu sync.RWMutex
	gaPartialTag   func(string) []string

	gaFilterDisambigMu sync.RWMutex
	gaFilterDisambig   filterDisambiguator
)

type filterDisambiguator interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// SetDefaultIrishPartialPosTagger wires Irish tagger POS (Java Irish.getTagger()).
func SetDefaultIrishPartialPosTagger(tag func(string) []string) {
	gaPartialTagMu.Lock()
	defer gaPartialTagMu.Unlock()
	gaPartialTag = tag
}

// ClearDefaultIrishPartialPosTagger clears the process-wide tagger (tests).
func ClearDefaultIrishPartialPosTagger() {
	SetDefaultIrishPartialPosTagger(nil)
}

func getDefaultIrishPartialPosTagger() func(string) []string {
	gaPartialTagMu.RLock()
	defer gaPartialTagMu.RUnlock()
	return gaPartialTag
}

// WireIrishFilterDisambiguator installs the GA hybrid for PartialPosTagFilter probes.
func WireIrishFilterDisambiguator(d filterDisambiguator) {
	gaFilterDisambigMu.Lock()
	defer gaFilterDisambigMu.Unlock()
	gaFilterDisambig = d
}

// ClearIrishFilterDisambiguator clears the process-wide filter disambiguator (tests).
func ClearIrishFilterDisambiguator() {
	WireIrishFilterDisambiguator(nil)
}

func getFilterDisambiguator() filterDisambiguator {
	gaFilterDisambigMu.RLock()
	defer gaFilterDisambigMu.RUnlock()
	return gaFilterDisambig
}

func irishNoDisambigTagPOS(partial string) []string {
	tag := getDefaultIrishPartialPosTagger()
	if tag == nil {
		return nil
	}
	return tag(partial)
}

// WireIrishFilterTaggerFromTagWord installs POS list hook from lt.TagWord.
func WireIrishFilterTaggerFromTagWord(tw func(token string) []languagetool.TokenTag) {
	if tw == nil {
		SetDefaultIrishPartialPosTagger(nil)
		return
	}
	SetDefaultIrishPartialPosTagger(func(token string) []string {
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
