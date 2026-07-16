package eval

// PrecisionRecall ports org.languagetool.dev.eval.PrecisionRecall.
type PrecisionRecall struct {
	Precision float64
	Recall    float64
}

func NewPrecisionRecall(precision, recall float64) PrecisionRecall {
	return PrecisionRecall{Precision: precision, Recall: recall}
}

func (p PrecisionRecall) GetPrecision() float64 { return p.Precision }
func (p PrecisionRecall) GetRecall() float64    { return p.Recall }

// F05 is the weighted F(0.5) for this pair.
func (p PrecisionRecall) F05() float64 {
	return GetWeightedFMeasure(p.Precision, p.Recall)
}
