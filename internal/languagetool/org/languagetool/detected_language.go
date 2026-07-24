package languagetool

// DetectedLanguage ports org.languagetool.DetectedLanguage.
// Java holds Language objects; Go uses short codes until full Language wiring is universal.
// String() matches getDetectedLanguage().getShortCodeWithCountryAndVariant().
type DetectedLanguage struct {
	GivenLanguageCode    string
	DetectedLanguageCode string
	DetectionConfidence  float32
	DetectionSource      *string
}

// NewDetectedLanguage ports DetectedLanguage(given, detected) with confidence 1.0, source null.
func NewDetectedLanguage(given, detected string) DetectedLanguage {
	return NewDetectedLanguageFull(given, detected, 1.0, nil)
}

// NewDetectedLanguageFull ports the 4-arg constructor (since 4.4).
func NewDetectedLanguageFull(given, detected string, confidence float32, source *string) DetectedLanguage {
	return DetectedLanguage{
		GivenLanguageCode:    given,
		DetectedLanguageCode: detected,
		DetectionConfidence:  confidence,
		DetectionSource:      source,
	}
}

func (d DetectedLanguage) GetGivenLanguageCode() string    { return d.GivenLanguageCode }
func (d DetectedLanguage) GetDetectedLanguageCode() string { return d.DetectedLanguageCode }
func (d DetectedLanguage) GetDetectionConfidence() float32 { return d.DetectionConfidence }
func (d DetectedLanguage) GetDetectionSource() *string     { return d.DetectionSource }

// String ports toString → detectedLanguage.getShortCodeWithCountryAndVariant().
func (d DetectedLanguage) String() string { return d.DetectedLanguageCode }
