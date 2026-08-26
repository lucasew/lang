package rules

// RuleMatchListener ports org.languagetool.RuleMatchListener.
// Called for every RuleMatch found (useful for streaming partial results).
type RuleMatchListener interface {
	MatchFound(match *RuleMatch)
}

// RuleMatchListenerFunc adapts a function to RuleMatchListener.
type RuleMatchListenerFunc func(match *RuleMatch)

func (f RuleMatchListenerFunc) MatchFound(match *RuleMatch) {
	if f != nil {
		f(match)
	}
}

// NotifyListeners fans out a match to zero or more listeners.
func NotifyListeners(match *RuleMatch, listeners ...RuleMatchListener) {
	for _, l := range listeners {
		if l != nil {
			l.MatchFound(match)
		}
	}
}
