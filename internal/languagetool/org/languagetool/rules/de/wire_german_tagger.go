package de

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
)

// Cached GermanTagger for process-wide filter hooks and JLanguageTool.TagWord
// (Java GermanyGerman.createDefaultTagger → GermanTagger.INSTANCE).
var (
	germanTaggerOnce sync.Once
	germanTagger     *detag.GermanTagger
)

// DiscoveredGermanTagger returns the process-wide GermanTagger when resources
// open (manual + optional german.dict). Nil if unavailable (fail-closed).
func DiscoveredGermanTagger() *detag.GermanTagger {
	germanTaggerOnce.Do(func() {
		germanTagger = openDiscoveredGermanTagger(DiscoverGermanResourceDir())
	})
	return germanTagger
}

// GermanTagWord ports GermanTagger lookup as JLanguageTool.TagWord inject.
// Returns POS+lemma readings; empty when unknown (no invent tags).
func GermanTagWord(tagger *detag.GermanTagger) func(token string) []languagetool.TokenTag {
	if tagger == nil {
		return nil
	}
	return func(token string) []languagetool.TokenTag {
		if token == "" {
			return nil
		}
		rd := tagger.Lookup(token)
		if rd == nil {
			return nil
		}
		var out []languagetool.TokenTag
		for _, r := range rd.GetReadings() {
			if r == nil {
				continue
			}
			tt := languagetool.TokenTag{}
			if p := r.GetPOSTag(); p != nil {
				tt.POS = *p
			}
			if l := r.GetLemma(); l != nil {
				tt.Lemma = *l
			}
			if tt.POS != "" || tt.Lemma != "" {
				out = append(out, tt)
			}
		}
		return out
	}
}

// WireGermanTagWord installs GermanTagger as lt.TagWord so Check/Analyze get
// morph tags (Java createDefaultTagger). Returns false if tagger unavailable.
func WireGermanTagWord(lt *languagetool.JLanguageTool) bool {
	return WireGermanTagWordFor(lt, "DE")
}

// WireGermanTagWordFor is WireGermanTagWord with language variant.
// Java SwissGerman.createDefaultTagger → SwissGermanTagger (ss→ß retry for untagged).
// Also sets SentenceTagger to GermanTagger.Tag (Java getTagger().tag(List) with ignoreCase).
func WireGermanTagWordFor(lt *languagetool.JLanguageTool, variant string) bool {
	if lt == nil {
		return false
	}
	base := DiscoveredGermanTagger()
	if base == nil {
		return false
	}
	if strings.EqualFold(variant, "CH") {
		// Wrap the same WordTagger stack with SwissGermanTagger.Lookup (ss→ß).
		wt := base.GetWordTagger()
		if wt == nil {
			return false
		}
		swiss := detag.NewSwissGermanTagger(wt)
		// Preserve spelling expansions / compound split / removal from base when present.
		if base.RemovalTagger != nil {
			swiss.RemovalTagger = base.RemovalTagger
		}
		if base.SplitCompound != nil {
			swiss.SplitCompound = base.SplitCompound
		}
		if base.AdjExpansion != nil {
			swiss.SetSpellingAdjExpansion(base.AdjExpansion)
		}
		if base.VerbExpansion != nil {
			swiss.SetSpellingVerbExpansion(base.VerbExpansion)
		}
		tw := SwissGermanTagWord(swiss)
		if tw == nil {
			return false
		}
		lt.TagWord = tw
		// Java createDefaultTagger().tag(tokens) — full sentence (ss→ß via SwissGermanTagger.Tag).
		lt.SentenceTagger = func(tokens []string) []*languagetool.AnalyzedTokenReadings {
			return swiss.Tag(tokens)
		}
		return true
	}
	tw := GermanTagWord(base)
	if tw == nil {
		return false
	}
	lt.TagWord = tw
	// Java GermanyGerman.createDefaultTagger → GermanTagger.INSTANCE.tag(tokens, ignoreCase=true).
	lt.SentenceTagger = func(tokens []string) []*languagetool.AnalyzedTokenReadings {
		return base.Tag(tokens)
	}
	return true
}

// SwissGermanTagWord ports SwissGermanTagger.Lookup as JLanguageTool.TagWord inject.
func SwissGermanTagWord(tagger *detag.SwissGermanTagger) func(token string) []languagetool.TokenTag {
	if tagger == nil {
		return nil
	}
	return func(token string) []languagetool.TokenTag {
		if token == "" {
			return nil
		}
		rd := tagger.Lookup(token)
		if rd == nil {
			return nil
		}
		var out []languagetool.TokenTag
		for _, r := range rd.GetReadings() {
			if r == nil {
				continue
			}
			tt := languagetool.TokenTag{}
			if p := r.GetPOSTag(); p != nil {
				tt.POS = *p
			}
			if l := r.GetLemma(); l != nil {
				tt.Lemma = *l
			}
			if tt.POS != "" || tt.Lemma != "" {
				out = append(out, tt)
			}
		}
		return out
	}
}
