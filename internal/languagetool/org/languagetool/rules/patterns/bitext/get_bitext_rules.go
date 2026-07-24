package bitext

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/bitext"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// GetBitextRules ports Tools.getBitextRules(source, target, externalFile).
// Order matches Java:
//  1. /{target}/bitext.xml pattern rules when present
//  2. external bitext rule file when path non-empty
//  3. false-friends.xml as bitext (source↔target)
//  4. builtin BitextRule classes (DifferentLength / SameTranslation / DifferentPunctuation)
//
// falseFriendsXML is the expanded false-friends document (or empty to skip).
// bitextXML is optional target bitext.xml content (or empty).
// externalBitext is optional external bitext rule file content.
func GetBitextRules(sourceLang, targetLang string, bitextXML, falseFriendsXML, externalBitext string) ([]bitext.BitextRule, error) {
	var out []bitext.BitextRule

	// 1) target bitext.xml
	if tools.JavaStringTrim(bitextXML) != "" {
		loader := NewBitextPatternRuleLoader()
		rules, err := loader.GetRules(strings.NewReader(bitextXML), targetLang+"/bitext.xml")
		if err != nil {
			return nil, err
		}
		for _, r := range rules {
			if r != nil {
				out = append(out, r)
			}
		}
	}

	// 2) external bitext file
	if tools.JavaStringTrim(externalBitext) != "" {
		loader := NewBitextPatternRuleLoader()
		rules, err := loader.GetRules(strings.NewReader(externalBitext), "external-bitext.xml")
		if err != nil {
			return nil, err
		}
		for _, r := range rules {
			if r != nil {
				out = append(out, r)
			}
		}
	}

	// 3) false friends as bitext
	if tools.JavaStringTrim(falseFriendsXML) != "" {
		ff := NewFalseFriendsAsBitextLoader()
		rules, err := ff.GetFalseFriendsAsBitext(
			strings.NewReader(falseFriendsXML),
			strings.NewReader(falseFriendsXML),
			sourceLang, targetLang,
		)
		if err != nil {
			return nil, err
		}
		for _, r := range rules {
			if r != nil {
				out = append(out, r)
			}
		}
	}

	// 4) builtin Java BitextRule classes
	out = append(out, bitext.RelevantBitextRules()...)
	return out, nil
}

// LoadGetBitextRulesFromPaths ports Tools.getBitextRules with filesystem resources.
// bitextXMLPath / falseFriendsPath / externalPath may be empty to skip that stage.
func LoadGetBitextRulesFromPaths(sourceLang, targetLang, bitextXMLPath, falseFriendsPath, externalPath string) ([]bitext.BitextRule, error) {
	read := func(p string) (string, error) {
		if p == "" {
			return "", nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		// Expand SYSTEM .ent when loading false-friends / bitext from disk
		return string(patterns.ExpandLTXMLEntitiesWithBase(filepath.Dir(p), b)), nil
	}
	bitextXML, err := read(bitextXMLPath)
	if err != nil {
		return nil, err
	}
	ff, err := read(falseFriendsPath)
	if err != nil {
		return nil, err
	}
	ext, err := read(externalPath)
	if err != nil {
		return nil, err
	}
	return GetBitextRules(sourceLang, targetLang, bitextXML, ff, ext)
}
