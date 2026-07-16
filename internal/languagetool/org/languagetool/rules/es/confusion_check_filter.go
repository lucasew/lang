package es

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/confusion_pairs.txt
var confusionFS embed.FS

var (
	esConfusionOnce sync.Once
	esConfusion     rules.ConfusionPairs
	esConfusionErr  error
)

func loadESConfusion() rules.ConfusionPairs {
	esConfusionOnce.Do(func() {
		f, err := confusionFS.Open("data/confusion_pairs.txt")
		if err != nil {
			esConfusionErr = err
			return
		}
		defer f.Close()
		esConfusion, esConfusionErr = rules.LoadConfusionPairs(f)
	})
	if esConfusionErr != nil {
		panic(esConfusionErr)
	}
	return esConfusion
}

// ConfusionCheckFilter ports org.languagetool.rules.es.ConfusionCheckFilter.
type ConfusionCheckFilter struct {
	*rules.ConfusionCheckFilter
}

func NewConfusionCheckFilter() *ConfusionCheckFilter {
	return &ConfusionCheckFilter{
		ConfusionCheckFilter: &rules.ConfusionCheckFilter{
			Pairs:              loadESConfusion(),
			MessageDiacritic:   "se escribe con tilde",
			MessageNoDiacritic: "se escribe de otra manera",
		},
	}
}
