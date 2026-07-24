package language

// Italian language twin — see more_variants.go for the Italian var.
func NewItalian() ItalianLang {
	return Italian
}

func ItalianShortCode() string { return Italian.ShortCode }
func ItalianName() string      { return Italian.Name }
