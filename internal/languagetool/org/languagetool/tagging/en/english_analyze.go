package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
)

// AnalyzeEnglishSentence ports EN getAnalyzedSentence for faithful speller/tagger twins:
// EnglishWordTokenizer + EnglishTagger.tag (english.dict, typographic ’→') +
// EnglishHybridDisambiguator multiword IGNORE_SPELLING.
// XML disambiguator is optional (not required for speller).
func AnalyzeEnglishSentence(text string) *languagetool.AnalyzedSentence {
	EnsureDefaultEnglishTagger() // wires IsTaggedEN for tokenizer apostrophe keep
	wt := entok.NewEnglishWordTokenizer()
	tokens := wt.Tokenize(text)
	// Java: language.getTagger().tag(tokens) — EnglishTagger rewrites ’→' on surfaces.
	var sent *languagetool.AnalyzedSentence
	if DefaultEnglishTagger != nil && enPOSDictPath != "" {
		tagged := DefaultEnglishTagger.Tag(tokens)
		// Prepend SENT_START like other Analyze* paths.
		ss := languagetool.SentenceStartTagName
		startTok := languagetool.NewAnalyzedToken("", &ss, nil)
		startR := languagetool.NewAnalyzedTokenReadings(startTok)
		startR.SetStartPos(0)
		all := make([]*languagetool.AnalyzedTokenReadings, 0, len(tagged)+1)
		all = append(all, startR)
		// Tag() already set start positions from 0 relative to token list (no SENT_START).
		// Rebuild with absolute positions matching sentence text (same as BuildPositions).
		pos := 0
		prev := ""
		for _, atr := range tagged {
			if atr == nil {
				continue
			}
			// Clone-ish: re-position ATR while keeping readings + typographic flag.
			surface := atr.GetToken()
			// Prefer Tag() surface (already ’→' when typographic was present).
			readings := atr.GetReadings()
			var ar *languagetool.AnalyzedTokenReadings
			if len(readings) == 0 {
				at := languagetool.NewAnalyzedToken(surface, nil, nil)
				ar = languagetool.NewAnalyzedTokenReadingsAt(at, pos)
			} else {
				ar = languagetool.NewAnalyzedTokenReadingsAt(readings[0], pos)
				for _, r := range readings[1:] {
					if r != nil {
						ar.AddReading(r, "")
					}
				}
			}
			if atr.HasTypographicApostrophe() {
				ar.SetTypographicApostrophe(true)
			}
			if prev != "" {
				ar.SetWhitespaceBeforeToken(prev)
			}
			all = append(all, ar)
			prev = surface
			pos += utf16LenEN(surface)
		}
		sent = languagetool.NewAnalyzedSentence(all)
	} else {
		sent = languagetool.AnalyzeWithTokenizer(text, wt)
	}
	if d := DefaultEnglishHybridDisambiguator(); d != nil {
		sent = d.Disambiguate(sent)
	}
	return sent
}

func utf16LenEN(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}
