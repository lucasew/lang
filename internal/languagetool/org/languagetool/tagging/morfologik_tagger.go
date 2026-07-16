package tagging

// MorfologikTagger ports org.languagetool.tagging.MorfologikTagger as a WordTagger
// surface. Binary Morfologik dictionary I/O is not wired yet; callers inject a
// Lookup function or fall back to an empty result.
type MorfologikTagger struct {
	DictPath   string
	InternTags bool
	// Lookup returns stem/tag pairs for a surface form (injected dictionary).
	Lookup func(word string) []TaggedWord
}

func NewMorfologikTagger(dictPath string) *MorfologikTagger {
	return &MorfologikTagger{DictPath: dictPath}
}

func NewMorfologikTaggerWithLookup(lookup func(word string) []TaggedWord) *MorfologikTagger {
	return &MorfologikTagger{Lookup: lookup}
}

func (t *MorfologikTagger) GetInternTags() bool        { return t.InternTags }
func (t *MorfologikTagger) SetInternTags(enabled bool) { t.InternTags = enabled }

// Tag ports MorfologikTagger.tag.
func (t *MorfologikTagger) Tag(word string) []TaggedWord {
	if t == nil || t.Lookup == nil {
		return nil
	}
	res := t.Lookup(word)
	if !t.InternTags {
		return res
	}
	// intern is a no-op for Go strings beyond returning copies
	out := make([]TaggedWord, len(res))
	copy(out, res)
	return out
}

func (t *MorfologikTagger) lookup(word string) []TaggedWord {
	if t.Lookup == nil {
		return nil
	}
	return append([]TaggedWord(nil), t.Lookup(word)...)
}

var _ WordTagger = (*MorfologikTagger)(nil)
