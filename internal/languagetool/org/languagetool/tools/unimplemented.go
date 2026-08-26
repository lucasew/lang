// Package tools ports org.languagetool.tools (partial, growing with tests).
package tools

// Unimplemented panics with a stable message for LT APIs not yet filled in.
// Phase 1: fail closed — never return empty success from missing logic.
func Unimplemented(what string) {
	panic("unimplemented: " + what)
}
