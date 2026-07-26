package it

// ItalianSynthesizer Java twin name file — type in it_synthesizer.go;
// Instance / open path in open_italian_synthesizer.go.
//
// INSTANCE ports ItalianSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *ItalianSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *ItalianSynthesizer) {
	INSTANCE = s
}
