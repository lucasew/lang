package ro

// RomanianSynthesizer Java twin name file — type in ro_synthesizer.go;
// Instance / open path in open_romanian_synthesizer.go.
//
// INSTANCE ports RomanianSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *RomanianSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *RomanianSynthesizer) {
	INSTANCE = s
}
