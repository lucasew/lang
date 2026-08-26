package eval

import "math"

// GetWeightedFMeasure ports FMeasure.getWeightedFMeasure (beta=0.5).
func GetWeightedFMeasure(precision, recall float64) float64 {
	return GetFMeasure(precision, recall, 0.5)
}

// GetFMeasure ports FMeasure.getFMeasure.
func GetFMeasure(precision, recall, beta float64) float64 {
	if precision == 0 && recall == 0 {
		return 0
	}
	betaSquared := math.Pow(beta, 2)
	den := betaSquared*precision + recall
	if den == 0 {
		return 0
	}
	return (1 + betaSquared) * (precision * recall) / den
}
