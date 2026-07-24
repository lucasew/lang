package tagging

// CombiningTagger ports org.languagetool.tagging.CombiningTagger.
type CombiningTagger struct {
	tagger1                  WordTagger
	tagger2                  WordTagger
	removalTagger            WordTagger // optional
	overwriteWithSecondTagger bool
}

func NewCombiningTagger(tagger1, tagger2 WordTagger, overwriteWithSecondTagger bool) *CombiningTagger {
	return NewCombiningTaggerWithRemoval(tagger1, tagger2, nil, overwriteWithSecondTagger)
}

func NewCombiningTaggerWithRemoval(tagger1, tagger2, removalTagger WordTagger, overwriteWithSecondTagger bool) *CombiningTagger {
	return &CombiningTagger{
		tagger1:                   tagger1,
		tagger2:                   tagger2,
		removalTagger:             removalTagger,
		overwriteWithSecondTagger: overwriteWithSecondTagger,
	}
}

func (c *CombiningTagger) Tag(word string) []TaggedWord {
	var result []TaggedWord
	result = append(result, c.tagger2.Tag(word)...)
	if !(c.overwriteWithSecondTagger && len(result) > 0) {
		result = append(result, c.tagger1.Tag(word)...)
	}
	if c.removalTagger != nil {
		removal := c.removalTagger.Tag(word)
		result = removeAllTagged(result, removal)
	}
	return result
}

func (c *CombiningTagger) GetRemovalTagger() WordTagger { return c.removalTagger }

func removeAllTagged(result, removal []TaggedWord) []TaggedWord {
	var out []TaggedWord
	for _, tw := range result {
		drop := false
		for _, r := range removal {
			if tw.Equal(r) {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, tw)
		}
	}
	return out
}
