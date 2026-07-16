package multiLang

// MultiLanguageTextCheckEval ports evaluation config for multi-language checks.
type MultiLanguageTextCheckEval struct {
	ServerURL        string
	MainLanguage     string
	InjectLanguages  []string
	MaxSentences     int
	ReportInjectHits bool
}

func NewMultiLanguageTextCheckEval(mainLang string) *MultiLanguageTextCheckEval {
	return &MultiLanguageTextCheckEval{
		ServerURL:    "http://localhost:8081",
		MainLanguage: mainLang,
		MaxSentences: 1000,
	}
}
