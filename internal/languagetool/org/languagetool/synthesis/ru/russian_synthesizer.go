package ru

// RussianSynthesizer Java twin name file — type in ru_synthesizer.go;
// Instance / open path in open_russian_synthesizer.go.
//
// INSTANCE ports RussianSynthesizer.INSTANCE (lazy official load via Instance()).
// Prefer Instance(); INSTANCE is kept in sync on successful load for Java-field naming.
var INSTANCE *RussianSynthesizer

// bindINSTANCE is called from Instance() once so INSTANCE matches Java singleton field name.
func bindINSTANCE(s *RussianSynthesizer) {
	INSTANCE = s
}
