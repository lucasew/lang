package rules

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// PartialPosTagFilter ports org.languagetool.rules.PartialPosTagFilter without a tagger.
// Tag maps a partial token to POS tags; when nil, Accept returns false (fail-closed).
type PartialPosTagFilter struct {
	// Tag returns POS tags for the extracted partial token.
	// Nil or empty → no match (Java tag() returning null drops the rule match).
	Tag func(partial string) []string
}

func NewPartialPosTagFilter(tag func(string) []string) *PartialPosTagFilter {
	return &PartialPosTagFilter{Tag: tag}
}

// AcceptRuleMatch ports PartialPosTagFilter.acceptRuleMatch.
// Args: no (1-based pattern token), regexp, postag_regexp; optional negate_pos, two_groups_regexp, prefix, suffix.
func (f *PartialPosTagFilter) AcceptRuleMatch(match *RuleMatch, args map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if args["no"] == "" || args["regexp"] == "" || args["postag_regexp"] == "" {
		panic("Set 'no', 'regexp' and 'postag_regexp' for filter PartialPosTagFilter")
	}
	tokenPos, err := strconv.Atoi(args["no"])
	if err != nil {
		return nil
	}
	if tokenPos < 1 || tokenPos > len(patternTokens) || patternTokens[tokenPos-1] == nil {
		return nil
	}
	_, negatePos := args["negate_pos"]
	_, twoGroups := args["two_groups_regexp"]
	ok, err := f.Accept(
		patternTokens[tokenPos-1].GetToken(),
		args["regexp"],
		args["postag_regexp"],
		negatePos,
		twoGroups,
		args["prefix"],
		args["suffix"],
	)
	if err != nil {
		// Java throws RuntimeException on wrong group count.
		panic(err.Error())
	}
	if !ok {
		return nil
	}
	return match
}

// Accept keeps the match when the partial token's POS matches postagRegexp.
// token is the full pattern token; regexp extracts group(s); negatePos inverts the check.
// Java: Matcher.matches() on full prefix+token+suffix; groupCount from pattern.
func (f *PartialPosTagFilter) Accept(token, tokenRegexp, postagRegexp string, negatePos, twoGroups bool, prefix, suffix string) (bool, error) {
	if f == nil || f.Tag == nil {
		return false, nil
	}
	re, err := regexp.Compile(tokenRegexp)
	if err != nil {
		return false, err
	}
	// Java checks Pattern groupCount before match (throws on mismatch).
	groups := re.NumSubexp()
	if twoGroups {
		if groups != 2 {
			return false, fmt.Errorf("Got %d groups for regex '%s', expected 2", groups, tokenRegexp)
		}
	} else if groups != 1 {
		return false, fmt.Errorf("Got %d groups for regex '%s', expected 1", groups, tokenRegexp)
	}
	full := prefix + token + suffix
	// Java matcher.matches() — entire string, not Find substring.
	m := re.FindStringSubmatch(full)
	if m == nil || m[0] != full {
		return false, nil
	}
	partial := m[1]
	if twoGroups {
		partial += m[2]
	}
	tags := f.Tag(partial)
	return partialTagHasRequiredTag(tags, postagRegexp, negatePos), nil
}

func partialTagHasRequiredTag(tags []string, requiredTagRegexp string, negatePos bool) bool {
	re, err := regexp.Compile(requiredTagRegexp)
	if err != nil {
		return false
	}
	postagCount := 0
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		// Java String.matches: entire POS tag must match the regexp.
		full := reFullMatchString(re, tag)
		if negatePos {
			postagCount++
			if full {
				return false
			}
		} else if full {
			return true
		}
	}
	if postagCount == 0 {
		return false
	}
	return negatePos
}

// reFullMatchString ports Matcher.matches() against compiled re.
func reFullMatchString(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}
