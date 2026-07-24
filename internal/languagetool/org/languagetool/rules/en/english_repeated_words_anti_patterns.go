package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"

// EnglishRepeatedWordsAntiPatterns ports EnglishRepeatedWordsRule.ANTI_PATTERNS (24/24).
// matchInflectedForms needs lemma readings; POS patterns need tags (fail-closed otherwise).
var EnglishRepeatedWordsAntiPatterns = [][]*patterns.PatternToken{
	// need to
	{
		patterns.NewPatternTokenBuilder().CsToken("need").MatchInflectedForms().Build(),
		patterns.Token("to"),
	},
	// solve … problems?
	{
		patterns.NewPatternTokenBuilder().TokenRegex("solve(s|d|ing)?").SetSkip(3).Build(),
		patterns.TokenRegex("problems?"),
	},
	// No problem,
	{
		patterns.PosRegex("SENT_START|PCT"),
		patterns.Token("no"),
		patterns.Token("problem"),
		patterns.Pos("PCT"),
	},
	// math/word problem
	{
		patterns.TokenRegex("math|word"),
		patterns.TokenRegex("problems?"),
	},
	// as a whole
	{
		patterns.TokenRegex("as"),
		patterns.TokenRegex("a"),
		patterns.TokenRegex("whole"),
	},
	// more often than not
	{
		patterns.Token("more"),
		patterns.Token("often"),
		patterns.Token("than"),
		patterns.Token("not"),
	},
	// often times
	{
		patterns.Token("often"),
		patterns.Token("times"),
	},
	// … suggest (after details/facts/…)
	{
		patterns.TokenRegex("details?|facts?|it|journals?|questions?|research|results?|study|studies|this|these|those|which"),
		patterns.NewPatternTokenBuilder().Pos("RB").Min(0).Build(),
		patterns.NewPatternTokenBuilder().CsToken("suggest").MatchInflectedForms().Build(),
	},
	// form + prep/punct
	{
		patterns.NewPatternTokenBuilder().CsToken("form").MatchInflectedForms().Build(),
		patterns.PosRegex("IN|PCT|RP|TO|SENT_END"),
	},
	// bonds… form
	{
		patterns.NewPatternTokenBuilder().TokenRegex("bonds?|crystals?|ions?|rocks?|.*valence").SetSkip(10).Build(),
		patterns.NewPatternTokenBuilder().CsToken("form").MatchInflectedForms().Build(),
	},
	// form… bonds
	{
		patterns.NewPatternTokenBuilder().TokenRegex("form(s|ed|ing)?").SetSkip(10).Build(),
		patterns.TokenRegex("bonds?|crystals?|ions?|rocks?|.*valence"),
	},
	// interesting facts/things
	{
		patterns.Token("interesting"),
		patterns.TokenRegex("facts?|things?"),
	},
	// several hundreds/thousands/millions
	{
		patterns.Token("several"),
		patterns.TokenRegex("hundreds?|thousands?|millions?"),
	},
	// must be nice
	{
		patterns.Token("must"),
		patterns.Token("be"),
		patterns.Token("nice"),
	},
	// nice day
	{
		patterns.Token("nice"),
		patterns.Token("day"),
	},
	// nice to meet? PRP_O
	{
		patterns.Token("nice"),
		patterns.Token("to"),
		patterns.NewPatternTokenBuilder().Token("meet").Min(0).Build(),
		patterns.PosRegex("PRP_O.*"),
	},
	// be nice and JJ
	{
		patterns.NewPatternTokenBuilder().CsToken("be").MatchInflectedForms().Build(),
		patterns.Token("nice"),
		patterns.Token("and"),
		patterns.Pos("JJ"),
		patterns.PosRegex("PCT|SENT_END"),
	},
	// proposed N
	{
		patterns.PosRegex("P?DT|PRP$.*"),
		patterns.Token("proposed"),
		patterns.PosRegex("N.*"),
	},
	// propose to|marriage
	{
		patterns.NewPatternTokenBuilder().CsToken("propose").MatchInflectedForms().Build(),
		patterns.TokenRegex("to|marriage"),
	},
	// too literally
	{
		patterns.Token("too"),
		patterns.Token("literally"),
	},
	// literally and figuratively
	{
		patterns.Token("literally"),
		patterns.Token("and"),
		patterns.Token("figuratively"),
	},
	// literally everything
	{
		patterns.Token("literally"),
		patterns.Token("everything"),
	},
	// literally + punct
	{
		patterns.Token("literally"),
		patterns.PosRegex("PCT|SENT_END"),
	},
	// CC maybe
	{
		patterns.PosRegex("CC"),
		patterns.Token("maybe"),
	},
}
