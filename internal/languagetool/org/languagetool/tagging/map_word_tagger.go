package tagging

// MapWordTagger is a simple WordTagger backed by an exact-form map.
type MapWordTagger map[string][]TaggedWord

func (m MapWordTagger) Tag(word string) []TaggedWord {
	return append([]TaggedWord(nil), m[word]...)
}
