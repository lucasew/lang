package languagetool

// DetectedLanguage ports org.languagetool.DetectedLanguage using short codes
// until full Language objects are wired everywhere.
type DetectedLanguage struct {
	GivenLanguageCode    string
	DetectedLanguageCode string
	DetectionConfidence  float32
	DetectionSource      *string
}

func NewDetectedLanguage(given, detected string) DetectedLanguage {
	return NewDetectedLanguageFull(given, detected, 1.0, nil)
}

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

func (d DetectedLanguage) String() string { return d.DetectedLanguageCode }
