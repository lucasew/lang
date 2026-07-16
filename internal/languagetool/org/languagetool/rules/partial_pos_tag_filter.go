package rules

import "regexp"

// PartialPosTagFilter ports org.languagetool.rules.PartialPosTagFilter without a tagger.
// Tag maps a partial token to POS tags; when nil, Accept returns false.
type PartialPosTagFilter struct {
	// Tag returns POS tags for the extracted partial token.
	Tag func(partial string) []string
}

func NewPartialPosTagFilter(tag func(string) []string) *PartialPosTagFilter {
	return &PartialPosTagFilter{Tag: tag}
}

// Accept keeps the match when the partial token's POS matches postagRegexp.
// token is the full pattern token; regexp extracts group(s); negatePos inverts the check.
func (f *PartialPosTagFilter) Accept(token, tokenRegexp, postagRegexp string, negatePos, twoGroups bool, prefix, suffix string) (bool, error) {
	if f.Tag == nil {
		return false, nil
	}
	re, err := regexp.Compile(tokenRegexp)
	if err != nil {
		return false, err
	}
	full := prefix + token + suffix
	m := re.FindStringSubmatch(full)
	if m == nil {
		return false, nil
	}
	// groupCount for FindStringSubmatch is len(m)-1
	groups := len(m) - 1
	if twoGroups {
		if groups != 2 {
			return false, nil
		}
	} else if groups != 1 {
		return false, nil
	}
	partial := m[1]
	if twoGroups && groups == 2 {
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
		if negatePos {
			postagCount++
			if re.MatchString(tag) {
				return false
			}
		} else if re.MatchString(tag) {
			return true
		}
	}
	if postagCount == 0 {
		return false
	}
	return negatePos
}
