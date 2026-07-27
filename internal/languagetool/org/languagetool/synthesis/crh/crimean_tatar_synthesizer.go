package crh

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"

const (
	CrimeanTatarSynthDict = "/crh/crimean_tatar_synth.dict"
	// Java BaseSynthesizer path: crimean_tatar_synth.dict_tags.txt
	CrimeanTatarTagsFile = "/crh/crimean_tatar_synth.dict_tags.txt"
)

// CrimeanTatarSynthesizer ports org.languagetool.synthesis.crh.CrimeanTatarSynthesizer.
type CrimeanTatarSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// INSTANCE ports CrimeanTatarSynthesizer.INSTANCE (Java private ctor + public static field).
// Official binary dict is a built artifact (not in inspiration source tree); surface carries
// Java resource paths. Full open-from-dict load is the synthesis/crh leaf when dict is present.
var INSTANCE = NewCrimeanTatarSynthesizer(nil)

// Instance returns CrimeanTatarSynthesizer.INSTANCE.
func Instance() *CrimeanTatarSynthesizer { return INSTANCE }

func NewCrimeanTatarSynthesizer(manual *synthesis.ManualSynthesizer) *CrimeanTatarSynthesizer {
	base := synthesis.NewBaseSynthesizer("crh", manual)
	base.ResourceFileName = CrimeanTatarSynthDict
	base.TagFileName = CrimeanTatarTagsFile
	return &CrimeanTatarSynthesizer{BaseSynthesizer: base}
}
