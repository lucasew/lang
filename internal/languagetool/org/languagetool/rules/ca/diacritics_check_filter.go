package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/confusion_pairs.txt
var confusionFS embed.FS

var (
	caConfusionOnce sync.Once
	caConfusion     rules.ConfusionPairs
	caConfusionErr  error
)

func loadCAConfusion() rules.ConfusionPairs {
	caConfusionOnce.Do(func() {
		f, err := confusionFS.Open("data/confusion_pairs.txt")
		if err != nil {
			caConfusionErr = err
			return
		}
		defer f.Close()
		caConfusion, caConfusionErr = rules.LoadConfusionPairs(f)
	})
	if caConfusionErr != nil {
		panic(caConfusionErr)
	}
	return caConfusion
}

// DiacriticsCheckFilter ports org.languagetool.rules.ca.DiacriticsCheckFilter.
type DiacriticsCheckFilter struct {
	*rules.ConfusionCheckFilter
}

func NewDiacriticsCheckFilter() *DiacriticsCheckFilter {
	return &DiacriticsCheckFilter{
		ConfusionCheckFilter: &rules.ConfusionCheckFilter{
			Pairs:                loadCAConfusion(),
			MessageDiacritic:     "s'escriu amb accent",
			MessageNoDiacritic:   "s'escriu d'una altra manera",
			GenderProbes:         rules.CAGenderNumberProbes,
			ExpandAllSuggestions: true, // CA rewrites every suggestion template
		},
	}
}
