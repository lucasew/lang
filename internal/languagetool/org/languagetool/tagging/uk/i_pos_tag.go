package uk

import "strings"

// IPOSTag ports org.languagetool.tagging.uk.IPOSTag.
type IPOSTag string

const (
	IPOSNoun    IPOSTag = "noun"
	IPOSAdj     IPOSTag = "adj"
	IPOSVerb    IPOSTag = "verb"
	IPOSAdv     IPOSTag = "adv"
	IPOSPart    IPOSTag = "part"
	IPOSIntj    IPOSTag = "intj"
	IPOSNumr    IPOSTag = "numr"
	IPOSNumber  IPOSTag = "number"
	IPOSDate    IPOSTag = "date"
	IPOSTime    IPOSTag = "time"
	IPOSAdvp    IPOSTag = "advp"
	IPOSPrep    IPOSTag = "prep"
	IPOSPredic  IPOSTag = "predic"
	IPOSInsert  IPOSTag = "insert"
	IPOSAbbr    IPOSTag = "abbr"
	IPOSBad     IPOSTag = "bad"
	IPOSOnomat  IPOSTag = "onomat"
	IPOSHashtag IPOSTag = "hashtag"
)

func (t IPOSTag) Text() string { return string(t) }

// Match reports whether posTagPrefix starts with this tag's name.
func (t IPOSTag) Match(posTagPrefix string) bool {
	return posTagPrefix != "" && strings.HasPrefix(posTagPrefix, string(t))
}

// IsNum reports numr or number tags.
func IsNum(posTag string) bool {
	return IPOSNumr.Match(posTag) || IPOSNumber.Match(posTag)
}

// POSContains reports substring presence.
func POSContains(posTag, postagMatch string) bool {
	return posTag != "" && strings.Contains(posTag, postagMatch)
}

// POSStartsWithAny reports if prefix starts with any of the given tags.
func POSStartsWithAny(posTagPrefix string, posTags ...IPOSTag) bool {
	if posTagPrefix == "" {
		return false
	}
	for _, t := range posTags {
		if strings.HasPrefix(posTagPrefix, t.Text()) {
			return true
		}
	}
	return false
}
