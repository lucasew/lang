package tagging

// TaggedWord ports org.languagetool.tagging.TaggedWord.
type TaggedWord struct {
	Lemma  string
	PosTag string
}

func NewTaggedWord(lemma, posTag string) TaggedWord {
	return TaggedWord{Lemma: lemma, PosTag: posTag}
}

func (t TaggedWord) GetLemma() string  { return t.Lemma }
func (t TaggedWord) GetPosTag() string { return t.PosTag }
