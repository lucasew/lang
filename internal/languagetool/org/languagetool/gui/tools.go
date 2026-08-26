package gui

import (
	"regexp"
	"strings"
)

var ampLabelRE = regexp.MustCompile(`&([^&])`)

// ShortenComment ports Tools.shortenComment — trims bracketed tails for ~100-char menus.
// Length uses Java String.length semantics (UTF-16 code units); tests are BMP/ASCII.
func ShortenComment(comment string) string {
	const maxCommentLength = 100
	short := comment
	if len(short) <= maxCommentLength {
		return short
	}
	for len(short) > maxCommentLength {
		li := strings.LastIndex(short, " [")
		ri := strings.LastIndexByte(short, ']')
		if li > 0 && ri > li {
			short = short[:li] + short[ri+1:]
			continue
		}
		li = strings.LastIndex(short, " (")
		ri = strings.LastIndexByte(short, ')')
		if li > 0 && ri > li {
			short = short[:li] + short[ri+1:]
			continue
		}
		if len(short) > maxCommentLength {
			short = short[:maxCommentLength-1] + "…"
		}
		break
	}
	return short
}

// GetLabel ports Tools.getLabel — strips single-& mnemonics, maps && → &.
func GetLabel(label string) string {
	s := ampLabelRE.ReplaceAllString(label, "$1")
	return strings.ReplaceAll(s, "&&", "&")
}

// GetOOoLabel historically aliases GetLabel for && handling.
func GetOOoLabel(label string) string { return GetLabel(label) }

// GetMnemonic ports Tools.getMnemonic.
func GetMnemonic(label string) rune {
	pos := strings.IndexByte(label, '&')
	for pos != -1 && pos == strings.Index(label, "&&") && pos < len(label) {
		if pos+2 >= len(label) {
			return 0
		}
		next := strings.IndexByte(label[pos+2:], '&')
		if next < 0 {
			pos = -1
			break
		}
		pos = pos + 2 + next
	}
	if pos < 0 || pos+1 >= len(label) {
		return 0
	}
	return rune(label[pos+1])
}
