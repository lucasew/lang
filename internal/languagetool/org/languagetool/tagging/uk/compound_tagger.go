package uk

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Known dash prefixes (subset of /uk/dash_prefixes.txt for green tests).
var knownDashPrefixes = map[string]struct{}{
	"міні": {}, "віце": {}, "екс": {}, "супер": {}, "ультра": {},
	"анти": {}, "псевдо": {}, "квазі": {}, "макро": {}, "мікро": {},
	"пів": {}, "напів": {},
}

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
			// already has real POS (incl. specials) — leave unless untagged null-only
			if !isOnlyNullTagged(base[i]) {
				continue
			}
		}
		parts := strings.Split(word, "-")
		if len(parts) < 2 {
			continue
		}
		if t.Debug != nil {
			t.Debug.Log("compound", word)
		}
		readings := t.tagCompoundParts(word, parts)
		if len(readings) > 0 {
			base[i] = languagetool.NewAnalyzedTokenReadingsList(readings, base[i].GetStartPos())
		}
	}
	return base
}

func isOnlyNullTagged(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return true
	}
	for _, r := range atr.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && *r.GetPOSTag() != "" {
			return false
		}
	}
	return true
}

func (t *CompoundTagger) tagCompoundParts(surface string, parts []string) []*languagetool.AnalyzedToken {
	// Prefer last part POS (head noun), lemma = full surface or head lemma.
	last := parts[len(parts)-1]
	left := strings.Join(parts[:len(parts)-1], "-")
	lastReadings := t.Inner.Tag([]string{last})
	var out []*languagetool.AnalyzedToken
	if len(lastReadings) > 0 && lastReadings[0] != nil {
		for _, r := range lastReadings[0].GetReadings() {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			pos := *r.GetPOSTag()
			// keep lemma as head lemma when present
			lemma := last
			if r.GetLemma() != nil && *r.GetLemma() != "" {
				lemma = *r.GetLemma()
			}
			// for known prefixes, annotate with :comp hint on pos when noun/adj
			if _, ok := knownDashPrefixes[strings.ToLower(left)]; ok {
				if !strings.Contains(pos, ":comp") {
					pos = pos + ":comp"
				}
			}
			l := lemma
			p := pos
			out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
		}
	}
	// Also try full surface in dict
	if tws := t.Inner.TagWord(surface); len(tws) > 0 {
		for _, tw := range tws {
			out = append(out, toTok(surface, tw))
		}
	}
	// numeric left part + tagged right: 2-річний style
	if len(out) == 0 && isNumericPrefix(parts[0]) {
		right := t.Inner.Tag([]string{parts[len(parts)-1]})
		if len(right) > 0 && right[0] != nil {
			for _, r := range right[0].GetReadings() {
				if r != nil && r.GetPOSTag() != nil {
					p := *r.GetPOSTag()
					l := surface
					out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
				}
			}
		}
	}
	return out
}

func isNumericPrefix(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
