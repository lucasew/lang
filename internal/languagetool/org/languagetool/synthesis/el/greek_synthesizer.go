package el

// GreekSynthesizer Java twin name file — type in el_synthesizer.go;
// Instance / open path in open_greek_synthesizer.go.
//
// INSTANCE ports GreekSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *GreekSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *GreekSynthesizer) {
	INSTANCE = s
}
