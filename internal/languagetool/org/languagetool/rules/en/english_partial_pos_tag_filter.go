package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishPartialPosTagFilter ports org.languagetool.rules.en.EnglishPartialPosTagFilter
// (PartialPosTagFilter that tags and disambiguates a single token in Java).
//
// Java (EnglishPartialPosTagFilter.tag):
//
//	var englishLanguage = Languages.getLanguageForShortCode("en");
//	List<AnalyzedTokenReadings> tags = englishLanguage.getTagger().tag(Collections.singletonList(token));
//	AnalyzedTokenReadings[] atr = tags.toArray(...);
//	AnalyzedSentence disambiguated = englishLanguage.getDisambiguator().disambiguate(new AnalyzedSentence(atr));
//	return Arrays.asList(disambiguated.getTokens());
//
// IOException → RuntimeException "Could not tag and disambiguate '…'".
//
// Warning (Java javadoc): single-token disambiguator only; do NOT use in
// disambiguation.xml (endless loop) — use NoDisambiguationEnglishPartialPosTagFilter.
//
// Without both process-wide hooks (WireEnglishFilterTagger +
// WireEnglishFilterDisambiguator), Accept fails closed — no invent POS/disambig.
type EnglishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

// NewEnglishPartialPosTagFilter constructs the filter.
// tag maps a partial token to POS tags (injected for tests); nil →
// englishPartialTagAndDisambiguatePOS (process-wide tagger + disambiguator).
func NewEnglishPartialPosTagFilter(tag func(string) []string) *EnglishPartialPosTagFilter {
	if tag == nil {
		tag = englishPartialTagAndDisambiguatePOS
	}
	return &EnglishPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

// englishPartialTagAndDisambiguatePOS ports EnglishPartialPosTagFilter.tag:
// tagger.tag([token]) → disambiguator.disambiguate(AnalyzedSentence) → POS list
// (PartialPosTagFilter.partialTagHasRequiredTag walks readings' POS tags).
func englishPartialTagAndDisambiguatePOS(partial string) []string {
	tw := getFilterTagWord()
	d := getFilterDisambiguator()
	if tw == nil || d == nil {
		// Fail closed: Java always has Languages.getLanguageForShortCode("en")
		// tagger + disambiguator; without both wires we must not invent.
		return nil
	}
	// Java Tagger.tag(Collections.singletonList(token)) — no SENT_START in that list.
	// BaseTagger always returns one ATR per input token (untagged when dict has no entry).
	atr := filterTagWordToATR(partial, tw)
	if atr == nil {
		return nil
	}
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{atr})
	out := d.Disambiguate(sent)
	if out == nil {
		// Java would NPE on null disambiguated; fail closed (no pre-disambig invent).
		return nil
	}
	// Java returns disambiguated.getTokens(); partialTagHasRequiredTag iterates every
	// reading with non-null POSTag. Skip pure structural SENT_* markers so they do
	// not false-match postag_regexp (single-token path normally has none).
	return collectPOSTagsFromSentence(out)
}

// collectPOSTagsFromSentence flattens POS tags from AnalyzedSentence.getTokens()
// the way PartialPosTagFilter.partialTagHasRequiredTag inspects List<ATR>.
func collectPOSTagsFromSentence(sent *languagetool.AnalyzedSentence) []string {
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
