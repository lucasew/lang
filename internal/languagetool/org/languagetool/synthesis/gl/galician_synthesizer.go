package gl

// GalicianSynthesizer Java twin name file — type in gl_synthesizer.go;
// Instance / open path in open_galician_synthesizer.go.
//
// INSTANCE ports GalicianSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *GalicianSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *GalicianSynthesizer) {
	INSTANCE = s
}
