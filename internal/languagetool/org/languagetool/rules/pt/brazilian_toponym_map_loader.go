package pt

// BrazilianToponymMapLoader ports org.languagetool.rules.pt.BrazilianToponymMapLoader.
type BrazilianToponymMapLoader struct{}

// Load returns the shared embedded municipality map (Java load path).
func (BrazilianToponymMapLoader) Load() *BrazilianToponymMap {
	return LoadBrazilianToponymMap()
}

// States lists Brazilian state codes used by the loader.
func (BrazilianToponymMapLoader) States() []string {
	return append([]string(nil), brStates...)
}
