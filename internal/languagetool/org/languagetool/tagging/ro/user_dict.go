package ro

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

// WithUserDict overlays extra entries on a WordTagger (later entries win for same key via merge).
func WithUserDict(base tagging.WordTagger, user map[string][]tagging.TaggedWord) tagging.WordTagger {
	if base == nil {
		return tagging.MapWordTagger(user)
	}
	// MapWordTagger is map type
	if m, ok := base.(tagging.MapWordTagger); ok {
		out := tagging.MapWordTagger{}
		for k, v := range m {
			out[k] = append([]tagging.TaggedWord(nil), v...)
		}
		for k, v := range user {
			out[k] = append(out[k], v...)
		}
		return out
	}
	return base
}

// MergeReadingsForLemmaGroups merges multiple lemma/POS pairs for one surface (green merge path).
func MergeTaggedWords(groups ...[]tagging.TaggedWord) []tagging.TaggedWord {
	var out []tagging.TaggedWord
	seen := map[string]struct{}{}
	for _, g := range groups {
		for _, tw := range g {
			key := tw.Lemma + "|" + tw.PosTag
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, tw)
		}
	}
	return out
}
