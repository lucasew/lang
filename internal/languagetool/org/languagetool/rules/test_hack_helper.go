package rules

import "os"

// TestHackHelper ports org.languagetool.rules.TestHackHelper for Go tests.
// IsTest reports whether the process looks like a `go test` run.
func IsTest() bool {
	// go test sets this binary suffix / env patterns
	for _, arg := range os.Args {
		if len(arg) >= 5 && (arg == "-test.v" || hasTestPrefix(arg)) {
			return true
		}
	}
	// testing package sets this when running tests
	return testingBinary()
}

func hasTestPrefix(arg string) bool {
	return len(arg) > 6 && arg[:6] == "-test."
}

func testingBinary() bool {
	// Heuristic: test binaries often end with .test
	if len(os.Args) == 0 {
		return false
	}
	exe := os.Args[0]
	return len(exe) >= 5 && (exe[len(exe)-5:] == ".test" || containsTest(exe))
}

func containsTest(s string) bool {
	return len(s) >= 5 && (s[len(s)-5:] == ".test" ||
		(len(s) > 10 && (s[len(s)-10:] == ".test.exe")))
}

// TestHackHelper is the Java-name twin for IsTest helpers.
type TestHackHelper struct{}

func (TestHackHelper) IsJUnitTest() bool { return IsTest() }
