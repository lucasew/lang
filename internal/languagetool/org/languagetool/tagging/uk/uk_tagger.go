package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"strings"
)

const UkrainianDictPath = "/uk/uk.dict"

type UkrainianTagger struct{ *tagging.BaseTagger }

func NewUkrainianTagger(wt tagging.WordTagger) *UkrainianTagger {
	return &UkrainianTagger{BaseTagger: tagging.NewBaseTagger(wt, UkrainianDictPath, "uk", false)}
}

func (t *UkrainianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		w := strings.ReplaceAll(word, "’", "'")
		var readings []*languagetool.AnalyzedToken
		if sp := SpecialPOSTag(w); sp != "" {
			p := sp
			lemma := w
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		if dyn := DynamicAdjReadings(w); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		if dyn := DynamicDirectionalAdjReadings(w); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		if dyn := DynamicNumericReadings(w); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		if lemma, ipos, ok := IntjReading(w); ok {
			p, l := ipos, lemma
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &l)}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		if dyn := FixedPartReadings(w); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		for _, tw := range t.TagWord(w) {
			readings = append(readings, toTok(word, tw))
		}
		lower := strings.ToLower(w)
		if len(readings) == 0 && w != lower && !tools.IsMixedCase(w) {
			for _, tw := range t.TagWord(lower) {
				readings = append(readings, toTok(word, tw))
			}
		}
		if len(readings) == 0 {
			if pref := TryNoDashPrefixTags(w, func(right string) []*languagetool.AnalyzedToken {
				var rs []*languagetool.AnalyzedToken
				for _, tw := range t.TagWord(right) {
					rs = append(rs, toTok(right, tw))
				}
				low := strings.ToLower(right)
				if len(rs) == 0 && right != low {
					for _, tw := range t.TagWord(low) {
						rs = append(rs, toTok(right, tw))
					}
				}
				return rs
			}); len(pref) > 0 {
				readings = pref
			}
		}
		if len(readings) == 0 {
			for _, cand := range MissingApostropheCandidates(w) {
				for _, tw := range t.TagWord(cand) {
					// mark :bad like Java dynamic missing apostrophe
					tw2 := tw
					if tw2.PosTag != "" && !strings.Contains(tw2.PosTag, ":bad") {
						tw2.PosTag = tw2.PosTag + ":bad"
					}
					readings = append(readings, toTok(word, tw2))
				}
				if len(readings) > 0 {
					break
				}
			}
		}
		if len(readings) == 0 {
			for _, cand := range MissingHyphenCandidates(w) {
				// re-tag hyphenated candidate via compound path specials
				if dyn := DynamicAdjReadings(cand); len(dyn) > 0 {
					for _, d := range dyn {
						p, l := d.POS, d.Lemma
						readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
					}
					break
				}
				if pref := TryNoDashPrefixTags(strings.ReplaceAll(cand, "-", ""), func(right string) []*languagetool.AnalyzedToken {
					var rs []*languagetool.AnalyzedToken
					for _, tw := range t.TagWord(right) {
						rs = append(rs, toTok(right, tw))
					}
					return rs
				}); len(pref) > 0 {
					readings = pref
					break
				}
				// -небудь indefinite pronouns (Java MISSING_HYPHEN + pronoun POS + :bad)
				if strings.HasSuffix(strings.ToLower(cand), "-небудь") {
					if nebud := NebudMissingHyphenReadings(word, cand, t.TagWord); len(nebud) > 0 {
						readings = nebud
						break
					}
				}
				// direct right-of-hyphen dict lookup (мінітест → міні-тест → тест)
				if i := strings.Index(cand, "-"); i > 0 {
					right := cand[i+1:]
					for _, tw := range t.TagWord(right) {
						readings = append(readings, toTok(word, tw))
					}
					if low := strings.ToLower(right); len(readings) == 0 && right != low {
						for _, tw := range t.TagWord(low) {
							readings = append(readings, toTok(word, tw))
						}
					}
					if len(readings) > 0 {
						break
					}
				}
			}
		}
		if len(readings) == 0 {
			if np := CompoundNumrPOS(w); np != "" {
				p, l := np, w
				readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &l)}
			}
		}
		if len(readings) == 0 && strings.Contains(w, "-") {
			if ft := FullTagMatchReadings(w, t.TagWord); len(ft) > 0 {
				readings = ft
			}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		pos += len([]rune(word))
	}
	return out
}

func toTok(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
	var pos, lemma *string
	if tw.PosTag != "" {
		p := tw.PosTag
		pos = &p
	}
	if tw.Lemma != "" {
		l := tw.Lemma
		lemma = &l
	}
	return languagetool.NewAnalyzedToken(surface, pos, lemma)
}
