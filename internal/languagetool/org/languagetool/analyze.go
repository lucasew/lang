package languagetool

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	betok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/be"
	brtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/br"
	catok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ca"
	crhtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/crh"
	detok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
	eltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/el"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	eotok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/eo"
	estok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/es"
	frtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/fr"
	gltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/gl"
	jatok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ja"
	kmtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/km"
	mltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ml"
	nltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/nl"
	pltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pl"
	pttok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pt"
	rotok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ro"
	rutok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ru"
	tltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/tl"
	uktok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/uk"
	zhtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/zh"
)

// AnalyzePlain ports a minimal getAnalyzedSentence for demo/rule unit tests:
// SENT_START + WordTokenizer tokens as untagged AnalyzedTokenReadings with start positions.
// Sets whitespaceBefore from the previous raw token (JLanguageTool analysis loop).
func AnalyzePlain(text string) *AnalyzedSentence {
	return AnalyzeWithTokenizer(text, tokenizers.NewWordTokenizer())
}

// AnalyzeWithTokenizer is AnalyzePlain with an explicit word tokenizer (e.g. FrenchWordTokenizer).
// JapaneseWordTokenizer follows Java: tokenize emits "surface POS lemma", then
// asAnalyzedToken splits into surface/POS/lemma (JapaneseTagger).
func AnalyzeWithTokenizer(text string, wt tokenizers.Tokenizer) *AnalyzedSentence {
	if wt == nil {
		wt = tokenizers.NewWordTokenizer()
	}
	if _, ok := wt.(*jatok.JapaneseWordTokenizer); ok {
		return analyzeJapaneseEncoded(wt.Tokenize(text))
	}
	raw := wt.Tokenize(text)
	positions := tokenizers.BuildPositions(raw)
	// tokens: SENT_START at 0, then each raw token
	readings := make([]*AnalyzedTokenReadings, 0, len(raw)+1)
	ss := SentenceStartTagName
	startTok := NewAnalyzedToken("", &ss, nil)
	startR := NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)
	readings = append(readings, startR)
	prev := ""
	for i, tok := range raw {
		at := NewAnalyzedToken(tok, nil, nil)
		ar := NewAnalyzedTokenReadingsAt(at, positions[i])
		if prev != "" {
			ar.SetWhitespaceBeforeToken(prev)
		}
		readings = append(readings, ar)
		prev = tok
	}
	// Soft: mirror LT analysis by tagging the last content token with SENT_END
	// (POINT_DIALOGUE and other rules match postag SENT_END on the final word).
	softAttachSentenceEnd(readings)
	return NewAnalyzedSentence(readings)
}

// analyzeJapaneseEncoded ports JLanguageTool analysis for Japanese:
// wordTokenizer.tokenize → "surface POS lemma"; tagger.asAnalyzedToken splits.
func analyzeJapaneseEncoded(encoded []string) *AnalyzedSentence {
	readings := make([]*AnalyzedTokenReadings, 0, len(encoded)+1)
	ss := SentenceStartTagName
	startTok := NewAnalyzedToken("", &ss, nil)
	startR := NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)
	readings = append(readings, startR)
	pos := 0
	prev := ""
	for _, word := range encoded {
		at := japaneseAsAnalyzedToken(word)
		ar := NewAnalyzedTokenReadingsAt(at, pos)
		if prev != "" {
			ar.SetWhitespaceBeforeToken(prev)
		}
		readings = append(readings, ar)
		pos += tokenizers.UTF16Len(at.GetToken())
		prev = at.GetToken()
	}
	softAttachSentenceEnd(readings)
	return NewAnalyzedSentence(readings)
}

// japaneseAsAnalyzedToken ports JapaneseTagger.asAnalyzedToken.
func japaneseAsAnalyzedToken(word string) *AnalyzedToken {
	parts := strings.Split(word, " ")
	if len(parts) != 3 {
		return NewAnalyzedToken(" ", nil, nil)
	}
	p, l := parts[1], parts[2]
	return NewAnalyzedToken(parts[0], &p, &l)
}

// softAttachSentenceEnd adds a SENT_END reading on the last non-SENT_START token.
func softAttachSentenceEnd(readings []*AnalyzedTokenReadings) {
	if len(readings) < 2 {
		return
	}
	last := readings[len(readings)-1]
	if last == nil || last.IsSentenceStart() {
		return
	}
	if last.HasPosTag(SentenceEndTagName) {
		return
	}
	se := SentenceEndTagName
	last.AddReading(NewAnalyzedToken(last.GetToken(), &se, nil), "")
}

// WordTokenizerForLanguage returns the language-specific soft word tokenizer.
// Falls back to the generic WordTokenizer when no language module is available.
func WordTokenizerForLanguage(lang string) tokenizers.Tokenizer {
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	switch strings.ToLower(base) {
	case "ar", "fa":
		// Arabic-script soft path; Persian uses the same digit/letter splits.
		if strings.EqualFold(base, "fa") {
			return tokenizers.NewPersianWordTokenizer()
		}
		return tokenizers.NewArabicWordTokenizer()
	case "be":
		return betok.NewBelarusianWordTokenizer()
	case "br":
		return brtok.NewBretonWordTokenizer()
	case "ca":
		return catok.NewCatalanWordTokenizer()
	case "crh":
		return crhtok.NewCrimeanTatarWordTokenizer()
	case "de":
		return detok.NewGermanWordTokenizer()
	case "el":
		return eltok.NewGreekWordTokenizer()
	case "en":
		return entok.NewEnglishWordTokenizer()
	case "eo":
		return eotok.NewEsperantoWordTokenizer()
	case "es":
		return estok.NewSpanishWordTokenizer()
	case "fr":
		return frtok.NewFrenchWordTokenizer()
	case "gl":
		return gltok.NewGalicianWordTokenizer()
	case "ja":
		return jatok.NewJapaneseWordTokenizer()
	case "km":
		return kmtok.NewKhmerWordTokenizer()
	case "ml":
		return mltok.NewMalayalamWordTokenizer()
	case "nl":
		return nltok.NewDutchWordTokenizer()
	case "pl":
		return pltok.NewPolishWordTokenizer()
	case "pt":
		return pttok.NewPortugueseWordTokenizer()
	case "ro":
		return rotok.NewRomanianWordTokenizer()
	case "ru":
		return rutok.NewRussianWordTokenizer()
	case "tl":
		return tltok.NewTagalogWordTokenizer()
	case "uk":
		return uktok.NewUkrainianWordTokenizer()
	case "zh":
		return zhtok.NewChineseWordTokenizer()
	default:
		return tokenizers.NewWordTokenizer()
	}
}

// AnalyzePlainStripSoftHyphen is AnalyzePlain after removing U+00AD (LT ignored chars).
func AnalyzePlainStripSoftHyphen(text string) *AnalyzedSentence {
	return AnalyzePlain(strings.ReplaceAll(text, "\u00AD", ""))
}

// CheckWhitespaceOnly runs MultipleWhitespace-style single-sentence check via callback.
// Kept in languagetool package for test helpers.
func AnalyzeSentences(text string) []*AnalyzedSentence {
	// single sentence for unit tests
	return []*AnalyzedSentence{AnalyzePlain(text)}
}

// SplitAndAnalyze splits on .!? boundaries for SentenceWhitespaceRule unit tests.
// Trailing single space after terminator is attached to the previous sentence
// (so prevSentenceEndsWithWhitespace matches LT SRX-ish behavior for these tests).
// Periods inside URLs/domains (e.g. example.com) are not treated as boundaries.
func SplitAndAnalyze(text string) []*AnalyzedSentence {
	if text == "" {
		return nil
	}
	var parts []string
	start := 0
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '.' || r == '!' || r == '?' {
			// Do not split on '.' when next is lowercase letter or digit (domain/decimal).
			// Uppercase after '.' without space is still a sentence boundary
			// (SentenceWhitespaceRule "text.And next").
			if r == '.' && i+1 < len(runes) {
				n := runes[i+1]
				if (n >= 'a' && n <= 'z') || (n >= '0' && n <= '9') {
					continue
				}
			}
			end := i + 1
			// include following single space/newline as part of this sentence
			if end < len(runes) && (runes[end] == ' ' || runes[end] == '\n' || runes[end] == '\u00A0') {
				// only one whitespace for "ends with whitespace" check
				if runes[end] == '\n' && end+1 < len(runes) && runes[end+1] == '\n' {
					// paragraph break: include first newline only? good tests have \n between sentences
					end++
					// if double newline, include second as well for "\n\n" good case
					if end < len(runes) && runes[end] == '\n' {
						end++
					}
				} else if runes[end] == ' ' || runes[end] == '\u00A0' {
					end++
				} else if runes[end] == '\n' {
					end++
				}
			}
			parts = append(parts, string(runes[start:end]))
			start = end
			i = end - 1
		}
	}
	if start < len(runes) {
		parts = append(parts, string(runes[start:]))
	}
	out := make([]*AnalyzedSentence, 0, len(parts))
	offset := 0
	for _, p := range parts {
		if p == "" {
			continue
		}
		s := AnalyzePlain(p)
		// shift token positions by offset for multi-sentence
		if offset > 0 {
			shiftSentence(s, offset)
		}
		out = append(out, s)
		// offset by UTF-16 length of part
		for _, r := range p {
			if r >= 0x10000 {
				offset += 2
			} else {
				offset++
			}
		}
	}
	return out
}

func shiftSentence(s *AnalyzedSentence, delta int) {
	for _, t := range s.GetTokens() {
		t.SetStartPos(t.GetStartPos() + delta)
	}
}

// AnalyzeTextDemo splits text into sentences for Demo-like unit tests.
// Paragraph boundaries: blank lines (\n\n). Sentence-local token positions
// (as LT does); TextLevelRule.match accumulates pos across sentences.
func AnalyzeTextDemo(text string) []*AnalyzedSentence {
	paras := strings.Split(text, "\n\n")
	var out []*AnalyzedSentence
	for pi, para := range paras {
		chunk := para
		var sents []*AnalyzedSentence
		if strings.Contains(chunk, ". ") || strings.Contains(chunk, ".\n") || strings.Contains(chunk, "! ") || strings.Contains(chunk, "? ") {
			sents = SplitAndAnalyze(chunk)
		} else if chunk != "" {
			sents = []*AnalyzedSentence{AnalyzePlain(chunk)}
		}
		if pi < len(paras)-1 && len(sents) > 0 {
			// Ensure last sentence of paragraph ends with \n\n for isParagraphEnd
			if len(sents) == 1 {
				sents = []*AnalyzedSentence{AnalyzePlain(chunk + "\n\n")}
			} else {
				sents = SplitAndAnalyze(chunk + "\n\n")
			}
		}
		out = append(out, sents...)
	}
	if len(out) == 0 && text != "" {
		return []*AnalyzedSentence{AnalyzePlain(text)}
	}
	return out
}

// AnalyzeTextLocal splits on .!? like SplitAndAnalyze but keeps sentence-local
// token positions (TextLevelRule accumulates GetCorrectedTextLength).
func AnalyzeTextLocal(text string) []*AnalyzedSentence {
	if text == "" {
		return nil
	}
	// Reuse SplitAndAnalyze structure without offset shift:
	var parts []string
	start := 0
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '.' || r == '!' || r == '?' {
			if r == '.' && i+1 < len(runes) {
				n := runes[i+1]
				if (n >= 'a' && n <= 'z') || (n >= '0' && n <= '9') {
					continue
				}
			}
			end := i + 1
			if end < len(runes) && (runes[end] == ' ' || runes[end] == '\n' || runes[end] == '\u00A0') {
				if runes[end] == '\n' && end+1 < len(runes) && runes[end+1] == '\n' {
					end++
					if end < len(runes) && runes[end] == '\n' {
						end++
					}
				} else if runes[end] == ' ' || runes[end] == '\u00A0' {
					end++
				} else if runes[end] == '\n' {
					end++
				}
			}
			parts = append(parts, string(runes[start:end]))
			start = end
			i = end - 1
		}
	}
	if start < len(runes) {
		parts = append(parts, string(runes[start:]))
	}
	out := make([]*AnalyzedSentence, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, AnalyzePlain(p))
	}
	return out
}


// TokenTag is a soft POS/lemma inject for AnalyzeWithTagger.
type TokenTag struct {
	POS   string
	Lemma string
}

// AnalyzeWithTagger is AnalyzePlain plus optional POS/lemma tags from tagWord.
// tagWord may return multiple readings; empty/nil falls back to untagged tokens.
func AnalyzeWithTagger(text string, tagWord func(token string) []TokenTag) *AnalyzedSentence {
	return AnalyzeWithTaggerAndTokenizer(text, tagWord, tokenizers.NewWordTokenizer())
}

// AnalyzeWithTaggerAndTokenizer tags tokens produced by wt.
// Japanese still uses Sen/kagome encode→decode (TagWord is per-surface and cannot
// replace full-sentence morph analysis).
func AnalyzeWithTaggerAndTokenizer(text string, tagWord func(token string) []TokenTag, wt tokenizers.Tokenizer) *AnalyzedSentence {
	if wt == nil {
		wt = tokenizers.NewWordTokenizer()
	}
	if _, ok := wt.(*jatok.JapaneseWordTokenizer); ok {
		return analyzeJapaneseEncoded(wt.Tokenize(text))
	}
	if tagWord == nil {
		return AnalyzeWithTokenizer(text, wt)
	}
	raw := wt.Tokenize(text)
	positions := tokenizers.BuildPositions(raw)
	readings := make([]*AnalyzedTokenReadings, 0, len(raw)+1)
	ss := SentenceStartTagName
	startTok := NewAnalyzedToken("", &ss, nil)
	startR := NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)
	readings = append(readings, startR)
	prev := ""
	for i, tok := range raw {
		tags := tagWord(tok)
		var ar *AnalyzedTokenReadings
		if len(tags) == 0 {
			at := NewAnalyzedToken(tok, nil, nil)
			ar = NewAnalyzedTokenReadingsAt(at, positions[i])
		} else {
			// first reading primary
			var posPtr, lemmaPtr *string
			if tags[0].POS != "" {
				p := tags[0].POS
				posPtr = &p
			}
			if tags[0].Lemma != "" {
				l := tags[0].Lemma
				lemmaPtr = &l
			}
			at := NewAnalyzedToken(tok, posPtr, lemmaPtr)
			ar = NewAnalyzedTokenReadingsAt(at, positions[i])
			for _, tg := range tags[1:] {
				var pp, lp *string
				if tg.POS != "" {
					p := tg.POS
					pp = &p
				}
				if tg.Lemma != "" {
					l := tg.Lemma
					lp = &l
				}
				ar.AddReading(NewAnalyzedToken(tok, pp, lp), "")
			}
		}
		if prev != "" {
			ar.SetWhitespaceBeforeToken(prev)
		}
		readings = append(readings, ar)
		prev = tok
	}
	softAttachSentenceEnd(readings)
	return NewAnalyzedSentence(readings)
}

// attachPolishHyphenTagger wires TagWord into PolishWordTokenizer like Java setTagger.
func attachPolishHyphenTagger(wt tokenizers.Tokenizer, tagWord func(string) []TokenTag) {
	if tagWord == nil {
		return
	}
	pl, ok := wt.(*pltok.PolishWordTokenizer)
	if !ok || pl == nil {
		return
	}
	pl.SetTagger(polishHyphenTagger(tagWord))
}

type polishHyphenTagger func(string) []TokenTag

func (t polishHyphenTagger) Tag(tokens []string) []pltok.PolishTokenReadings {
	out := make([]pltok.PolishTokenReadings, len(tokens))
	for i, tok := range tokens {
		tags := t(tok)
		if len(tags) == 0 {
			continue
		}
		r := pltok.PolishTokenReadings{IsTagged: true}
		for _, tg := range tags {
			pos := tg.POS
			if pos == "adja" {
				r.HasAdja = true
			}
			if strings.HasPrefix(pos, "adj:") || pos == "adja" {
				r.HasAdjPartial = true
			}
			if strings.HasPrefix(pos, "subst:") {
				r.HasSubstPartial = true
			}
			if strings.HasPrefix(pos, "num:") {
				r.HasNumPartial = true
			}
		}
		out[i] = r
	}
	return out
}
