package ga

import (
	"os"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ga"
	taggingga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// IrishPartialPosTagFilter ports org.languagetool.rules.ga.IrishPartialPosTagFilter
// (PartialPosTagFilter that tags and disambiguates a single token in Java).
//
// Java (IrishPartialPosTagFilter.tag):
//
//	var irishLanguage = Irish.getInstance();
//	List<AnalyzedTokenReadings> tags = irishLanguage.getTagger().tag(Collections.singletonList(token));
//	AnalyzedTokenReadings[] atr = tags.toArray(...);
//	AnalyzedSentence disambiguated = irishLanguage.getDisambiguator().disambiguate(new AnalyzedSentence(atr));
//	return Arrays.asList(disambiguated.getTokens());
//
// IOException → RuntimeException "Could not tag and disambiguate '…'".
//
// Warning (Java javadoc): single-token disambiguator only; do NOT use in
// disambiguation.xml (endless loop) — use NoDisambiguationIrishPartialPosTagFilter.
//
// Default path (nil constructor tag): full IrishTagger.Tag ATR (lemmas preserved)
// + Irish hybrid disambiguator (process-wide WireIrishFilterDisambiguator, else
// NewIrishHybridDisambiguator auto-wire). Fail-closed without tagger and without
// invent POS; nil Disambiguate does not fall back to pre-disambig POS.
type IrishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

// NewIrishPartialPosTagFilter constructs the filter.
// tag maps a partial token to POS tags (injected for tests); nil →
// irishPartialTagAndDisambiguatePOS (tagger + disambiguator).
func NewIrishPartialPosTagFilter(tag func(string) []string) *IrishPartialPosTagFilter {
	if tag == nil {
		tag = irishPartialTagAndDisambiguatePOS
	}
	return &IrishPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

// irishPartialTagAndDisambiguatePOS ports IrishPartialPosTagFilter.tag:
// tagger.tag([token]) → disambiguator.disambiguate(AnalyzedSentence) → POS list
// (PartialPosTagFilter.partialTagHasRequiredTag walks readings' POS tags).
func irishPartialTagAndDisambiguatePOS(partial string) []string {
	d := resolveIrishFilterDisambiguator()
	if d == nil {
		// Fail closed: Java always has Irish.getDisambiguator(); without a
		// disambiguator we must not invent disambiguated POS.
		return nil
	}
	atr := resolveIrishPartialATR(partial)
	if atr == nil {
		// Fail closed: no tagger / no invent POS.
		return nil
	}
	// Java: new AnalyzedSentence(atr) from tagger.tag(singletonList(token)).
	// Tagger always returns one ATR per input token (untagged when dict has no entry).
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{atr})
	out := d.Disambiguate(sent)
	if out == nil {
		// Java would NPE on null disambiguated; fail closed (no pre-disambig invent).
		return nil
	}
	// Java returns disambiguated.getTokens(); partialTagHasRequiredTag iterates every
	// reading with non-null POSTag. Skip pure structural SENT_* markers so they do
	// not false-match postag_regexp (single-token path normally has none).
	return collectPOSTagsFromDisambiguatedSentenceGA(out)
}

// resolveIrishPartialATR builds the single-token ATR Java feeds the disambiguator.
//
// Order:
//  1. Process-wide SetDefaultIrishPartialPosTagger / WireIrishFilterTagger* (tests /
//     explicit wires) — POS-list rebuild (lemmas may be nil when only POS was stored).
//  2. Official IrishTagger.Tag (Java Irish.getTagger().tag) — full ATR with lemmas.
//
// Fail-closed (nil) when neither path can produce an ATR.
func resolveIrishPartialATR(partial string) *languagetool.AnalyzedTokenReadings {
	if tag := getDefaultIrishPartialPosTagger(); tag != nil {
		return posListToATR(partial, tag(partial))
	}
	return tagATRFromDefaultIrishTagger(partial)
}

// tagATRFromDefaultIrishTagger ports Irish.getTagger().tag(singletonList(token)).get(0)
// with full readings (POS + lemmas). Nil when official irish.dict is not loaded.
func tagATRFromDefaultIrishTagger(partial string) *languagetool.AnalyzedTokenReadings {
	taggingga.EnsureDefaultIrishTagger()
	if taggingga.IrishPOSDictPath() == "" {
		return nil
	}
	tagger := taggingga.DefaultIrishTagger
	if tagger == nil {
		return nil
	}
	// Java: tagger.tag(Collections.singletonList(token)) — no SENT_START in that list.
	trs := tagger.Tag([]string{partial})
	if len(trs) == 0 {
		return nil
	}
	return trs[0]
}

// posListToATR rebuilds an ATR from POS strings (process-wide inject / wire path).
// Untagged when posList is empty (Java BaseTagger still returns one ATR).
func posListToATR(partial string, posList []string) *languagetool.AnalyzedTokenReadings {
	if len(posList) == 0 {
		return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(partial, nil, nil))
	}
	readings := make([]*languagetool.AnalyzedToken, 0, len(posList))
	for _, p := range posList {
		var pos *string
		if p != "" {
			pp := p
			pos = &pp
		}
		readings = append(readings, languagetool.NewAnalyzedToken(partial, pos, nil))
	}
	return languagetool.NewAnalyzedTokenReadingsList(readings, 0)
}

// resolveIrishFilterDisambiguator returns the process-wide wire, or auto-wires
// IrishHybridDisambiguator (Java Irish.getInstance().getDisambiguator()).
func resolveIrishFilterDisambiguator() filterDisambiguator {
	if d := getFilterDisambiguator(); d != nil {
		return d
	}
	return defaultIrishHybridDisambiguator()
}

var (
	gaDefaultHybridOnce sync.Once
	gaDefaultHybrid     filterDisambiguator
)

// defaultIrishHybridDisambiguator caches NewIrishHybridDisambiguator (multiwords +
// XML stages load from official resources when discoverable).
func defaultIrishHybridDisambiguator() filterDisambiguator {
	gaDefaultHybridOnce.Do(func() {
		gaDefaultHybrid = disambigga.NewIrishHybridDisambiguator()
	})
	return gaDefaultHybrid
}

// collectPOSTagsFromDisambiguatedSentenceGA flattens POS tags from
// AnalyzedSentence.getTokens() the way PartialPosTagFilter.partialTagHasRequiredTag
// inspects List<ATR> (mirrors EN collectPOSTagsFromSentence).
func collectPOSTagsFromDisambiguatedSentenceGA(sent *languagetool.AnalyzedSentence) []string {
	if sent == nil {
		return nil
	}
	var posTags []string
	for _, t := range sent.GetTokens() {
		if t == nil {
			continue
		}
		for _, r := range t.GetReadings() {
			if r == nil {
				continue
			}
			p := r.GetPOSTag()
			if p == nil || *p == "" {
				continue
			}
			switch *p {
			case languagetool.SentenceStartTagName,
				languagetool.SentenceEndTagName,
				languagetool.ParagraphEndTagName:
				continue
			}
			posTags = append(posTags, *p)
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
//
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
// IrishPartialPosTagFilter still auto-wires NewIrishHybridDisambiguator when the
// process-wide hook is nil (Java Irish.getDisambiguator() always exists).
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
	atr := tagATRFromDefaultIrishTagger(partial)
	if atr == nil {
		return nil
	}
	return collectPOSTagsGA(atr)
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
