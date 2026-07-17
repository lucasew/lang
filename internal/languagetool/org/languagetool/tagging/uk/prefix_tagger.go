package uk

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// noDashPrefixes ports a subset of no-dash prefix tagging (пів/напів/...).
var noDashPrefixes = []string{
	"напів", "пів", "анти", "псевдо", "квазі", "супер", "ультра", "макро", "мікро",
}

// TryNoDashPrefixTags attempts to tag word by stripping a known prefix and using
// the right-hand head readings from tagRight (inner tagger for the remainder).
func TryNoDashPrefixTags(surface string, tagRight func(string) []*languagetool.AnalyzedToken) []*languagetool.AnalyzedToken {
	lower := strings.ToLower(surface)
	for _, prefix := range noDashPrefixes {
		if !strings.HasPrefix(lower, prefix) || len(lower) <= len(prefix) {
			continue
		}
		// don't split if prefix ends with vowel and remainder starts with soft mark oddly — green skip
		right := surface[lenPrefixBytes(surface, prefix):]
		if right == "" {
			continue
		}
		// require remainder starts with letter
		r, _ := utf8.DecodeRuneInString(right)
		if !unicode.IsLetter(r) {
			continue
		}
		heads := tagRight(right)
		if len(heads) == 0 {
			continue
		}
		var out []*languagetool.AnalyzedToken
		for _, h := range heads {
			if h == nil || h.GetPOSTag() == nil {
				continue
			}
			pos := *h.GetPOSTag()
			if !strings.Contains(pos, ":alt") && prefix == "пів" {
				// пів- compounds often :alt in full system; mark lightly
			}
			lemma := surface
			if h.GetLemma() != nil && *h.GetLemma() != "" {
				// keep head lemma with prefix for green
				lemma = prefix + *h.GetLemma()
			}
			p := pos
			l := lemma
			out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

func lenPrefixBytes(surface, prefixLower string) int {
	// prefix matched on lower; map back to rune count of prefix on original
	pr := []rune(prefixLower)
	sr := []rune(surface)
	if len(pr) > len(sr) {
		return 0
	}
	return len(string(sr[:len(pr)]))
}
