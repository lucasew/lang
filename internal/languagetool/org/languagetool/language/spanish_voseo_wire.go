package language

import (
	"os"
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// WireSpanishVoseoWordTagger opens a Morfologik POS dictionary for
// Spanish.filterRuleMatches voseo POS checks (Java getTagger().tag).
// Returns false if path empty or open fails — leave empty MapWordTagger (fail-closed).
func WireSpanishVoseoWordTagger(dictPath string) bool {
	if tools.JavaStringTrim(dictPath) == "" {
		return false
	}
	mt := tagging.OpenMorfologikTagger(dictPath)
	if mt == nil {
		return false
	}
	SpanishVoseoWordTagger = mt
	SpanishSuggestionIsVoseo = SpanishSuggestionIsVoseoDefault
	return true
}

// TryWireSpanishVoseoWordTagger probes env and known resource locations for spanish.dict.
// Does not invent POS without a real dict file.
func TryWireSpanishVoseoWordTagger() bool {
	if p := os.Getenv("LANG_ES_POS_DICT"); p != "" {
		if WireSpanishVoseoWordTagger(p) {
			return true
		}
	}
	// Walk from cwd toward root for official resource path (Java /es/spanish.dict).
	dir, err := os.Getwd()
	if err != nil {
		return false
	}
	rels := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "es",
			"src", "main", "resources", "org", "languagetool", "resource", "es", "spanish.dict"),
		filepath.Join("third_party", "es", "spanish.dict"),
		filepath.Join("testdata", "es", "spanish.dict"),
	}
	for i := 0; i < 12; i++ {
		for _, rel := range rels {
			cand := filepath.Join(dir, rel)
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				if WireSpanishVoseoWordTagger(cand) {
					return true
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return false
}
