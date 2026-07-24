package languagetool

// ApiCleanupNeeded marks places where the API would need to be cleaned up.
// Ports org.languagetool.ApiCleanupNeeded (Java annotation with value()).
//
// Go has no runtime annotations; Value holds the Java annotation message when
// a call site documents the marker in code.
type ApiCleanupNeeded struct {
	// Value is the cleanup note (Java: ApiCleanupNeeded.value()).
	Value string
}
