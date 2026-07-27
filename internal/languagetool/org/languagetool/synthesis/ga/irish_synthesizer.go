package ga

// IrishSynthesizer Java twin name file — type in ga_synthesizer.go;
// Instance / open path in open_irish_synthesizer.go.
//
// INSTANCE ports IrishSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *IrishSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *IrishSynthesizer) {
	INSTANCE = s
}
