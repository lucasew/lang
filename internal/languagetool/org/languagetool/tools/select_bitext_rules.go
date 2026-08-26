package tools

// BitextRuleID is the minimal surface for Tools.selectBitextRules (Java BitextRule.getId).
type BitextRuleID interface {
	GetID() string
}

// SelectBitextRules ports org.languagetool.tools.Tools.selectBitextRules.
//
// Java algorithm (bug-for-bug):
//   - useEnabledOnly: for each enabled id, mark every rule whose id differs for removal
//     (multiple enabled ids can remove all rules — upstream quirk).
//   - else: remove rules whose id is listed in disabledRules.
// Returns a new slice; does not mutate bRules.
func SelectBitextRules[T BitextRuleID](bRules []T, disabledRules, enabledRules []string, useEnabledOnly bool) []T {
	if len(bRules) == 0 {
		return nil
	}
	// Java: newBRules = new ArrayList<>(bRules); addAll(bRules)
	out := append([]T(nil), bRules...)
	toDisable := map[int]struct{}{}
	if useEnabledOnly {
		// Java iterates enabledRules × bRules (original list, not newBRules).
		for _, enabledRule := range enabledRules {
			for i, b := range bRules {
				if b.GetID() != enabledRule {
					toDisable[i] = struct{}{}
				}
			}
		}
	} else {
		// Java: for disabledRule : disabledRules { for b : newBRules if id equals }
		disabled := map[string]struct{}{}
		for _, id := range disabledRules {
			disabled[id] = struct{}{}
		}
		for i, b := range out {
			if _, off := disabled[b.GetID()]; off {
				toDisable[i] = struct{}{}
			}
		}
	}
	if len(toDisable) == 0 {
		return out
	}
	kept := make([]T, 0, len(out)-len(toDisable))
	for i, b := range out {
		if _, drop := toDisable[i]; drop {
			continue
		}
		kept = append(kept, b)
	}
	return kept
}
