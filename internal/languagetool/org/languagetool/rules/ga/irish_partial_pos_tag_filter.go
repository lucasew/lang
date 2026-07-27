package ga

import (
	"os"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taggingga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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
// (PartialPosTagFilter + IrishTagger only, no disambiguator).
//
// Java: Irish.getInstance().getTagger().tag(Collections.singletonList(token))
// — no getDisambiguator(). When tag is nil:
//  1. process-wide SetDefaultIrishPartialPosTagger / WireIrishFilterTagger*, else
//  2. official IrishTagger via EnsureDefaultIrishTagger (irish.dict).
// Without dict, Accept fails closed (no invent POS).
type NoDisambiguationIrishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

// NewNoDisambiguationIrishPartialPosTagFilter constructs the filter.
// tag maps a partial token to POS tags (injected for tests); nil → irishNoDisambigTagPOS.
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
// After clear, NoDisambiguation falls back to IrishTagger when irish.dict is discoverable.
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

// irishNoDisambigTagPOS ports NoDisambiguationIrishPartialPosTagFilter.tag
// (Java: Irish.getInstance().getTagger().tag(singletonList(token)) — no disambiguator).
func irishNoDisambigTagPOS(partial string) []string {
	if tag := getDefaultIrishPartialPosTagger(); tag != nil {
		return tag(partial)
	}
	return tagPOSFromDefaultIrishTagger(partial)
}

// tagPOSFromDefaultIrishTagger collects POS strings from IrishTagger.Tag on one token.
// Fail-closed when official irish.dict is not loaded (no invent POS).
func tagPOSFromDefaultIrishTagger(partial string) []string {
	taggingga.EnsureDefaultIrishTagger()
	if taggingga.IrishPOSDictPath() == "" {
		return nil
	}
	tagger := taggingga.DefaultIrishTagger
	if tagger == nil {
		return nil
	}
	// Java: tagger.tag(Collections.singletonList(token))
	trs := tagger.Tag([]string{partial})
	if len(trs) == 0 {
		return nil
	}
	var out []string
	for _, atr := range trs {
		out = append(out, collectPOSTagsGA(atr)...)
	}
	return out
}

// WireIrishFilterTagger wires process-wide POS from official IrishTagger (irish.dict).
// Empty dictPath discovers via tagging/ga (LANG_IRISH_DICT / LANG_GA_DICT / /tmp / third_party).
// Non-empty dictPath must be a readable file; discovery still loads via EnsureDefaultIrishTagger.
// Returns false when irish.dict cannot be loaded (fail-closed).
// Mirrors EN WireEnglishFilterTagger for process-wide default.
func WireIrishFilterTagger(dictPath string) bool {
	if tools.JavaStringTrim(dictPath) != "" {
		st, err := os.Stat(dictPath)
		if err != nil || !st.Mode().IsRegular() {
			return false
		}
	}
	taggingga.EnsureDefaultIrishTagger()
	if taggingga.IrishPOSDictPath() == "" {
		return false
	}
	SetDefaultIrishPartialPosTagger(func(token string) []string {
		return tagPOSFromDefaultIrishTagger(token)
	})
	return true
}

// TryWireIrishFilterTagger discovers irish.dict and wires IrishTagger when present.
func TryWireIrishFilterTagger() bool {
	return WireIrishFilterTagger("")
}

// WireIrishFilterTaggerFromTagWord installs POS list hook from lt.TagWord.
// Prefer WireIrishFilterTagger / IrishTagger.Tag for faithful cascade (filterMorph).
// Nil tw clears the process-wide hook (NoDisambiguation then falls back to IrishTagger).
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
