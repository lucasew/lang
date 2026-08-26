package crh

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const CrimeanTatarTaggerDictPath = "/crh/crimean_tatar.dict"

// CrimeanTatarTagger ports org.languagetool.tagging.crh.CrimeanTatarTagger.
type CrimeanTatarTagger struct {
	*tagging.BaseTagger
}

func NewCrimeanTatarTagger(wt tagging.WordTagger) *CrimeanTatarTagger {
	return &CrimeanTatarTagger{BaseTagger: tagging.NewBaseTagger(wt, CrimeanTatarTaggerDictPath, "crh", false)}
}
