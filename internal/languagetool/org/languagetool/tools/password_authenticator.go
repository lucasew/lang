package tools

// SetPasswordAuthenticator ports Tools.setPasswordAuthenticator.
// Java: Authenticator.setDefault(new PasswordAuthenticator()) so URLs of the form
// http://user:pass@host work when loading XML; SecurityException is swallowed.
// Go has no global URL authenticator equivalent for classpath/XML loads — no-op
// (same user-visible outcome as Java when SecurityManager blocks setDefault).
func SetPasswordAuthenticator() {}
