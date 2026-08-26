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

// Equal matches Java TaggedWord.equals (lemma + posTag).
func (t TaggedWord) Equal(o TaggedWord) bool {
	return t.Lemma == o.Lemma && t.PosTag == o.PosTag
}

// String ports TaggedWord.toString → "lemma/posTag".
func (t TaggedWord) String() string {
	return t.Lemma + "/" + t.PosTag
}
