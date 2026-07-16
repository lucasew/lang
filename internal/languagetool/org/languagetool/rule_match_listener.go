package languagetool

// RuleMatchListener ports org.languagetool.RuleMatchListener.
// Match is *rules.RuleMatch at call sites; typed as any to avoid package cycles.
type RuleMatchListener func(match any)

// MatchFound invokes the listener if non-nil.
func (l RuleMatchListener) MatchFound(match any) {
	if l != nil {
		l(match)
	}
}
