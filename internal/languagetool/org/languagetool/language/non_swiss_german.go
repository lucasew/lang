package language

// NonSwissGerman marks German variants that use ß (not Swiss ss-only orthography).
// Ports org.languagetool.language.NonSwissGerman marker interface surface.
func IsNonSwissGerman(code string) bool {
	switch code {
	case "de", "de-DE", "de-AT", "de-LU", "de-LI":
		return true
	case "de-CH":
		return false
	default:
		return false
	}
}
