package tagging

// WordTagger ports org.languagetool.tagging.WordTagger.
type WordTagger interface {
	Tag(word string) []TaggedWord
}
