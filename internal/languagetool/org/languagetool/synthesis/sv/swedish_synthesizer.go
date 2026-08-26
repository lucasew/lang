package sv

// SwedishSynthesizer Java twin name file — type in sv_synthesizer.go;
// Instance / open path in open_swedish_synthesizer.go.
//
// INSTANCE ports SwedishSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *SwedishSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *SwedishSynthesizer) {
	INSTANCE = s
}
