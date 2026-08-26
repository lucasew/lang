package tools

// linguisticServices ports Tools.linguServices (Java org.languagetool.LinguServices).
// Stored as any to avoid tools → languagetool import cycle.
var linguisticServices any

// SetLinguisticServices ports Tools.setLinguisticServices.
func SetLinguisticServices(ls any) {
	linguisticServices = ls
}

// IsExternSpeller ports Tools.isExternSpeller — true when linguistic services are set.
func IsExternSpeller() bool {
	return linguisticServices != nil
}

// GetLinguisticServices ports Tools.getLinguisticServices.
func GetLinguisticServices() any {
	return linguisticServices
}
