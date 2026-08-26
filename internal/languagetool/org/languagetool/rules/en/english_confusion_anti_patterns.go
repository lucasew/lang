package en

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// EnglishConfusionAntiPatterns ports EnglishConfusionProbabilityRule.ANTI_PATTERNS (32/32).
var EnglishConfusionAntiPatterns = [][]*patterns.PatternToken{
	{ // "Those wee changes made a big difference"
		patterns.TokenRegex("the|these|those"),
		patterns.Token("wee"),
		patterns.PosRegex("NNS"),
	},
	{ // "...click on the icon, from there turn it to standard."
		patterns.Token("from"),
		patterns.Token("there"),
		patterns.TokenRegex("turn|walk|go|drive"),
	},
	{ // "They told me that they got there first."
		patterns.PosRegex("VB.*"),
		patterns.Token("there"),
		patterns.Token("first"),
	},
	{ // Or go to the individual site and then click on the icon, from there turn it to standard.
		patterns.PosRegex("IN|CC"),
		patterns.Token("there"),
		patterns.PosRegex("VB"),
		patterns.PosRegex("PRP_O.*|DT"),
	},
	{ // "I just can't tell them no when they look at me with those puppy dog eyes"
		patterns.TokenRegex("tells?|told|telling|answers?|answering|answered|reply|replies|replied|replying"),
		patterns.TokenRegex("them|him|her"),
		patterns.Token("no"),
	},
	{ // way vs was: This way a person could learn ....
		patterns.Token("this"),
		patterns.Token("way"),
		patterns.PosRegex("DT|PRP\\$"),
		patterns.PosRegex("NN.*"),
	},
	{ // sense vs since: Neubauer has been a youth ambassador of the non-governmental organization ONE since 2016.
		patterns.Token("since"),
		patterns.TokenRegex("\\d{1,2}|\\d{4}"),
	},
	{ // way vs was
		patterns.Token("in"),
		patterns.Token("a"),
		patterns.TokenRegex("more|less"),
		patterns.PosRegex("JJ"),
		patterns.Token("way"),
	},
	{ // he wondered what way the country is ...
		patterns.Token("what"),
		patterns.Token("way"),
		patterns.PosRegex("DT|PRP\\$"),
		patterns.PosRegex("NN.*"),
		patterns.TokenRegex("is|was|were|has|have|had"),
	},
	{ // This way Columbus could do ...
		patterns.Token("this"),
		patterns.Token("way"),
		patterns.PosRegex("NNP|PRP"),
		patterns.PosRegex("VB.*|MD"),
	},
	{ // This way only he ...
		patterns.Token("this"),
		patterns.Token("way"),
		patterns.Token("only"),
		patterns.PosRegex("DT|PRP.*"),
	},
	{ // This way neither of you ...
		patterns.Token("this"),
		patterns.Token("way"),
		patterns.TokenRegex("none|neither"),
		patterns.Token("of"),
	},
	{ // This way Christopher Columbus could do ...
		patterns.Token("this"),
		patterns.Token("way"),
		patterns.PosRegex("NNP"),
		patterns.PosRegex("NNP"),
		patterns.PosRegex("VB.*|MD"),
	},
	{ // way vs was: This way a person could learn ....
		patterns.Token("way"),
		patterns.Token("too"),
	},
	{ // "from ... to ..." (to/the)
		patterns.PosRegex("NNP|UNKNOWN"),
		patterns.TokenRegex("to"),
		patterns.PosRegex("NNP|UNKNOWN"),
	},
	{ // "Meltzer taught Crim for Section 5 last year." (taught/thought)
		patterns.PosRegex("NNP|UNKNOWN"),
		patterns.TokenRegex("taught|threw"),
		patterns.PosRegex("NNP|UNKNOWN"),
	},
	{ // "Were you never taught to say your prayers?" (taught/thought)
		patterns.Token("taught"),
		patterns.Token("to"),
		patterns.TokenRegex("say|do|make|be|become|treat"),
	},
	{ // "way easier" (was/way)
		patterns.Token("way"),
		patterns.PosRegex("JJR|RBR"),
	},
	{ // "He acts way different" (was/way)
		patterns.PosRegex("VB.*"),
		patterns.Token("way"),
		patterns.Token("different"),
	},
	{ // "way much easier" (was/way)
		patterns.Token("way"),
		patterns.Token("much"),
		patterns.PosRegex("JJR"),
	},
	{ // "way out of" (was/way)
		patterns.Token("way"),
		patterns.Token("out"),
		patterns.TokenRegex("of|in|on"),
	},
	{ // "way to long" (was/way)
		patterns.Token("way"),
		patterns.Token("to"),
		patterns.PosRegex("JJ"),
	},
	{ // "He was there way before" (was/way)
		patterns.Token("way"),
		patterns.TokenRegex("before|after|outside|inside|back"),
	},
	{ // "In a logic way" (was/way)
		patterns.Token("in"),
		patterns.TokenRegex("an?"),
		patterns.PosRegex("JJ"),
		patterns.Token("way"),
	},
	{ // They "awarded" us a contract ...
		patterns.PosRegex("VB.*"),
		patterns.TokenRegex("[\"”“]"),
		patterns.Token("us"),
	},
	{ // Text us at (410) 4535
		patterns.TokenRegex("message(s|d)?|text(s|ed)?|DM"),
		patterns.Token("us"),
		patterns.PosRegex("PCT|IN|TO|CC|DT"),
	},
	{ // Clinton will pay us based on actuals.
		patterns.PosRegex("VB.*"),
		patterns.Token("us"),
		patterns.TokenRegex("based|depending"),
		patterns.Token("on"),
	},
	{ // How Apple and FB do events:
		patterns.Token("how"),
		patterns.PosRegex("NNP"),
		patterns.TokenRegex("and|&"),
		patterns.PosRegex("NNP"),
		patterns.Token("do"),
	},
	{ // ...responded quickly to Mustaf Sheikh's request to wear his religious head gear, use break time for prayer, and combine his breaks for Friday attendance...
		patterns.Token("break"),
		patterns.NewPatternTokenBuilder().TokenRegex("times?").SetSkip(10).Build(),
		patterns.Token("breaks"),
	},
	{ // The language added in 3.5 was an attempt to show how much EEFT does vs. EnronOnline.
		patterns.TokenRegex("how|what"),
		patterns.NewPatternTokenBuilder().TokenRegex("much|many").Min(0).SetSkip(3).Build(),
		patterns.Token("does"),
	},
	{ // Maybe we should pull over and dose him.
		patterns.PosRegex("MD"),
		patterns.PosRegex("VB"),
		patterns.PosRegex("IN|RP"),
		patterns.Token("and"),
		patterns.Token("dose"),
	},
	{ // But your Mother's constant implications that your some big victim and I'm a very big asshole has worn very thin and it's inappropriate and rude.
		patterns.Token("your"),
		patterns.PosRegex("DT"),
	},
}

var (
	enConfusionAPOnce  sync.Once
	enConfusionAPRules []*disambigrules.DisambiguationPatternRule
)

func englishConfusionAntiPatternRules() []*disambigrules.DisambiguationPatternRule {
	enConfusionAPOnce.Do(func() {
		aps := EnglishConfusionAntiPatterns
		enConfusionAPRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "en",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			enConfusionAPRules = append(enConfusionAPRules, rule)
		}
	})
	return enConfusionAPRules
}

func englishConfusionSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := englishConfusionAntiPatternRules()
	if len(aps) == 0 {
		return sentence
	}
	src := sentence.GetTokens()
	cloned := make([]*languagetool.AnalyzedTokenReadings, len(src))
	for i, t := range src {
		if t == nil {
			continue
		}
		cloned[i] = languagetool.NewAnalyzedTokenReadingsFromOld(t, t.GetReadings(), "")
	}
	immunized := languagetool.NewAnalyzedSentence(cloned)
	for _, ap := range aps {
		if ap != nil {
			immunized = ap.Replace(immunized)
		}
	}
	return immunized
}

// enConfusionIsCoveredByAntiPattern ports ConfusionProbabilityRule.isCoveredByAntiPattern
// via DisambiguationPatternRule immunization (Java getSentenceWithImmunization).
func enConfusionIsCoveredByAntiPattern(sentence *languagetool.AnalyzedSentence, startPos, endPos int) bool {
	imm := englishConfusionSentenceWithImmunization(sentence)
	if imm == nil {
		return false
	}
	for _, t := range imm.GetTokensWithoutWhitespace() {
		if t == nil || !t.IsImmunized() {
			continue
		}
		if t.GetStartPos() <= startPos && t.GetEndPos() >= endPos {
			return true
		}
	}
	return false
}

