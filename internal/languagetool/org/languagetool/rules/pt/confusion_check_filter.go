package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/confusion_pairs.txt
var confusionFS embed.FS

var (
	ptConfusionOnce sync.Once
	ptConfusion     rules.ConfusionPairs
	ptConfusionErr  error
)

func loadPTConfusion() rules.ConfusionPairs {
	ptConfusionOnce.Do(func() {
		f, err := confusionFS.Open("data/confusion_pairs.txt")
		if err != nil {
			ptConfusionErr = err
			return
		}
		defer f.Close()
		ptConfusion, ptConfusionErr = rules.LoadConfusionPairs(f)
	})
	if ptConfusionErr != nil {
		panic(ptConfusionErr)
	}
	return ptConfusion
}

// ConfusionCheckFilter ports org.languagetool.rules.pt.ConfusionCheckFilter (base PT pairs).
type ConfusionCheckFilter struct {
	*rules.ConfusionCheckFilter
}

func NewConfusionCheckFilter() *ConfusionCheckFilter {
	return &ConfusionCheckFilter{
		ConfusionCheckFilter: &rules.ConfusionCheckFilter{
			Pairs:              loadPTConfusion(),
			MessageDiacritic:   "se escribe con tilde",
			MessageNoDiacritic: "se escribe de otra manera",
		},
	}
}
