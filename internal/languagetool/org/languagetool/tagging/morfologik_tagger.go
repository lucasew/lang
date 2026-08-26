package tagging

import (
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
)

// MorfologikTagger ports org.languagetool.tagging.MorfologikTagger as a WordTagger.
// Lookup may be injected for tests; otherwise DictPath is opened lazily via
// attic morfologik (same as Java DictionaryLookup). Empty if open fails.
type MorfologikTagger struct {
	DictPath   string
	InternTags bool
	// Lookup returns stem/tag pairs for a surface form (injected dictionary).
	Lookup func(word string) []TaggedWord

	once sync.Once
	dict *atticmorfo.Dictionary
}

func NewMorfologikTagger(dictPath string) *MorfologikTagger {
	return &MorfologikTagger{DictPath: dictPath}
}

func NewMorfologikTaggerWithLookup(lookup func(word string) []TaggedWord) *MorfologikTagger {
	return &MorfologikTagger{Lookup: lookup}
}

// OpenMorfologikTagger opens dictPath immediately; nil if the file cannot be opened.
func OpenMorfologikTagger(dictPath string) *MorfologikTagger {
	if dictPath == "" {
		return nil
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return nil
	}
	t := &MorfologikTagger{DictPath: dictPath, dict: d}
	// once already "done" with dict set — force once to no-op
	t.once.Do(func() {})
	return t
}

func (t *MorfologikTagger) GetInternTags() bool        { return t.InternTags }
func (t *MorfologikTagger) SetInternTags(enabled bool) { t.InternTags = enabled }

func (t *MorfologikTagger) ensureDict() *atticmorfo.Dictionary {
	if t == nil {
		return nil
	}
	if t.dict != nil {
		return t.dict
	}
	if t.DictPath == "" {
		return nil
	}
	t.once.Do(func() {
		d, err := atticmorfo.OpenDictionary(t.DictPath)
		if err == nil {
			t.dict = d
		}
	})
	return t.dict
}

// Tag ports MorfologikTagger.tag.
func (t *MorfologikTagger) Tag(word string) []TaggedWord {
	if t == nil {
		return nil
	}
	if t.Lookup != nil {
		res := t.Lookup(word)
		if !t.InternTags {
			return res
		}
		out := make([]TaggedWord, len(res))
		copy(out, res)
		return out
	}
	d := t.ensureDict()
	if d == nil || word == "" {
		return nil
	}
	forms, err := d.Lookup(word)
	if err != nil || len(forms) == 0 {
		return nil
	}
	// Java MorfologikTagger.tag: strip last byte when frequency-included
	// (freq data is the last byte of the tag, without a separator).
	freqStrip := d.FrequencyIncluded()
	out := make([]TaggedWord, 0, len(forms))
	for _, f := range forms {
		tag := f.Tag
		if freqStrip && len(tag) > 1 {
			tag = tag[:len(tag)-1]
		}
		out = append(out, NewTaggedWord(f.Stem, tag))
	}
	return out
}

func (t *MorfologikTagger) lookup(word string) []TaggedWord {
	return append([]TaggedWord(nil), t.Tag(word)...)
}

var _ WordTagger = (*MorfologikTagger)(nil)
