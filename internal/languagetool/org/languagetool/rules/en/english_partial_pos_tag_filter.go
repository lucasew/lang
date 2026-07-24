package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishPartialPosTagFilter ports org.languagetool.rules.en.EnglishPartialPosTagFilter
// (PartialPosTagFilter that tags and disambiguates a single token in Java).
//
// Java: Languages.getLanguageForShortCode("en").getTagger() then getDisambiguator()
// on a one-token AnalyzedSentence. Without both process-wide hooks (WireEnglishFilterTagger
// + WireEnglishFilterDisambiguator), Accept fails closed — no invent.
type EnglishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewEnglishPartialPosTagFilter(tag func(string) []string) *EnglishPartialPosTagFilter {
	if tag == nil {
		tag = englishPartialTagAndDisambiguatePOS
	}
	return &EnglishPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

// englishPartialTagAndDisambiguatePOS ports EnglishPartialPosTagFilter.tag:
// tagger.tag([token]) → disambiguator.disambiguate(AnalyzedSentence) → POS list.
func englishPartialTagAndDisambiguatePOS(partial string) []string {
	tw := getFilterTagWord()
	d := getFilterDisambiguator()
	if tw == nil || d == nil {
		return nil
	}
	// Java Tagger.tag(Collections.singletonList(token)) — no SENT_START in that list.
	tags := tw(partial)
	var readings []*languagetool.AnalyzedToken
	if len(tags) == 0 {
		readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(partial, nil, nil)}
	} else {
		readings = make([]*languagetool.AnalyzedToken, 0, len(tags))
		for _, t := range tags {
			var pos, lemma *string
			if t.POS != "" {
				p := t.POS
				pos = &p
			}
			if t.Lemma != "" {
				l := t.Lemma
				lemma = &l
			}
			readings = append(readings, languagetool.NewAnalyzedToken(partial, pos, lemma))
		}
	}
	atr := languagetool.NewAnalyzedTokenReadingsList(readings, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{atr})
	out := d.Disambiguate(sent)
	if out == nil {
		return collectPOSTags(atr)
	}
	// Java returns disambiguated.getTokens(); collect POS from all non-empty tokens.
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

func collectPOSTags(atr *languagetool.AnalyzedTokenReadings) []string {
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
