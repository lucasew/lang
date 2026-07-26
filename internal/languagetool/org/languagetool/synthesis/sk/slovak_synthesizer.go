package sk

// SlovakSynthesizer Java twin name file — type in sk_synthesizer.go;
// Instance / open path in open_slovak_synthesizer.go.
//
// INSTANCE ports SlovakSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *SlovakSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *SlovakSynthesizer) {
	INSTANCE = s
}
