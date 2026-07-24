package tools

import "sync"

// StringInterner ports org.languagetool.tools.StringInterner with a process-local pool.
// Go strings are already immutable; interning reduces duplicate allocations for hot keys.
var (
	internMu   sync.Mutex
	internPool = map[string]string{}
)

// Intern returns a canonical instance of s.
func Intern(s string) string {
	internMu.Lock()
	defer internMu.Unlock()
	if v, ok := internPool[s]; ok {
		return v
	}
	internPool[s] = s
	return s
}
