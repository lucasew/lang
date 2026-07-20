package languagetool

// Experimental marks an experimental feature that may change without warning
// in future versions. Ports org.languagetool.Experimental (Java annotation).
//
// Go has no runtime annotations; callers document experimental APIs with this
// type name / comments. Presence of this package-level marker mirrors the
// Java @interface for twin discovery and docs.
type Experimental struct{}
