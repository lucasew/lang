package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	estag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/es"
)

// SpanishVoseoWordTagger is the WordTagger used by SpanishSuggestionIsVoseo
// (Java Spanish.getTagger().tag(suggestion).matchesPosTagRegex("V....V.*")).
// Nil → empty MapWordTagger (no POS → no drop; fail-closed without invent).
// RegisterCore or tests set a dict-backed tagger when available.
var SpanishVoseoWordTagger tagging.WordTagger

func init() {
	// Default wire: POS-based voseo drop when WordTagger yields V....V.* tags.
	SpanishSuggestionIsVoseo = SpanishSuggestionIsVoseoDefault
}

// SpanishSuggestionIsVoseoDefault ports Spanish.filterRuleMatches voseo POS check.
func SpanishSuggestionIsVoseoDefault(suggestion string) bool {
	if suggestion == "" {
		return false
	}
	wt := SpanishVoseoWordTagger
	if wt == nil {
		wt = tagging.MapWordTagger{}
	}
	tagger := estag.NewSpanishTagger(wt)
	atrs := tagger.Tag([]string{suggestion})
	if len(atrs) == 0 || atrs[0] == nil {
		return false
	}
	// Java: private final Pattern voseoPostagPatern = Pattern.compile("V....V.*");
	return atrs[0].MatchesPosTagRegex("V....V.*")
}
