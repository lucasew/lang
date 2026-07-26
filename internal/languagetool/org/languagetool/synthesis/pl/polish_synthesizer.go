package pl

// PolishSynthesizer Java twin name file — type/overrides in pl_synthesizer.go;
// Instance / open path in open_polish_synthesizer.go.
//
// INSTANCE ports PolishSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *PolishSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *PolishSynthesizer) {
	INSTANCE = s
}
