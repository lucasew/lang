package languagetool

// AdaptTextLevelLocalMatch ports TextCheckCallable.getTextLevelRuleMatches
// adaptation of one match:
//
//	from/to = findLineColumn(match.from/to)
//	optional annotatedText original mapping
//	column -= (line == 0 ? 1 : 0)
//
// mapOriginal may be nil (plain text / no markup). When set it is the same
// signature as AnnotatedText.GetOriginalTextPositionFor.
func AdaptTextLevelLocalMatch(
	m LocalMatch,
	sentences []SentenceData,
	mapOriginal func(pos int, isToPos bool) int,
) LocalMatch {
	from := FindLineColumnInSentences(sentences, m.FromPos)
	to := FindLineColumnInSentences(sentences, m.ToPos)
	out := m
	newFrom, newTo := m.FromPos, m.ToPos
	if mapOriginal != nil {
		newFrom = mapOriginal(m.FromPos, false)
		if m.ToPos > 0 {
			newTo = mapOriginal(m.ToPos-1, true) + 1
		}
	}
	out.FromPos = newFrom
	out.ToPos = newTo
	out.Line = from.Line
	out.EndLine = to.Line
	// Java: setColumn(from.column - (from.line == 0 ? 1 : 0))
	colAdj := 0
	if from.Line == 0 {
		colAdj = 1
	}
	endColAdj := 0
	if to.Line == 0 {
		endColAdj = 1
	}
	out.Column = from.Column - colAdj
	out.EndColumn = to.Column - endColAdj
	return out
}

// IgnoreRangesFromLanguageMatches ports TextCheckCallable handling of
// match.getNewLanguageMatches() → Range(startOffset, startOffset+text.length(), lang).
// Uses the first map entry's key as language (Java iterator().next().getKey()).
// sentenceFrom/To are document UTF-16 offsets for the whole sentence span.
//
// Java LinkedHashMap preserves insertion order. Go maps do not; when multiple
// keys exist we pick the lexicographically smallest key for determinism.
// Callers that need insertion order should pass a single-entry map (Java
// speller typically sets one preferred foreign language).
func IgnoreRangesFromLanguageMatches(sentenceFrom, sentenceTo int, rates map[string]float32) (Range, bool) {
	if len(rates) == 0 {
		return Range{}, false
	}
	lang := ""
	for k := range rates {
		if lang == "" || k < lang {
			lang = k
		}
	}
	if lang == "" {
		return Range{}, false
	}
	return NewRange(sentenceFrom, sentenceTo, lang), true
}

// AppendUniqueIgnoreRange adds r if not already Equal to an existing range.
func AppendUniqueIgnoreRange(ranges []Range, r Range) []Range {
	for _, x := range ranges {
		if x.Equal(r) {
			return ranges
		}
	}
	return append(ranges, r)
}

// ApplyNewLanguageMatchesToSentence ports the per-match block in getOtherRuleMatches:
// ignore range + extendedSentenceRange.updateLanguageConfidenceRates.
func ApplyNewLanguageMatchesToSentence(
	ignore []Range,
	ext *ExtendedSentenceRange,
	sentenceFrom, sentenceTo int,
	rates map[string]float32,
) []Range {
	if len(rates) == 0 {
		return ignore
	}
	if r, ok := IgnoreRangesFromLanguageMatches(sentenceFrom, sentenceTo, rates); ok {
		ignore = AppendUniqueIgnoreRange(ignore, r)
	}
	if ext != nil {
		ext.UpdateLanguageConfidenceRates(rates)
	}
	return ignore
}
