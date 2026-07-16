package crh

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"

const (
	CrimeanTatarSynthDict = "/crh/crimean_tatar_synth.dict"
	CrimeanTatarTagsFile  = "/crh/crimean_tatar_tags.txt"
)

// CrimeanTatarSynthesizer ports org.languagetool.synthesis.crh.CrimeanTatarSynthesizer.
type CrimeanTatarSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewCrimeanTatarSynthesizer(manual *synthesis.ManualSynthesizer) *CrimeanTatarSynthesizer {
	base := synthesis.NewBaseSynthesizer("crh", manual)
	base.ResourceFileName = CrimeanTatarSynthDict
	base.TagFileName = CrimeanTatarTagsFile
	return &CrimeanTatarSynthesizer{BaseSynthesizer: base}
}
