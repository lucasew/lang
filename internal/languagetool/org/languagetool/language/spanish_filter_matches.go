package language

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// esJavaTrailingWS ports Java replaceAll("\\s+$", "") (ASCII whitespace only).
var esJavaTrailingWS = regexp.MustCompile(`[ \t\n\v\f\r]+$`)

// Spanish AI_ES_GGEC obsolete-diacritic suggestions (Spanish.suggestionsToAvoid).
// Java: Arrays.asList(...).contains(suggestion.toLowerCase()).
var spanishSuggestionsToAvoid = map[string]struct{}{
	"aquél": {}, "aquélla": {}, "aquéllas": {}, "aquéllos": {},
	"ésa": {}, "ésas": {}, "ése": {}, "ésos": {},
	"ésta": {}, "éstas": {}, "éste": {}, "éstos": {},
	"sólo": {},
}

// SpanishSuggestionIsVoseo optional hook for Spanish.filterRuleMatches voseo drop
// (Java: tagger.tag(suggestion).matchesPosTagRegex("V....V.*")).
// Nil → skip voseo filter (fail-closed: do not invent drops without POS).
var SpanishSuggestionIsVoseo func(suggestion string) bool

func init() {
	languagetool.FilterSpanishRuleMatchesHook = FilterSpanishRuleMatches
}

// FilterSpanishRuleMatches ports Spanish.filterRuleMatches (AI_ES_GGEC filters + casing rewrite).
//
// Voseo drop: SpanishSuggestionIsVoseo (default POS V....V.* via SpanishVoseoWordTagger).
// Incomplete vs Java (explicit, not invent):
//   - Default SpanishVoseoWordTagger is empty until dict-backed tagger is set.
//   - Period / sentence-start drops need SentenceText; without it, keep (fail-closed).
func FilterSpanishRuleMatches(matches []languagetool.LocalMatch) []languagetool.LocalMatch {
	if len(matches) == 0 {
		return nil
	}
	out := make([]languagetool.LocalMatch, 0, len(matches))
	for i := range matches {
		m := matches[i]
		if len(m.Suggestions) == 1 && strings.HasPrefix(m.RuleID, "AI_ES_GGEC") {
			sug := m.Suggestions[0]
			// ignore adding punctuation at the sentence end
			if m.RuleID == "AI_ES_GGEC_MISSING_PUNCTUATION" && strings.HasSuffix(sug, ".") {
				if spanishDropTrailingPeriod(m, sug) {
					continue
				}
			}
			// avoid obsolete diacritics
			if _, bad := spanishSuggestionsToAvoid[strings.ToLower(sug)]; bad {
				continue
			}
			// avoid lowercase at the sentence start
			if spanishDropLowercaseSentenceStart(m, sug) {
				continue
			}
			// avoid voseo forms in suggestions
			if SpanishSuggestionIsVoseo != nil && SpanishSuggestionIsVoseo(sug) {
				continue
			}
			// casing-only rewrite (Java setOriginalErrorStr then equalsIgnoreCase)
			errStr := m.OriginalSurface()
			if errStr != "" && strings.EqualFold(sug, errStr) {
				m.Message = "Mayúsculas y minúsculas recomendadas."
				m.ShortMessage = "Mayúsculas y minúsculas"
				m.IssueType = "typographical"
				m.CategoryID = "CASING"
				m.CategoryName = "Mayúsculas y minúsculas"
				m.RuleID = strings.ReplaceAll(m.RuleID, "ORTHOGRAPHY", "CASING")
			}
		}
		out = append(out, m)
	}
	return out
}

// spanishDropTrailingPeriod ports AI_ES_GGEC_MISSING_PUNCTUATION period-at-end skip.
// Java: getText().replaceAll("\\s+$", "") — Pattern \s without UNICODE_CHARACTER_CLASS
// is [ \t\n\x0B\f\r] (not NBSP).
func spanishDropTrailingPeriod(m languagetool.LocalMatch, sug string) bool {
	sent := m.SentenceText
	if sent == "" {
		return false
	}
	trimmed := esJavaTrailingWS.ReplaceAllString(sent, "")
	if len(sug) == 0 {
		return false
	}
	prefix := sug[:len(sug)-1]
	return strings.HasSuffix(trimmed, prefix)
}

// spanishDropLowercaseSentenceStart ports:
// sentence.getText().trim().startsWith(StringTools.uppercaseFirstChar(suggestion)).
func spanishDropLowercaseSentenceStart(m languagetool.LocalMatch, sug string) bool {
	sent := m.SentenceText
	if sent == "" {
		return false
	}
	trimmed := tools.JavaStringTrim(sent)
	up := tools.UppercaseFirstChar(sug)
	return strings.HasPrefix(trimmed, up)
}
