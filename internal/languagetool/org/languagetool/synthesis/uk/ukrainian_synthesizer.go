package uk

// UkrainianSynthesizer Java twin name file — type in uk_synthesizer.go;
// Instance / open path in open_ukrainian_synthesizer.go.
//
// INSTANCE ports UkrainianSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *UkrainianSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *UkrainianSynthesizer) {
	INSTANCE = s
}
