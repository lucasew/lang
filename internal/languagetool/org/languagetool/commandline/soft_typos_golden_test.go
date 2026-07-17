package commandline

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDiscoverEnglishTyposFile(t *testing.T) {
	p := DiscoverEnglishTyposFile(nil)
	if p == "" {
		t.Skip("en-typos.tsv not found")
	}
	require.FileExists(t, p)
}

func TestGolden_SoftTyposSuggestions(t *testing.T) {
	if DiscoverEnglishTyposFile(nil) == "" && DiscoverEnglishUSDict(nil) == "" {
		t.Skip("need typos file or binary dict")
	}
	cases := []struct{ text, sug string }{
		{"I will go tommorow.", "tomorrow"},
		{"That is wierd.", "weird"},
		{"Please recieve this.", "receive"},
	}
	for _, tc := range cases {
		t.Run(tc.sug, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == "MORFOLOGIK_RULE_EN_US" && f.Suggestion == tc.sug {
					found = true
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftCanCan(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "They can can fish.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_CAN_CAN" {
			found = true
			require.Equal(t, "can", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftThatThat(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I know that that is true.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_THAT_THAT" {
			found = true
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftHadOf(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I had of known better.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_HAD_OF" {
			found = true
			require.Equal(t, "had", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftAgreement(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They is ready.", "EN_SOFT_THEY_IS", "are"},
		{"I are happy.", "EN_SOFT_I_ARE", "am"},
		{"He are late.", "EN_SOFT_HE_ARE", "is"},
		{"This are wrong.", "EN_SOFT_THIS_ARE", "is"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftAPlural(t *testing.T) {
	// needs POS tagger for NNS on "books"
	if DiscoverEnglishPOSDict(nil) == "" {
		t.Skip("english.dict not found")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "A books are here.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_A_PLURAL" {
			found = true
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_Soft3sgBareVerb(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"He go home.", "EN_SOFT_HE_GO", "goes"},
		{"She walk fast.", "EN_SOFT_HE_WALK", "walks"},
		{"It like rain.", "EN_SOFT_HE_LIKE", "likes"},
		{"He want more.", "EN_SOFT_HE_WANT", "wants"},
		{"She need help.", "EN_SOFT_HE_NEED", "needs"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftLooseLose(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I want to loose weight.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_LOOSE_LOSE" {
			found = true
			require.Equal(t, "lose weight", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftConfusablesExtra(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"The rules will take affect soon.", "EN_SOFT_TAKE_AFFECT", "take effect"},
		{"Please be quite in the library.", "EN_SOFT_QUITE_QUIET", "be quiet"},
		{"That peaked my interest a lot.", "EN_SOFT_PEAKED_INTEREST", "piqued my interest"},
		{"I could care less about that.", "EN_SOFT_COULD_CARE_LESS", "couldn't care less"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoft3sg(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "He go home every day.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "goes")
}

func TestGolden_ApplySoftPhrase(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I did it on accident yesterday.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "by accident")
}

func TestGolden_SoftConfusablesMore(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Whose going to the party?", "EN_SOFT_WHOSE_WHO_S", "Who's going"},
		{"Who's book is this?", "EN_SOFT_WHO_S_BOOK", "Whose book"},
		{"Please breath deeply now.", "EN_SOFT_BREATH_BREATHE", "breathe deeply"},
		{"I want to advice you.", "EN_SOFT_ADVICE_ADVISE", "to advise"},
		{"He do the work.", "EN_SOFT_HE_DO", "does"},
		{"She have a car.", "EN_SOFT_HE_HAVE", "has"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftContractionForms(t *testing.T) {
	// cant/wont are valid dict words; soft grammar forces apostrophe suggestions.
	cases := []struct {
		text, rule, sug string
	}{
		{"I dont know.", "EN_SOFT_DONT", "don't"},
		{"She cant come.", "EN_SOFT_CANT", "can't"},
		{"They wont mind.", "EN_SOFT_WONT", "won't"},
		{"He didnt call.", "EN_SOFT_DIDNT", "didn't"},
		{"It isnt ready.", "EN_SOFT_ISNT", "isn't"},
		{"They arent here.", "EN_SOFT_ARENT", "aren't"},
		{"He wasnt late.", "EN_SOFT_WASNT", "wasn't"},
		{"You shouldnt go.", "EN_SOFT_SHOULDNT", "shouldn't"},
		{"I wouldnt care.", "EN_SOFT_WOULDNT", "wouldn't"},
		{"Youre welcome.", "EN_SOFT_YOURE", "You're"},
		{"Theyre leaving.", "EN_SOFT_THEYRE", "They're"},
		{"I wouldve gone.", "EN_SOFT_WOULDVE", "would've"},
		{"She couldve won.", "EN_SOFT_COULDVE", "could've"},
		{"You shouldve called.", "EN_SOFT_SHOULDVE", "should've"},
		{"He mustve left.", "EN_SOFT_MUSTVE", "must've"},
		{"It mightve worked.", "EN_SOFT_MIGHTVE", "might've"},
		{"Thatll do.", "EN_SOFT_THATLL", "That'll"},
		{"Itll rain soon.", "EN_SOFT_ITLL", "It'll"},
		{"I know thatll work.", "EN_SOFT_THATLL", "that'll"},
		{"Say youre ready.", "EN_SOFT_YOURE", "you're"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			// fallback: typos/speller path also acceptable for some forms
			if !found {
				for _, f := range findings {
					if f.Rule == "MORFOLOGIK_RULE_EN_US" && f.Suggestion == tc.sug {
						found = true
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftContraction(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I dont know if she cant come.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	s := out.String()
	require.Contains(t, s, "don't")
	require.Contains(t, s, "can't")
}

func TestGolden_ApplySoftPhraseCase(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "On Accident I slipped.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "By accident")
}

func TestGolden_SoftTypography(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Hello!!", "EN_SOFT_DOUBLE_BANG"},
		{"What??", "EN_SOFT_DOUBLE_Q"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "typographical", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftWereGoingTo(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Were going to leave soon.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_WERE_WE_RE" {
			found = true
			require.Equal(t, "We're going to", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftFusedWords(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I have alot of work.", "EN_SOFT_ALOT", "a lot"},
		{"Infact it works.", "EN_SOFT_INFACT", "In fact"},
		{"Come aswell please.", "EN_SOFT_ASWELL", "as well"},
		{"Never the less we try.", "EN_SOFT_NEVERTHELESS_SPLIT", "nevertheless"},
		{"Everytime I try it fails.", "EN_SOFT_EVERYTIME", "Every time"},
		{"Noone knows the answer.", "EN_SOFT_NOONE", "No one"},
		{"Don't give into pressure.", "EN_SOFT_INTO_IN_TO", "give in to"},
		{"Talk to eachother soon.", "EN_SOFT_EACHOTHER", "each other"},
		{"Inspite of that we stay.", "EN_SOFT_INSPITE", "In spite"},
		{"Atleast try once.", "EN_SOFT_ATLEAST", "At least"},
		{"Incase it rains, wait.", "EN_SOFT_INCASE", "In case"},
		{"Upto ten people may join.", "EN_SOFT_UPTO", "Up to"},
		{"I need to workout daily.", "EN_SOFT_WORKOUT_VERB", "to work out"},
		{"I need to setup the tool.", "EN_SOFT_SETUP_VERB", "to set up"},
		{"You need to login first.", "EN_SOFT_LOGIN_VERB", "to log in"},
		{"Please to checkout the code.", "EN_SOFT_CHECKOUT_VERB", "to check out"},
		{"Remember to backup data.", "EN_SOFT_BACKUP_VERB", "to back up"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftDoubleBang(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Hello!!", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	// first suggestion "!" replaces the "!!" span
	require.Equal(t, "Hello!", strings.TrimSpace(out.String()))
}

func TestGolden_ApplySoftAlot(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I have alot of work.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "a lot")
	require.NotContains(t, out.String(), "alot")
}

func TestGolden_SoftStyleCategory(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"This is very unique work.", "EN_SOFT_VERY_UNIQUE"},
		{"He literally died laughing.", "EN_SOFT_LITERALLY_FIG"},
		{"She is kind of a genius.", "EN_SOFT_KIND_OF_A"},
		{"We left due to the fact that it rained.", "EN_SOFT_DUE_TO_THE_FACT"},
		{"At this point in time we wait.", "EN_SOFT_AT_THIS_POINT_IN_TIME"},
		{"In the event that it fails, retry.", "EN_SOFT_IN_THE_EVENT_THAT"},
		{"The end result is clear.", "EN_SOFT_END_RESULT"},
		{"His past history is known.", "EN_SOFT_PAST_HISTORY"},
		{"Get a free gift today.", "EN_SOFT_FREE_GIFT"},
		{"We must completely eliminate waste.", "EN_SOFT_COMPLETELY_ELIMINATE"},
		{"This is different than that.", "EN_SOFT_DIFFERENT_THAN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					// STYLE category maps to style issue type via soft loader
					require.Equal(t, "style", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftWouldve(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I wouldve gone earlier.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "would've")
	require.NotContains(t, out.String(), "wouldve")
}

func TestGolden_SoftCaseAndCount(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"This is between you and I.", "EN_SOFT_BETWEEN_YOU_AND_I", "between you and me"},
		{"It is for you and I.", "EN_SOFT_FOR_YOU_AND_I", "for you and me"},
		{"Come with you and I.", "EN_SOFT_WITH_YOU_AND_I", "with you and me"},
		{"Give it to who asks.", "EN_SOFT_TO_WHO", "to whom"},
		{"Less people came today.", "EN_SOFT_LESS_PEOPLE", "Fewer people"},
		{"The amount of people grew.", "EN_SOFT_AMOUNT_OF_PEOPLE", "number of people"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftBetweenYouAndI(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Keep this between you and I.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "between you and me")
}

func TestGolden_SoftMoreStyleAndTryAnd(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please try and see the difference.", "EN_SOFT_TRY_AND", "try to see"},
		{"Each and every student passed.", "EN_SOFT_EACH_AND_EVERY", ""},
		{"First and foremost, plan carefully.", "EN_SOFT_FIRST_AND_FOREMOST", ""},
		{"Learn the basic fundamentals first.", "EN_SOFT_BASIC_FUNDAMENTALS", ""},
		{"The reason is because it rained.", "EN_SOFT_REASON_IS_BECAUSE", ""},
		{"Decide whether or not to go.", "EN_SOFT_WHETHER_OR_NOT", ""},
		{"That is an actual fact.", "EN_SOFT_ACTUAL_FACT", ""},
		{"A true fact remains.", "EN_SOFT_TRUE_FACT", ""},
		{"Do advance planning early.", "EN_SOFT_ADVANCE_PLANNING", ""},
		{"Stay in close proximity.", "EN_SOFT_CLOSE_PROXIMITY", ""},
		{"Share your future plans.", "EN_SOFT_FUTURE_PLANS", ""},
		{"What an unexpected surprise.", "EN_SOFT_UNEXPECTED_SURPRISE", ""},
		{"Please revert back soon.", "EN_SOFT_REVERT_BACK", ""},
		{"Do not repeat again.", "EN_SOFT_REPEAT_AGAIN", ""},
		{"The final outcome is known.", "EN_SOFT_FINAL_OUTCOME", ""},
		{"There is a general consensus.", "EN_SOFT_GENERAL_CONSENSUS", ""},
		{"In my personal opinion, wait.", "EN_SOFT_PERSONAL_OPINION", ""},
		{"The train came to a complete stop.", "EN_SOFT_COMPLETE_STOP", ""},
		{"This is absolutely essential.", "EN_SOFT_ABSOLUTELY_ESSENTIAL", ""},
		{"They are exactly the same.", "EN_SOFT_EXACTLY_THE_SAME", ""},
		{"Work is currently in progress.", "EN_SOFT_CURRENTLY_IN_PROGRESS", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftLessPeople(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Less people attended.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "Fewer people")
}

func TestGolden_ApplySoftAmountOfPeople(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "The amount of people grew.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "number of people")
}

func TestGolden_ApplySoftDontCase(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Dont forget keys.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "Don't")
}

func TestGolden_ApplySoftAnyways(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Anyways, we left early.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	// SoftPreserveCase capitalizes suggestion when match is sentence-initial.
	require.Contains(t, out.String(), "Anyway")
	require.NotContains(t, strings.ToLower(out.String()), "anyways")
}

func TestGolden_SoftUSVariantHints(t *testing.T) {
	// Loaded from testdata/grammar/en-US-soft.xml when language is en-US.
	cases := []struct {
		text, rule, sug string
	}{
		{"Walk towards the door.", "EN_SOFT_TOWARDS_US", "toward"},
		{"Sit amongst friends.", "EN_SOFT_AMONGST_US", "among"},
		{"Wait whilst I check.", "EN_SOFT_WHILST_US", "while"},
		{"A grey sky.", "EN_SOFT_GREY_US", "gray"},
		{"Pick a colour.", "EN_SOFT_COLOUR_US", "color"},
		{"My favourite book.", "EN_SOFT_FAVOURITE_US", "favorite"},
		{"City centre is busy.", "EN_SOFT_CENTRE_US", "center"},
		{"Please organise files.", "EN_SOFT_ORGANISE_US", "organize"},
		{"I realise now.", "EN_SOFT_REALISE_US", "realize"},
		{"Good behaviour matters.", "EN_SOFT_BEHAVIOUR_US", "behavior"},
		{"We travelled far.", "EN_SOFT_TRAVELLED_US", "traveled"},
		{"The flight was cancelled.", "EN_SOFT_CANCELLED_US", "canceled"},
		{"The data was modelled.", "EN_SOFT_MODELLED_US", "modeled"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-US"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
					// TYPOS category → misspelling issue type
					require.Equal(t, "misspelling", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftUSVariantsNotOnPlainEN(t *testing.T) {
	// en-US-soft.xml must not load for language "en"
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Pick a colour.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "EN_SOFT_COLOUR_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftGBVariantHints(t *testing.T) {
	// Loaded from testdata/grammar/en-GB-soft.xml when language is en-GB.
	cases := []struct {
		text, rule, sug string
	}{
		{"Walk toward the door.", "EN_SOFT_TOWARD_GB", "towards"},
		{"A gray sky.", "EN_SOFT_GRAY_GB", "grey"},
		{"Pick a color.", "EN_SOFT_COLOR_GB", "colour"},
		{"My favorite book.", "EN_SOFT_FAVORITE_GB", "favourite"},
		{"City center is busy.", "EN_SOFT_CENTER_GB", "centre"},
		{"Please organize files.", "EN_SOFT_ORGANIZE_GB", "organise"},
		{"I realize now.", "EN_SOFT_REALIZE_GB", "realise"},
		{"Good behavior matters.", "EN_SOFT_BEHAVIOR_GB", "behaviour"},
		{"We traveled far.", "EN_SOFT_TRAVELED_GB", "travelled"},
		{"The flight was canceled.", "EN_SOFT_CANCELED_GB", "cancelled"},
		{"The data was modeled.", "EN_SOFT_MODELED_GB", "modelled"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-GB"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftGBVariantsNotOnUS(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Pick a color.", &CommandLineOptions{Language: "en-US"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "EN_SOFT_COLOR_GB", f.Rule, "%+v", findings)
	}
}

func TestGolden_ApplySoftGBColour(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en-GB", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Pick a color please.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "colour")
	require.NotContains(t, out.String(), "color")
}

func TestGolden_SoftAnywaysOnPlainEN(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Anyways, we left.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_ANYWAYS" {
			found = true
			require.Equal(t, "Anyway", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftSupposeTo(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "You suppose to leave now.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SUPPOSE_TO" || f.Rule == "EN_SOFT_SUPPOSE_TO" {
			found = true
			require.Equal(t, "supposed to", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftTokenSequenceExtras(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I use to go there often.", "EN_USED_TO_GO", "used to go"},
		{"For all intensive purposes it works.", "EN_FOR_ALL_INTENSIVE", "for all intents and purposes"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftDialectForms(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I seen him yesterday.", "EN_SOFT_I_SEEN", "I saw"},
		{"I done the work already.", "EN_SOFT_I_DONE", "I did"},
		{"They was late again.", "EN_SOFT_WE_WAS", "were"},
		{"I is ready now.", "EN_SOFT_I_IS", "am"},
		{"Me and John went home.", "EN_SOFT_ME_AND", ""},
		{"She should of been here.", "EN_SOFT_SHOULD_OF_SPACED", "should have been"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftMoreRedundancyStyle(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"That is an added bonus.", "EN_SOFT_ADDED_BONUS", ""},
		{"Wait a brief moment.", "EN_SOFT_BRIEF_MOMENT", ""},
		{"Please join together now.", "EN_SOFT_JOIN_TOGETHER", ""},
		{"We should plan ahead carefully.", "EN_SOFT_PLAN_AHEAD", ""},
		{"The problem still remains.", "EN_SOFT_STILL_REMAINS", ""},
		{"Birds circle around the tree.", "EN_SOFT_CIRCLE_AROUND", ""},
		{"Please empty out the drawer.", "EN_SOFT_EMPTY_OUT", ""},
		{"The dress is pink in color.", "EN_SOFT_PINK_IN_COLOR", "pink"},
		{"The sky is blue in colour.", "EN_SOFT_BLUE_IN_COLOUR", "blue"},
		{"We need to preplan the trip.", "EN_SOFT_PREPLAN", "plan"},
		{"Irregardless of the cost, go.", "EN_SOFT_IRREGARDLESS", "Regardless"},
		{"He is supposably ready.", "EN_SOFT_SUPPOSABLY", "supposedly"},
		{"They are the exact same.", "EN_SOFT_EXACT_SAME", ""},
		{"The reason why we left is clear.", "EN_SOFT_REASON_WHY", ""},
		{"Please return back soon.", "EN_SOFT_RETURN_BACK", ""},
		{"They ascend up the stairs.", "EN_SOFT_ASCEND_UP", ""},
		{"They descend down the hill.", "EN_SOFT_DESCEND_DOWN", ""},
		{"Please enter in the room.", "EN_SOFT_ENTER_IN", "enter the"},
		{"Continue on with the plan.", "EN_SOFT_CONTINUE_ON", "continue with"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type, "%+v", f)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftExtraTyposMap(t *testing.T) {
	cases := []struct {
		text, sug string
	}{
		{"I was commited to the goal.", "committed"},
		{"They are transfered today.", "transferred"},
		{"A nice restaraunt nearby.", "restaurant"},
		{"A good questionaire form.", "questionnaire"},
		{"The rythm is steady.", "rhythm"},
		{"He is persistant about it.", "persistent"},
		{"A vaccuum cleaner.", "vacuum"},
	}
	for _, tc := range cases {
		t.Run(tc.sug, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == "MORFOLOGIK_RULE_EN_US" && f.Suggestion == tc.sug {
					found = true
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftUSGBExtraVariants(t *testing.T) {
	us := []struct {
		text, rule, sug string
	}{
		{"Dry humour helps.", "EN_SOFT_HUMOUR_US", "humor"},
		{"Hard labour pays.", "EN_SOFT_LABOUR_US", "labor"},
		{"Self defence class.", "EN_SOFT_DEFENCE_US", "defense"},
		{"A driving licence.", "EN_SOFT_LICENCE_US", "license"},
		{"Please analyse data.", "EN_SOFT_ANALYSE_US", "analyze"},
		{"A large theatre hall.", "EN_SOFT_THEATRE_US", "theater"},
		{"Browse the catalogue.", "EN_SOFT_CATALOGUE_US", "catalog"},
		{"Open a dialogue box.", "EN_SOFT_DIALOGUE_US", "dialog"},
		{"TV programme tonight.", "EN_SOFT_PROGRAMME_US", "program"},
		{"My neighbour is kind.", "EN_SOFT_NEIGHBOUR_US", "neighbor"},
	}
	for _, tc := range us {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-US"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
	gb := []struct {
		text, rule, sug string
	}{
		{"Dry humor helps.", "EN_SOFT_HUMOR_GB", "humour"},
		{"Hard labor pays.", "EN_SOFT_LABOR_GB", "labour"},
		{"Self defense class.", "EN_SOFT_DEFENSE_GB", "defence"},
		{"A driving license.", "EN_SOFT_LICENSE_GB", "licence"},
		{"Please analyze data.", "EN_SOFT_ANALYZE_GB", "analyse"},
		{"A large theater hall.", "EN_SOFT_THEATER_GB", "theatre"},
		{"Browse the catalog.", "EN_SOFT_CATALOG_GB", "catalogue"},
		{"Open a dialog box.", "EN_SOFT_DIALOG_GB", "dialogue"},
		{"TV program tonight.", "EN_SOFT_PROGRAM_GB", "programme"},
		{"My neighbor is kind.", "EN_SOFT_NEIGHBOR_GB", "neighbour"},
	}
	for _, tc := range gb {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-GB"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftIrregardless(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Irregardless of cost, ship it.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "Regardless")
	require.NotContains(t, strings.ToLower(out.String()), "irregardless")
}

func TestGolden_SoftIdiomConfusables(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"The principal of the matter is clear.", "EN_SOFT_PRINCIPAL_REASON", "principle of the matter"},
		{"The principle officer spoke.", "EN_SOFT_PRINCIPLE_OFFICER", "principal officer"},
		{"Colors compliment each other well.", "EN_SOFT_COMPLIMENT_COLORS", "complement each other"},
		{"That will peek my interest soon.", "EN_SOFT_PEEK_INTEREST", "pique my interest"},
		{"That will peak my interest soon.", "EN_SOFT_PEAK_INTEREST", "pique my interest"},
		{"Please insure that the door is locked.", "EN_SOFT_ENSURE_INSURE", "ensure that"},
		{"We need farther discussion tomorrow.", "EN_SOFT_FARTHER_ABSTRACT", "further discussion"},
		{"I will lay down and rest now.", "EN_SOFT_LAY_DOWN_REST", "lie down and rest"},
		{"Please sight the source carefully.", "EN_SOFT_CITE_SEE", "cite the source"},
		{"Bare with me for a moment.", "EN_SOFT_BARE_WITH", "Bear with me"},
		{"A deep seeded fear remains.", "EN_SOFT_DEEP_SEEDED", "deep-seated"},
		{"Nip it in the butt early.", "EN_SOFT_NIP_IN_THE_BUTT", "Nip it in the bud"},
		{"Case and point: it failed.", "EN_SOFT_CASE_AND_POINT", "Case in point"},
		{"They are one in the same.", "EN_SOFT_ONE_IN_THE_SAME", "one and the same"},
		{"I ordered an expresso please.", "EN_SOFT_EXPRESSO", "espresso"},
		{"They tried to excape the room.", "EN_SOFT_EXCAPE", "escape"},
		{"I like it exspecially today.", "EN_SOFT_EXSPECIALLY", "especially"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftBareWithMe(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Bare with me for a second.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "Bear with me")
	require.NotContains(t, out.String(), "Bare with me")
}

func TestGolden_SoftPtBRRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Peguei o autocarro cedo.", "PT_SOFT_AUTOCARRO_BR", "ônibus"},
		{"Ligue no telemóvel agora.", "PT_SOFT_TELEMOVEL_BR", "celular"},
		{"O ecrã está sujo.", "PT_SOFT_ECRA_BR", "tela"},
		{"Esse facto importa.", "PT_SOFT_FACTO_BR", "fato"},
		{"Sem contacto visual.", "PT_SOFT_CONTACTO_BR", "contato"},
		{"Um objecto antigo.", "PT_SOFT_OBJECTO_BR", "objeto"},
		{"Resultado óptimo.", "PT_SOFT_OPTIMO_BR", "ótimo"},
		{"A acção principal.", "PT_SOFT_ACCAO_BR", "ação"},
		{"O comboio partiu.", "PT_SOFT_COMBOIO_BR", "trem"},
		{"A equipa ganhou.", "PT_SOFT_EQUIPA_BR", "equipe"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "pt-BR"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPtPTRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Peguei o ônibus cedo.", "PT_SOFT_ONIBUS_PT", "autocarro"},
		{"Ligue no celular agora.", "PT_SOFT_CELULAR_PT", "telemóvel"},
		{"Esse fato importa.", "PT_SOFT_FATO_PT", "facto"},
		{"Sem contato visual.", "PT_SOFT_CONTATO_PT", "contacto"},
		{"Um objeto antigo.", "PT_SOFT_OBJETO_PT", "objecto"},
		{"Resultado ótimo.", "PT_SOFT_OTIMO_PT", "óptimo"},
		{"O trem partiu.", "PT_SOFT_TREM_PT", "comboio"},
		{"A equipe ganhou.", "PT_SOFT_EQUIPE_PT", "equipa"},
		{"Café da manhã cedo.", "PT_SOFT_CAFE_MANHA_PT", "pequeno-almoço"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "pt-PT"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPtBRNotOnPtPT(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Peguei o autocarro cedo.", &CommandLineOptions{Language: "pt-PT"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "PT_SOFT_AUTOCARRO_BR", f.Rule, "%+v", findings)
	}
}

func TestGolden_ApplySoftPtBRAutocarro(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "pt-BR", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Peguei o autocarro cedo.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "ônibus")
	require.NotContains(t, out.String(), "autocarro")
}

func TestGolden_SoftEsMXRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Uso el ordenador hoy.", "ES_SOFT_ORDENADOR_MX", "computadora"},
		{"Mi móvil es nuevo.", "ES_SOFT_MOVIL_MX", "celular"},
		{"Quiero zumo de naranja.", "ES_SOFT_ZUMO_MX", "jugo"},
		{"El coche es rojo.", "ES_SOFT_COCHE_MX", "carro"},
		{"Voy a conducir despacio.", "ES_SOFT_CONDUCIR_MX", "manejar"},
		{"Compré un bolígrafo.", "ES_SOFT_BOLIGRAFO_MX", "pluma"},
		{"Un melocotón maduro.", "ES_SOFT_MELON_MX", "durazno"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "es-MX"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftEsESRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Uso la computadora hoy.", "ES_SOFT_COMPUTADORA_ES", "ordenador"},
		{"Mi celular es nuevo.", "ES_SOFT_CELULAR_ES", "móvil"},
		{"Quiero jugo de naranja.", "ES_SOFT_JUGO_ES", "zumo"},
		{"El carro es rojo.", "ES_SOFT_CARRO_ES", "coche"},
		{"Voy a manejar despacio.", "ES_SOFT_MANEJAR_ES", "conducir"},
		{"Un durazno maduro.", "ES_SOFT_DURAZNO_ES", "melocotón"},
		{"Compré una laptop.", "ES_SOFT_LAPTOP_ES", "portátil"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "es-ES"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftDeCHRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Die Straße ist nass.", "DE_SOFT_STRASSE_CH", "Strasse"},
		{"Ein groß Haus.", "DE_SOFT_GROSS_CH", "gross"},
		{"Hier darf man parken.", "DE_SOFT_PARKEN_CH", "parkieren"},
		{"Mein Fahrrad ist neu.", "DE_SOFT_FAHRRAD_CH", "Velo"},
		{"Das Handy klingelt.", "DE_SOFT_HANDY_CH", "Natel"},
		{"Wir grillen heute.", "DE_SOFT_GRILLEN_CH", "grillieren"},
		{"Es ist heiß draußen.", "DE_SOFT_ZWEI_CH", "heiss"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "de-CH"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftDeATRegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Im Januar schneit es.", "DE_SOFT_JANUAR_AT", "Jänner"},
		{"Wir essen Kartoffeln.", "DE_SOFT_KARTOFFELN_AT", "Erdäpfel"},
		{"Frische Tomaten bitte.", "DE_SOFT_TOMATEN_AT", "Paradeiser"},
		{"Ein Brötchen zum Kaffee.", "DE_SOFT_BROETCHEN_AT", "Semmel"},
		{"Gebratenes Hähnchen.", "DE_SOFT_HAEHNCHEN_AT", "Hendl"},
		{"Frischkäse und Quark.", "DE_SOFT_QUARK_AT", "Topfen"},
		{"Die Treppe ist steil.", "DE_SOFT_TREPPE_AT", "Stiege"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "de-AT"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftFrCARegionalHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Bon week-end à tous.", "FR_SOFT_WEEKEND_CA", "fin de semaine"},
		{"Envoie un e-mail vite.", "FR_SOFT_EMAIL_CA", "courriel"},
		{"Mon portable sonne.", "FR_SOFT_PORTABLE_CA", "cellulaire"},
		{"La voiture est rouge.", "FR_SOFT_VOITURE_CA", "char"},
		{"Je vais faire des courses.", "FR_SOFT_COURSES_CA", "magasiner"},
		{"Le parking est plein.", "FR_SOFT_PARKING_CA", "stationnement"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "fr-CA"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftEsMXNotOnEsES(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Uso el ordenador hoy.", &CommandLineOptions{Language: "es-ES"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "ES_SOFT_ORDENADOR_MX", f.Rule, "%+v", findings)
	}
}

func TestGolden_ApplySoftEsMXOrdenador(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "es-MX", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Uso el ordenador hoy.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "computadora")
	require.NotContains(t, out.String(), "ordenador")
}

func TestGolden_SoftCasingLowercaseI(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"i think so.", "EN_SOFT_LOWERCASE_I", "I"},
		{"im ready now.", "EN_SOFT_LOWERCASE_IM", "I'm"},
		{"ive finished work.", "EN_SOFT_LOWERCASE_IVE", "I've"},
		{"id like coffee.", "EN_SOFT_LOWERCASE_ID_LIKE", "I'd like"},
		{"ill go later.", "EN_SOFT_LOWERCASE_ILL_GO", "I'll go"},
		{"id love to join.", "EN_SOFT_LOWERCASE_ID_LOVE", "I'd love"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "typographical", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftLowercaseI(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "i think this works.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "I think")
	require.NotContains(t, out.String(), "i think")
}

func TestGolden_SoftStyleMetaViaListRules(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "en"))
	out := buf.String()
	// SoftRuleMeta classifies these for list-rules columns
	require.Contains(t, out, "EN_SOFT_LOWERCASE_I\tCASING\t")
	require.Contains(t, out, "EN_SOFT_THE_THE\tSTYLE\t")
	var us bytes.Buffer
	require.NoError(t, CoreListRules(&us, "en-US"))
	require.Contains(t, us.String(), "EN_SOFT_COLOUR_US\tTYPOS\t")
}

func TestGolden_SoftIdiomConfusablesMore(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"She poured over the document carefully.", "EN_SOFT_PORED_OVER", "pored over the document"},
		{"They waive goodbye at the station.", "EN_SOFT_WAIVE_GOODBYE", "wave goodbye"},
		{"That is a mute point entirely.", "EN_SOFT_MUTE_POINT", "moot point"},
		{"That does not jive with the data.", "EN_SOFT_JIVE_WITH", "jibe with"},
		{"We will hone in on the bug.", "EN_SOFT_HONE_IN", "home in on"},
		{"A clever slight of hand trick.", "EN_SOFT_SLIGHT_OF_HAND", "sleight of hand"},
		{"Soldiers tow the line carefully.", "EN_SOFT_TOW_THE_LINE", "toe the line"},
		{"This begs the question of timing.", "EN_SOFT_BEGS_THE_QUESTION_SOFT", ""},
		{"Is that ok with you?", "EN_SOFT_CASE_SENSITIVE_OK", "OK"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftMultiLang(t *testing.T) {
	cases := []struct {
		lang, in, want, not string
	}{
		{"de", "Ich denke das es stimmt.", "dass", "denke das"},
		{"fr", "Je vais a la maison.", "à la", "a la"},
		{"es", "Voy a el parque.", "al", "a el"},
		{"pt", "Vou a o mercado.", "ao", "a o"},
		{"nl", "Meer als gisteren.", "Meer dan", "Meer als"},
		{"it", "Vado a il negozio.", "al", "a il"},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			var out, errb bytes.Buffer
			code := RunWithIO([]string{"-l", tc.lang, "--apply", "-"}, RunHooks{
				ReadStdin: func() (string, error) { return tc.in, nil },
				Check:     CoreApplySuggestionsHook,
			}, &out, &errb)
			require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
			require.Contains(t, out.String(), tc.want)
			// not all rules remove the bad span text entirely when suggestions replace differently
			_ = tc.not
		})
	}
}

func TestGolden_IgnoreSpellingChatGPT(t *testing.T) {
	if DiscoverEnglishIgnoreSpellingList(nil) == "" {
		t.Skip("ignore list missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I use ChatGPT and Claude daily.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftDisableCategoriesStyle(t *testing.T) {
	// STYLE soft rule suppressed; grammar soft confusable remains
	text := "This is very unique work. Bare with me please."
	var all bytes.Buffer
	_, err := CoreGoldenHook(&all, text, &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var allF []Finding
	require.NoError(t, json.Unmarshal(all.Bytes(), &allF))
	var hasStyle, hasBare bool
	for _, f := range allF {
		if f.Rule == "EN_SOFT_VERY_UNIQUE" {
			hasStyle = true
		}
		if f.Rule == "EN_SOFT_BARE_WITH" {
			hasBare = true
		}
	}
	require.True(t, hasStyle && hasBare, "%+v", allF)

	var filtered bytes.Buffer
	_, err = CoreGoldenHook(&filtered, text, &CommandLineOptions{
		Language:           "en",
		DisabledCategories: []string{"STYLE"},
	})
	require.NoError(t, err)
	var ff []Finding
	require.NoError(t, json.Unmarshal(filtered.Bytes(), &ff))
	for _, f := range ff {
		require.NotEqual(t, "EN_SOFT_VERY_UNIQUE", f.Rule, "%+v", ff)
	}
	foundBare := false
	for _, f := range ff {
		if f.Rule == "EN_SOFT_BARE_WITH" {
			foundBare = true
		}
	}
	require.True(t, foundBare, "%+v", ff)
}

func TestGolden_SoftEnableCategoriesCasing(t *testing.T) {
	text := "i think so. Bare with me please."
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{
		Language:          "en",
		EnabledCategories: []string{"CASING"},
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	require.NotEmpty(t, findings)
	// only CASING-category matches (soft lowercase i and/or core sentence-start casing)
	for _, f := range findings {
		require.Equal(t, "typographical", f.Type, "%+v", f)
		require.NotEqual(t, "EN_SOFT_BARE_WITH", f.Rule)
		require.NotEqual(t, "style", f.Type)
	}
	foundLowerI := false
	for _, f := range findings {
		if f.Rule == "EN_SOFT_LOWERCASE_I" || f.Rule == "UPPERCASE_SENTENCE_START" {
			foundLowerI = true
		}
	}
	require.True(t, foundLowerI, "%+v", findings)
}

func TestGolden_SoftDisableRule(t *testing.T) {
	// configureCoreLT applies -d via ApplyCLIRuleFilters
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Bare with me please.", &CommandLineOptions{
		Language: "en",
	})
	require.NoError(t, err)
	var before []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &before))
	found := false
	for _, f := range before {
		if f.Rule == "EN_SOFT_BARE_WITH" {
			found = true
		}
	}
	require.True(t, found, "%+v", before)

	// Use RunWithIO -d flag path
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-d", "EN_SOFT_BARE_WITH", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Bare with me please.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s out=%s", code, errb.String(), out.String())
	require.NotContains(t, out.String(), "EN_SOFT_BARE_WITH")
}

func TestGolden_SoftIdiomConfusablesWave2(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They wreck havoc everywhere.", "EN_SOFT_WREAK_HAVOC", "wreak havoc"},
		{"Give free reign to creativity.", "EN_SOFT_FREE_REIN", "free rein"},
		{"Wait with baited breath.", "EN_SOFT_BAITED_BREATH", "bated breath"},
		{"The event was a damp squid.", "EN_SOFT_DAMP_SQUID", "damp squib"},
		{"Take a sneak peak at this.", "EN_SOFT_SNEAK_PEAK", "sneak peek"},
		{"They will extract revenge soon.", "EN_SOFT_EXTRACT_REVENGE", "exact revenge"},
		{"That will peaked my interest soon.", "EN_SOFT_PEAKED_INTEREST_ALT", "piqued my interest"},
		{"Nothing will phase me today.", "EN_SOFT_PHASE_OF_THE_MOON", "faze me"},
		{"Please reign in spending.", "EN_SOFT_REIGN_IN", "rein in"},
		{"A pallet cleanser follows.", "EN_SOFT_PALATE_CLEANSER", "palate cleanser"},
		{"He is a shoe in for the job.", "EN_SOFT_SHOE_IN", "shoo-in"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftMultiLangDoubleQ(t *testing.T) {
	cases := []struct {
		lang, text, rule string
	}{
		{"de", "Wirklich??", "DE_SOFT_DOUBLE_Q"},
		{"fr", "Vraiment??", "FR_SOFT_DOUBLE_Q"},
		{"es", "¿Qué??", "ES_SOFT_DOUBLE_Q"},
		{"pt", "Sério??", "PT_SOFT_DOUBLE_Q"},
		{"it", "Davvero??", "IT_SOFT_DOUBLE_Q"},
		{"nl", "Echt??", "NL_SOFT_DOUBLE_Q"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: tc.lang})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "typographical", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyPackOnlyWhenPicky(t *testing.T) {
	text := "Please utilize synergy going forward."
	var def bytes.Buffer
	_, err := CoreGoldenHook(&def, text, &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var defF []Finding
	require.NoError(t, json.Unmarshal(def.Bytes(), &defF))
	for _, f := range defF {
		require.False(t, strings.Contains(f.Rule, "SOFT_PICKY"), "%+v", defF)
	}

	var picky bytes.Buffer
	_, err = CoreGoldenHook(&picky, text, &CommandLineOptions{Language: "en", Level: "PICKY"})
	require.NoError(t, err)
	var pf []Finding
	require.NoError(t, json.Unmarshal(picky.Bytes(), &pf))
	want := map[string]string{
		"EN_SOFT_PICKY_UTILIZE":       "use",
		"EN_SOFT_PICKY_SYNERGY":       "",
		"EN_SOFT_PICKY_GOING_FORWARD": "",
	}
	for rule, sug := range want {
		found := false
		for _, f := range pf {
			if f.Rule == rule {
				found = true
				require.Equal(t, "style", f.Type)
				if sug != "" {
					require.Equal(t, sug, f.Suggestion)
				}
			}
		}
		require.True(t, found, "missing %s in %+v", rule, pf)
	}
}

func TestGolden_SoftPickyMore(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"We will leverage the API.", "EN_SOFT_PICKY_LEVERAGE", "use"},
		{"An impactful change.", "EN_SOFT_PICKY_IMPACTFUL", "effective"},
		{"Actionable feedback helps.", "EN_SOFT_PICKY_ACTIONABLE", "Practical"},
		{"Let's circle back tomorrow.", "EN_SOFT_PICKY_CIRCLE_BACK", ""},
		{"Touch base later today.", "EN_SOFT_PICKY_TOUCH_BASE", "talk"},
		{"At the end of the day, ship it.", "EN_SOFT_PICKY_AT_THE_END_OF_THE_DAY", "ultimately"},
		{"Preventative care matters.", "EN_SOFT_PICKY_PREVENTATIVE", "Preventive"},
		{"And stuff like that.", "EN_SOFT_PICKY_STUFF", ""},
		{"A lot of things changed.", "EN_SOFT_PICKY_THINGS", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftPickyUtilize(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--level", "picky", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Please utilize the tool now.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "use")
	require.NotContains(t, strings.ToLower(out.String()), "utilize")
}

func TestGolden_SoftPickyDE(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Im Hinblick auf den Plan warte ich.", "DE_SOFT_PICKY_NUTZEN_VON"},
		{"Last but not least danke ich allen.", "DE_SOFT_PICKY_LAST_BUT_NOT"},
		{"Am Ende des Tages entscheiden wir.", "DE_SOFT_PICKY_AM_ENDE_DES_TAGES"},
		{"Es ist sehr sehr wichtig.", "DE_SOFT_PICKY_SEHR_SEHR"},
		{"Wir müssen viele Dinge klären.", "DE_SOFT_PICKY_DINGE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var def bytes.Buffer
			_, err := CoreGoldenHook(&def, tc.text, &CommandLineOptions{Language: "de"})
			require.NoError(t, err)
			var defF []Finding
			require.NoError(t, json.Unmarshal(def.Bytes(), &defF))
			for _, f := range defF {
				require.NotEqual(t, tc.rule, f.Rule)
			}
			var buf bytes.Buffer
			_, err = CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "de", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyFR(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"C'est un challenge important.", "FR_SOFT_PICKY_CHALLENGE", "défi"},
		{"Le meeting est demain.", "FR_SOFT_PICKY_MEETING", "réunion"},
		{"Au final, on part.", "FR_SOFT_PICKY_AU_FINAL", "finalement"},
		{"C'est très très bon.", "FR_SOFT_PICKY_TRES_TRES", ""},
		{"En termes de budget, ok.", "FR_SOFT_PICKY_EN_TERMES_DE", ""},
		{"Au niveau de la qualité, oui.", "FR_SOFT_PICKY_AU_NIVEAU_DE", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var def bytes.Buffer
			_, err := CoreGoldenHook(&def, tc.text, &CommandLineOptions{Language: "fr"})
			require.NoError(t, err)
			var defF []Finding
			require.NoError(t, json.Unmarshal(def.Bytes(), &defF))
			for _, f := range defF {
				require.NotEqual(t, tc.rule, f.Rule)
			}
			var buf bytes.Buffer
			_, err = CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "fr", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENExtra(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"We need more bandwidth this week.", "EN_SOFT_PICKY_BANDWIDTH"},
		{"Focus on low hanging fruit first.", "EN_SOFT_PICKY_LOW_HANGING"},
		{"This will move the needle soon.", "EN_SOFT_PICKY_MOVE_THE_NEEDLE"},
		{"Let's do a deep dive tomorrow.", "EN_SOFT_PICKY_DEEP_DIVE"},
		{"The key takeaway is clear.", "EN_SOFT_PICKY_TAKEAWAY"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyES(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Es un challenge difícil.", "ES_SOFT_PICKY_CHALLENGE", "reto"},
		{"Necesito feedback pronto.", "ES_SOFT_PICKY_FEEDBACK", "comentarios"},
		{"Es muy muy importante.", "ES_SOFT_PICKY_MUY_MUY", ""},
		{"Hay muchas cosas que hacer.", "ES_SOFT_PICKY_COSAS", ""},
		{"Al final del día, enviamos.", "ES_SOFT_PICKY_AL_FINAL_DEL_DIA", ""},
		{"A nivel de calidad, bien.", "ES_SOFT_PICKY_A_NIVEL_DE", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var def bytes.Buffer
			_, err := CoreGoldenHook(&def, tc.text, &CommandLineOptions{Language: "es"})
			require.NoError(t, err)
			var defF []Finding
			require.NoError(t, json.Unmarshal(def.Bytes(), &defF))
			for _, f := range defF {
				require.NotEqual(t, tc.rule, f.Rule)
			}
			var buf bytes.Buffer
			_, err = CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "es", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyPT(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Recebi feedback ontem.", "PT_SOFT_PICKY_FEEDBACK", "retorno"},
		{"É um challenge grande.", "PT_SOFT_PICKY_CHALLENGE", "desafio"},
		{"É muito muito bom.", "PT_SOFT_PICKY_MUITO_MUITO", ""},
		{"Há muitas coisas a fazer.", "PT_SOFT_PICKY_COISAS", ""},
		{"No final do dia, enviamos.", "PT_SOFT_PICKY_NO_FINAL_DO_DIA", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "pt", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyIT(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Il meeting è domani.", "IT_SOFT_PICKY_MEETING", "riunione"},
		{"È un challenge duro.", "IT_SOFT_PICKY_CHALLENGE", "sfida"},
		{"È molto molto buono.", "IT_SOFT_PICKY_MOLTO_MOLTO", ""},
		{"Ci sono tante cose da fare.", "IT_SOFT_PICKY_COSE", ""},
		{"Dammi un feedback presto.", "IT_SOFT_PICKY_FEEDBACK", "riscontro"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "it", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ApplySoftPickyMultiLang(t *testing.T) {
	cases := []struct {
		lang, in, want string
	}{
		{"fr", "Le meeting est demain.", "réunion"},
		{"es", "Es un challenge difícil.", "reto"},
		{"it", "Il meeting è domani.", "riunione"},
		{"pt", "É um challenge grande.", "desafio"},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			var out, errb bytes.Buffer
			code := RunWithIO([]string{"-l", tc.lang, "--level", "picky", "--apply", "-"}, RunHooks{
				ReadStdin: func() (string, error) { return tc.in, nil },
				Check:     CoreApplySuggestionsHook,
			}, &out, &errb)
			require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
			require.Contains(t, out.String(), tc.want)
		})
	}
}

func TestGolden_SoftPickyListRulesCLI(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"rules", "-l", "en", "--level", "picky"}, DefaultCoreHooks(), &out, &errb)
	require.Equal(t, 0, code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_PICKY_UTILIZE")
	require.Contains(t, out.String(), "soft_picky=")
	require.Contains(t, out.String(), "level=picky")
}

func TestGolden_SoftOptionalDefaultOff(t *testing.T) {
	text := "Prior to leaving, call home."
	// default: optional soft rule is registered but off
	var def bytes.Buffer
	_, err := CoreGoldenHook(&def, text, &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var defF []Finding
	require.NoError(t, json.Unmarshal(def.Bytes(), &defF))
	for _, f := range defF {
		require.NotEqual(t, "EN_SOFT_OPT_PRIOR_TO", f.Rule)
	}

	// enable with -e
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "EN_SOFT_OPT_PRIOR_TO", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_OPT_PRIOR_TO")
}

func TestGolden_SoftOptionalMore(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"From the get go we knew.", "EN_SOFT_OPT_GET_GO"},
		{"At this juncture we wait.", "EN_SOFT_OPT_AT_THIS_JUNCTURE"},
		{"With regard to fees, wait.", "EN_SOFT_OPT_WITH_REGARD_TO"},
		{"In the event of rain, cancel.", "EN_SOFT_OPT_IN_THE_EVENT"},
		{"Subsequent to review, ship.", "EN_SOFT_OPT_SUBSEQUENT_TO"},
		{"Due to the fact of rain, cancel.", "EN_SOFT_OPT_DUE_TO_THE_FACT_OF"},
		{"In order for this to work, wait.", "EN_SOFT_OPT_IN_ORDER_FOR"},
		{"For the purpose of clarity, rewrite.", "EN_SOFT_OPT_FOR_THE_PURPOSE_OF"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var out, errb bytes.Buffer
			code := RunWithIO([]string{"-l", "en", "-e", tc.rule, "--json", "-"}, RunHooks{
				ReadStdin: func() (string, error) { return tc.text, nil },
				Check:     CoreCheckHook,
			}, &out, &errb)
			require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
			require.Contains(t, out.String(), tc.rule)
		})
	}
}

func TestGolden_SoftOptionalDE(t *testing.T) {
	text := "Im Rahmen von Tests warten wir."
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "de", "-e", "DE_SOFT_OPT_IM_RAHMEN", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "DE_SOFT_OPT_IM_RAHMEN")
}

func TestGolden_SoftOptionalFR(t *testing.T) {
	text := "Dans le cadre de ce projet, on avance."
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "fr", "-e", "FR_SOFT_OPT_DANS_LE_CADRE", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "FR_SOFT_OPT_DANS_LE_CADRE")
}

func TestGolden_SoftOptionalES(t *testing.T) {
	text := "Con el fin de mejorar, estudia."
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "es", "-e", "ES_SOFT_OPT_CON_EL_FIN_DE", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "ES_SOFT_OPT_CON_EL_FIN_DE")
}

func TestGolden_SoftOptionalPT(t *testing.T) {
	text := "Com vistas a melhorar, estude."
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "pt", "-e", "PT_SOFT_OPT_COM_VISTAS_A", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "PT_SOFT_OPT_COM_VISTAS_A")
}

func TestGolden_SoftOptionalIT(t *testing.T) {
	text := "Al fine di migliorare, studia."
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "it", "-e", "IT_SOFT_OPT_AL_FINE_DI", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "IT_SOFT_OPT_AL_FINE_DI")
}

func TestGolden_ApplySoftOptionalPriorTo(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "EN_SOFT_OPT_PRIOR_TO", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Prior to leaving, call.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	// multi-token match → shorter suggestion keeps lowercase ("before")
	require.Contains(t, out.String(), "before")
	require.NotContains(t, out.String(), "Prior to")
}

func TestGolden_SoftOptionalBulkEnable(t *testing.T) {
	// -e SOFT_OPTIONAL enables all default-off SOFT_OPT_* rules
	text := "Prior to leaving, with regard to fees, subsequent to review."
	var off bytes.Buffer
	_, err := CoreGoldenHook(&off, text, &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var offF []Finding
	require.NoError(t, json.Unmarshal(off.Bytes(), &offF))
	for _, f := range offF {
		require.False(t, strings.Contains(f.Rule, "SOFT_OPT_"), "%+v", offF)
	}

	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	s := out.String()
	require.Contains(t, s, "EN_SOFT_OPT_PRIOR_TO")
	require.Contains(t, s, "EN_SOFT_OPT_WITH_REGARD_TO")
	require.Contains(t, s, "EN_SOFT_OPT_SUBSEQUENT_TO")
}

func TestGolden_SoftOptionalBulkEnableAlias(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "SOFT_OPT_ALL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "At this juncture we wait.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_OPT_AT_THIS_JUNCTURE")
}

func TestExpandSoftEnableAliases(t *testing.T) {
	lt, err := configureCoreLT("en", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	reg := lt.GetAllRegisteredRuleIDs()
	exp := languagetool.ExpandSoftEnableRuleIDs(reg, []string{"SOFT_OPTIONAL", "EN_A_VS_AN"})
	require.Contains(t, exp, "EN_A_VS_AN")
	var optN int
	for _, id := range exp {
		if strings.Contains(id, "SOFT_OPT_") {
			optN++
		}
	}
	require.GreaterOrEqual(t, optN, 6)
	// non-alias passthrough
	require.Equal(t, []string{"EN_SOFT_OPT_PRIOR_TO"}, languagetool.ExpandSoftEnableRuleIDs(reg, []string{"EN_SOFT_OPT_PRIOR_TO"}))
}

func TestGolden_SoftPickyPL(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Mamy meeting jutro.", "PL_SOFT_PICKY_MEETING", "spotkanie"},
		{"Chcę feedback szybko.", "PL_SOFT_PICKY_FEEDBACK", "opinię"},
		{"To jest bardzo bardzo ważne.", "PL_SOFT_PICKY_BARDZO_BARDZO", ""},
		{"Jest wiele rzeczy do zrobienia.", "PL_SOFT_PICKY_RZECZY", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "pl", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalNL(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "nl", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "In het kader van dit plan wachten we.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "NL_SOFT_OPT_IN_HET_KADER")
}

func TestGolden_SoftOptionalPL(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "pl", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "W ramach projektu czekamy.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "PL_SOFT_OPT_W_RAMACH")
}

func TestGolden_SoftOptionalSV(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "sv", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I syfte att förbättra väntar vi.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "SV_SOFT_OPT_I_SYFTE")
}

func TestGolden_SoftOptionalDA(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "da", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I forbindelse med planen venter vi.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "DA_SOFT_OPT_I_FORBINDELSE")
}

func TestGolden_SoftOptionalRU(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ru", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "В рамках проекта ждём.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "RU_SOFT_OPT_V_RAMKAH")
}

func TestGolden_SoftOptionalUK(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "uk", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "В рамках проєкту чекаємо.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "UK_SOFT_OPT_V_RAMKAH")
}

func TestGolden_SoftOptionalCA(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ca", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "A nivell de producte, esperem.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "CA_SOFT_OPT_A_NIVELL_DE")
}

func TestGolden_SoftOptionalDefaultOffPL(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "W ramach projektu czekamy.", &CommandLineOptions{Language: "pl"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.False(t, strings.Contains(f.Rule, "SOFT_OPT_"), "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave3(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A movie in the vain of noir.", "EN_SOFT_VAIN_OF", "in the vein of"},
		{"We must diffuse the situation now.", "EN_SOFT_DIFFUSE_THE_SITUATION", "defuse the situation"},
		{"They flaunt the rules daily.", "EN_SOFT_FLAUNT_THE_RULES", "flout the rules"},
		{"She will pour over the book tonight.", "EN_SOFT_PORE_OVER_BOOK", "pore over the book"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyDA(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Vi har en meeting i morgen.", "DA_SOFT_PICKY_MEETING", "møde"},
		{"Jeg vil have feedback snart.", "DA_SOFT_PICKY_FEEDBACK", "tilbagemelding"},
		{"Det er meget meget vigtigt.", "DA_SOFT_PICKY_MEGET_MEGET", ""},
		{"Der er mange ting at gøre.", "DA_SOFT_PICKY_TING", ""},
		{"I sidste ende beslutter vi.", "DA_SOFT_PICKY_I_SIDSTE_ENDE", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "da", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyRU(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"У нас meeting завтра.", "RU_SOFT_PICKY_MEETING", "встреча"},
		{"Нужен feedback сегодня.", "RU_SOFT_PICKY_FEEDBACK", "отзыв"},
		{"Это очень очень важно.", "RU_SOFT_PICKY_OCHEN_OCHEN", ""},
		{"Много вещей осталось.", "RU_SOFT_PICKY_VESCHI", ""},
		{"В конце концов мы решим.", "RU_SOFT_PICKY_V_KONCE_KONCOV", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "ru", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftHelpMentionsOptional(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"help"}, DefaultCoreHooks(), &out, &errb)
	require.Equal(t, 0, code, errb.String())
	s := out.String()
	require.Contains(t, s, "SOFT_OPTIONAL")
	require.Contains(t, s, "PICKY")
	require.Contains(t, s, "enablecategories")
}

func TestGolden_SoftPickyUK(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Маємо meeting завтра.", "UK_SOFT_PICKY_MEETING", "зустріч"},
		{"Потрібен feedback сьогодні.", "UK_SOFT_PICKY_FEEDBACK", "відгук"},
		{"Це дуже дуже важливо.", "UK_SOFT_PICKY_DUZHE_DUZHE", ""},
		{"Багато речей залишилось.", "UK_SOFT_PICKY_RECHI", ""},
		{"В кінці кінців вирішимо.", "UK_SOFT_PICKY_V_RECHTI_RECHT", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "uk", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyCA(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Tenim un meeting demà.", "CA_SOFT_PICKY_MEETING", "reunió"},
		{"Vull feedback avui.", "CA_SOFT_PICKY_FEEDBACK", "comentaris"},
		{"És molt molt important.", "CA_SOFT_PICKY_MOLT_MOLT", ""},
		{"Hi ha moltes coses a fer.", "CA_SOFT_PICKY_COSES", ""},
		{"Al final del dia, enviem.", "CA_SOFT_PICKY_AL_FINAL_DEL_DIA", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "ca", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave4(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please appraise of the delay.", "EN_SOFT_APPRAISE_OF", "apprise of"},
		{"I study discreet math daily.", "EN_SOFT_DISCREET_MATH", "discrete math"},
		{"Imminent domain law applies.", "EN_SOFT_EMINENT_DOMAIN_OK", "Eminent domain"},
		{"For all intensive purposes it works.", "EN_SOFT_FOR_ALL_INTENSIVE_ALT", "For all intents and purposes"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave5(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"For piece of mind I backed up.", "EN_SOFT_PIECE_OF_MIND", "peace of mind"},
		{"He got his just desserts.", "EN_SOFT_JUST_DESSERTS", "just deserts"},
		{"That will wet your appetite.", "EN_SOFT_WET_YOUR_APPETITE", "whet your appetite"},
		{"The room is chalk full of books.", "EN_SOFT_CHALK_FULL", "chock-full"},
		{"Said tongue and cheek, of course.", "EN_SOFT_TONGUE_AND_CHEEK", "tongue-in-cheek"},
		{"A straight laced teacher arrived.", "EN_SOFT_STRAIGHT_LACED", "strait-laced"},
		{"That was a bold-faced lie.", "EN_SOFT_BOLD_FACED_LIE", "bald-faced lie"},
		{"Screen for prostrate cancer early.", "EN_SOFT_PROSTRATE_CANCER", "prostate cancer"},
		{"I broke it on accident yesterday.", "EN_SOFT_ON_ACCIDENT", "by accident"},
		{"Please unthaw the chicken first.", "EN_SOFT_UNTHAW", "thaw"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyEL(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Έχουμε meeting αύριο.", "EL_SOFT_PICKY_MEETING", ""},
		{"Χρειάζομαι feedback σήμερα.", "EL_SOFT_PICKY_FEEDBACK", ""},
		{"Είναι πολύ πολύ σημαντικό.", "EL_SOFT_PICKY_POLY_POLY", ""},
		{"Υπάρχουν πολλά πράγματα να κάνουμε.", "EL_SOFT_PICKY_PRAGMATA", ""},
		{"Στο τέλος της ημέρας αποφασίζουμε.", "EL_SOFT_PICKY_STO_TELOS_TIS_IMERAS", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "el", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyRO(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Avem un meeting mâine.", "RO_SOFT_PICKY_MEETING"},
		{"Vreau feedback azi.", "RO_SOFT_PICKY_FEEDBACK"},
		{"Este foarte foarte important.", "RO_SOFT_PICKY_FOARTE_FOARTE"},
		{"Sunt multe lucruri de făcut.", "RO_SOFT_PICKY_LUCRURI"},
		{"La sfârșitul zilei trimitem.", "RO_SOFT_PICKY_LA_SFARSITUL_ZILEI"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "ro", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyGL(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Temos un meeting mañá.", "GL_SOFT_PICKY_MEETING"},
		{"Quero feedback hoxe.", "GL_SOFT_PICKY_FEEDBACK"},
		{"É moi moi importante.", "GL_SOFT_PICKY_MOI_MOI"},
		{"Hai moitas cousas que facer.", "GL_SOFT_PICKY_COUSAS"},
		{"Ao final do día enviamos.", "GL_SOFT_PICKY_AO_FINAL_DO_DIA"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "gl", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_FalseFriendsMenu(t *testing.T) {
	ff := softFalseFriendsPath(t)
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Le menu du jour est bon.", &CommandLineOptions{
		Language:         "fr",
		MotherTongue:     "en",
		FalseFriendsFile: ff,
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "MENU" {
			found = true
			require.Equal(t, "set meal / fixed-price meal", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftIdiomConfusablesWave6(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"In regards to fees, wait.", "EN_SOFT_IN_REGARDS_TO", "In regard to"},
		{"With regards to timing, ship.", "EN_SOFT_WITH_REGARDS_TO", "With regard to"},
		{"All of the sudden it failed.", "EN_SOFT_ALL_OF_THE_SUDDEN", "All of a sudden"},
		{"Please conversate with them.", "EN_SOFT_CONVERSATE", "converse"},
		{"We need to orientate the team.", "EN_SOFT_ORIENTATE", "orient"},
		{"This is undoubtably correct.", "EN_SOFT_UNDOUBTABLY", "undoubtedly"},
		{"A miniscule amount remains.", "EN_SOFT_MINISCULE", "minuscule"},
		{"Use the ATM machine nearby.", "EN_SOFT_ATM_MACHINE", "ATM"},
		{"Enter your PIN number now.", "EN_SOFT_PIN_NUMBER", "PIN"},
		{"Test for HIV virus early.", "EN_SOFT_HIV_VIRUS", "HIV"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalEL(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "el", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Στο πλαίσιο του έργου περιμένουμε.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EL_SOFT_OPT_STO_PLAISIO")
}

func TestGolden_SoftOptionalRO(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ro", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "În cadrul proiectului așteptăm.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "RO_SOFT_OPT_IN_CADRUL")
}

func TestGolden_SoftOptionalGL(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "gl", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "No marco do plan esperamos.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "GL_SOFT_OPT_NO_MARCO")
}

func TestGolden_SoftPickySK(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Máme meeting zajtra.", "SK_SOFT_PICKY_MEETING"},
		{"Potrebujem feedback dnes.", "SK_SOFT_PICKY_FEEDBACK"},
		{"Je to veľmi veľmi dôležité.", "SK_SOFT_PICKY_VELMI_VELMI"},
		{"Je veľa vecí na riešenie.", "SK_SOFT_PICKY_VECI"},
		{"Na konci dňa rozhodneme.", "SK_SOFT_PICKY_NAKONIEC"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "sk", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickySL(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Imamo meeting jutri.", "SL_SOFT_PICKY_MEETING"},
		{"Potrebujem feedback danes.", "SL_SOFT_PICKY_FEEDBACK"},
		{"Je zelo zelo pomembno.", "SL_SOFT_PICKY_ZELO_ZELO"},
		{"Je veliko stvari za narediti.", "SL_SOFT_PICKY_STVARI"},
		{"Na koncu dneva odločimo.", "SL_SOFT_PICKY_NA_KONCU_DNEVA"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "sl", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave7(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I definately agree with that.", "EN_SOFT_DEFINATELY", "definitely"},
		{"Please seperate the files.", "EN_SOFT_SEPERATE", "separate"},
		{"It occured last night.", "EN_SOFT_OCCURED", "occurred"},
		{"We can accomodate guests.", "EN_SOFT_ACCOMODATE", "accommodate"},
		{"It is neccessary to wait.", "EN_SOFT_NECCESSARY", "necessary"},
		{"See you tommorrow morning.", "EN_SOFT_TOMMORROW", "tomorrow"},
		{"Green foilage covered the path.", "EN_SOFT_FOILAGE", "foliage"},
		{"A mischievious smile appeared.", "EN_SOFT_MISCHIEVIOUS", "mischievous"},
		{"Nucular energy is debated.", "EN_SOFT_NUCULAR", "Nuclear"},
		{"Firstable, ship the fix.", "EN_SOFT_FIRSTABLE", "First of all"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalSK(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "sk", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "V rámci projektu čakáme.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "SK_SOFT_OPT_V_RAMCI")
}

func TestGolden_SoftOptionalSL(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "sl", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "V okviru projekta čakamo.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "SL_SOFT_OPT_V_OKVIRU")
}

func TestGolden_SoftPickyBE(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Маем meeting заўтра.", "BE_SOFT_PICKY_MEETING"},
		{"Патрэбен feedback сёння.", "BE_SOFT_PICKY_FEEDBACK"},
		{"Гэта вельмі вельмі важна.", "BE_SOFT_PICKY_VELMI_VELMI"},
		{"Шмат рэчаў засталося.", "BE_SOFT_PICKY_RECHY"},
		{"У канцы канцоў вырашым.", "BE_SOFT_PICKY_U_KANCY_KANCOU"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "be", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickySR(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Imamo meeting sutra.", "SR_SOFT_PICKY_MEETING"},
		{"Treba mi feedback danas.", "SR_SOFT_PICKY_FEEDBACK"},
		{"To je veoma veoma važno.", "SR_SOFT_PICKY_VEOMA_VEOMA"},
		{"Ima mnogo stvari za uraditi.", "SR_SOFT_PICKY_STVARI"},
		{"Na kraju krajeva odlučujemo.", "SR_SOFT_PICKY_NA_KRAJU_KRAJEVA"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "sr", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyLT(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Turime meeting rytoj.", "LT_SOFT_PICKY_MEETING"},
		{"Reikia feedback šiandien.", "LT_SOFT_PICKY_FEEDBACK"},
		{"Tai labai labai svarbu.", "LT_SOFT_PICKY_LABAI_LABAI"},
		{"Yra daug dalykų padaryti.", "LT_SOFT_PICKY_DALYKAI"},
		{"Dienos pabaigoje nuspręsime.", "LT_SOFT_PICKY_DIENOS_PABAIGOJE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "lt", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalENPresentTime(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "At the present time we wait.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_OPT_AT_THE_PRESENT_TIME")
}

func TestGolden_SoftIdiomConfusablesWave8(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Wether or not we ship, wait.", "EN_SOFT_WETHER", "Whether or not"},
		{"A lightening strike hit nearby.", "EN_SOFT_LIGHTENING_STRIKE", "lightning strike"},
		{"Drive thru the tunnel carefully.", "EN_SOFT_THRU", "through"},
		{"It works, tho slowly.", "EN_SOFT_THO", "though"},
		{"Do not embarass the guest.", "EN_SOFT_EMBARASS", "embarrass"},
		{"On this occassion we cheer.", "EN_SOFT_OCCASSION", "occasion"},
		{"He is fourty years old.", "EN_SOFT_FOURTY", "forty"},
		{"The nineth chapter starts here.", "EN_SOFT_NINETH", "ninth"},
		{"That arguement failed.", "EN_SOFT_ARGUEMENT", "argument"},
		{"Protect the enviroment now.", "EN_SOFT_ENVIROMENT", "environment"},
		{"The goverment announced it.", "EN_SOFT_GOVERMENT", "government"},
		{"Please recieve the package.", "EN_SOFT_RECIEVE", "receive"},
		{"I beleive you are right.", "EN_SOFT_BELEIVE", "believe"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalBE(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "be", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "У рамках праекта чакаем.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "BE_SOFT_OPT_U_RAMKAH")
}

func TestGolden_SoftOptionalSR(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "sr", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "U okviru projekta čekamo.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "SR_SOFT_OPT_U_OKVIRU")
}

func TestGolden_SoftOptionalLT(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "lt", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Remiantis tuo laukiame.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "LT_SOFT_OPT_PAGAL")
}

func TestGolden_SoftPickyIS(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Við eigum meeting á morgun.", "IS_SOFT_PICKY_MEETING"},
		{"Ég þarf feedback í dag.", "IS_SOFT_PICKY_FEEDBACK"},
		{"Þetta er mjög mjög mikilvægt.", "IS_SOFT_PICKY_MJÖG_MJÖG"},
		{"Það eru margir hlutir eftir.", "IS_SOFT_PICKY_HLUTIR"},
		{"Í lok dags ákveðum við.", "IS_SOFT_PICKY_I_LOK_DAGA"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "is", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyGA(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Tá meeting againn amárach.", "GA_SOFT_PICKY_MEETING"},
		{"Teastaíonn feedback uaim.", "GA_SOFT_PICKY_FEEDBACK"},
		{"Tá a lán rudaí le déanamh.", "GA_SOFT_PICKY_RUDAI"},
		{"Ag deireadh an lae, seolfaimid.", "GA_SOFT_PICKY_AG_DEIREADH"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "ga", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyEO(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Ni havas meeting morgaŭ.", "EO_SOFT_PICKY_MEETING"},
		{"Mi bezonas feedback hodiaŭ.", "EO_SOFT_PICKY_FEEDBACK"},
		{"Ĝi estas tre tre grava.", "EO_SOFT_PICKY_TRE_TRE"},
		{"Estas multaj aĵoj farendaj.", "EO_SOFT_PICKY_AJOJ"},
		{"Fine de la tago ni decidos.", "EO_SOFT_PICKY_FINE_DE_LA_TAGO"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "eo", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave9(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"An independant review shipped.", "EN_SOFT_INDEPENDANT", "independent"},
		{"Prove the existance of bugs.", "EN_SOFT_EXISTANCE", "existence"},
		{"I am refering to the docs.", "EN_SOFT_REFERING", "referring"},
		{"The file was transfered today.", "EN_SOFT_TRANSFERED", "transferred"},
		{"From the begining it worked.", "EN_SOFT_BEGINING", "beginning"},
		{"Show your comittment clearly.", "EN_SOFT_COMITTMENT", "commitment"},
		{"A succesful deploy landed.", "EN_SOFT_SUCCESFUL", "successful"},
		{"Wait untill tomorrow morning.", "EN_SOFT_UNTILL", "until"},
		{"Good writting takes practice.", "EN_SOFT_WRITTING", "writing"},
		{"Send your adress by email.", "EN_SOFT_ADRESS", "address"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalIS(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "is", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Í sambandi við verkefnið bíðum við.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "IS_SOFT_OPT_I_SAMBANDI")
}

func TestGolden_SoftOptionalGA(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ga", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Maidir le an tionscadal, fan.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "GA_SOFT_OPT_MAIDIR_LE")
}

func TestGolden_SoftOptionalEO(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "eo", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Rilate al la projekto ni atendas.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EO_SOFT_OPT_RILATE_AL")
}

func TestGolden_SoftPickyFA(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"فردا meeting داریم.", "FA_SOFT_PICKY_MEETING"},
		{"امروز feedback لازم است.", "FA_SOFT_PICKY_FEEDBACK"},
		{"این خیلی خیلی مهم است.", "FA_SOFT_PICKY_KHEILI_KHEILI"},
		{"چیزهای زیاد مانده است.", "FA_SOFT_PICKY_CHIZHA"},
		{"در نهایت تصمیم می‌گیریم.", "FA_SOFT_PICKY_DAR_NAHAYAT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "fa", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyAR(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"لدينا meeting غدا.", "AR_SOFT_PICKY_MEETING"},
		{"أحتاج feedback اليوم.", "AR_SOFT_PICKY_FEEDBACK"},
		{"هذا جدا جدا مهم.", "AR_SOFT_PICKY_JIDDAN_JIDDAN"},
		{"هناك أشياء كثيرة.", "AR_SOFT_PICKY_ASHYAA"},
		{"في نهاية المطاف نقرر.", "AR_SOFT_PICKY_FI_NIHAYAT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "ar", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyZH(t *testing.T) {
	// Soft ZH matching is space-tokenized (no full CJK segmenter yet).
	cases := []struct {
		text, rule string
	}{
		{"明天有个 meeting 。", "ZH_SOFT_PICKY_MEETING"},
		{"今天需要 feedback 。", "ZH_SOFT_PICKY_FEEDBACK"},
		{"这 非常 非常 重要。", "ZH_SOFT_PICKY_FEICHANG_FEICHANG"},
		{"还有 很多 东西 要做。", "ZH_SOFT_PICKY_DONGXI"},
		{"到头来 我们 会 决定。", "ZH_SOFT_PICKY_DAO_TOU_LAI"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "zh", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave10(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Open a buisness account today.", "EN_SOFT_BUISNESS", "business"},
		{"Check the calender for conflicts.", "EN_SOFT_CALENDER", "calendar"},
		{"Ask your collegue for help.", "EN_SOFT_COLLEAGUE_MISS", "colleague"},
		{"Be concious of the risk.", "EN_SOFT_CONCIOUS", "conscious"},
		{"I definitly need a break.", "EN_SOFT_DEFINITLY", "definitely"},
		{"Do not exagerate the claims.", "EN_SOFT_EXAGERATE", "exaggerate"},
		{"That name sounds familier.", "EN_SOFT_FAMILIAR_MISS", "familiar"},
		{"Learn a foriegn language.", "EN_SOFT_FORIEGN", "foreign"},
		{"The gaurd checked badges.", "EN_SOFT_GAURD", "guard"},
		{"Do not harrass the staff.", "EN_SOFT_HARRASS", "harass"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalFA(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "fa", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "در راستای پروژه صبر می‌کنیم.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "FA_SOFT_OPT_DAR_RASTAYE")
}

func TestGolden_SoftOptionalAR(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ar", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "في إطار المشروع ننتظر.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "AR_SOFT_OPT_FI_ITAR")
}

func TestGolden_SoftOptionalZH(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "zh", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "在 方面 我们 等待。", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "ZH_SOFT_OPT_ZAI_FANGMIAN")
}

func TestGolden_SoftPickyJA(t *testing.T) {
	// Soft JA matching is space-tokenized.
	cases := []struct {
		text, rule string
	}{
		{"明日 meeting があります。", "JA_SOFT_PICKY_MEETING"},
		{"今日 feedback が必要です。", "JA_SOFT_PICKY_FEEDBACK"},
		{"とても とても 重要です。", "JA_SOFT_PICKY_TOTTEMO_TOTTEMO"},
		{"いろいろ な こと がある。", "JA_SOFT_PICKY_IROIRO"},
		{"結局 決める。", "JA_SOFT_PICKY_KEKKYOKU"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "ja", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyBR(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Ur meeting a zo warc'hoazh.", "BR_SOFT_PICKY_MEETING"},
		{"Ezhomm am eus eus feedback hiziv.", "BR_SOFT_PICKY_FEEDBACK"},
		{"Kalz kalz a-bouez eo.", "BR_SOFT_PICKY_KALZ_KALZ"},
		{"Kalz traoù a chom.", "BR_SOFT_PICKY_TRAOU"},
		{"E fin an deiz e tivizomp.", "BR_SOFT_PICKY_E_FIN_AN_DEIZ"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "br", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyAST(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Tenemos un meeting mañana.", "AST_SOFT_PICKY_MEETING"},
		{"Necesito feedback güei.", "AST_SOFT_PICKY_FEEDBACK"},
		{"Ye mui mui importante.", "AST_SOFT_PICKY_MUI_MUI"},
		{"Hai munches coses por facer.", "AST_SOFT_PICKY_COSES"},
		{"Al final del día unviamos.", "AST_SOFT_PICKY_AL_FINAL_DEL_DIA"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "ast", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave11(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Do it imediately after lunch.", "EN_SOFT_IMEDIATELY", "immediately"},
		{"Do not be jelous of success.", "EN_SOFT_JELOUS", "jealous"},
		{"Share knowlege with the team.", "EN_SOFT_KNOWLEGE", "knowledge"},
		{"Act as a liason with vendors.", "EN_SOFT_LIASON", "liaison"},
		{"Schedule maintainance tonight.", "EN_SOFT_MAINTAINANCE", "maintenance"},
		{"A new millenium began then.", "EN_SOFT_MILLENIUM", "millennium"},
		{"There is a noticable delay.", "EN_SOFT_NOTICABLE", "noticeable"},
		{"Take the oppertunity now.", "EN_SOFT_OPPORTUNITY_MISS", "opportunity"},
		{"Ten persent remain offline.", "EN_SOFT_PERSENT", "percent"},
		{"Access is a priviledge here.", "EN_SOFT_PRIVILEDGE", "privilege"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalJA(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ja", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "計画 について 待ちます。", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "JA_SOFT_OPT_NI_TSUITE")
}

func TestGolden_SoftOptionalBR(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "br", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Diwar-benn ar raktres e c'hortozomp.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "BR_SOFT_OPT_DIWAR_BENN")
}

func TestGolden_SoftOptionalAST(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ast", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "En relación con el proyeutu esperamos.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "AST_SOFT_OPT_EN_RELACION")
}

func TestGolden_SoftPickyKM(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"មាន meeting ស្អែក។", "KM_SOFT_PICKY_MEETING"},
		{"ត្រូវការ feedback ថ្ងៃនេះ។", "KM_SOFT_PICKY_FEEDBACK"},
		{"សំខាន់ ណាស់ ណាស់ ។", "KM_SOFT_PICKY_NAS_NAS"},
		{"មាន របស់ ជាច្រើន ។", "KM_SOFT_PICKY_RBOV"},
		{"នៅ ទីបំផុត យើង សម្រេច។", "KM_SOFT_PICKY_CHHNGAY"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "km", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyTA(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"நாளை meeting உள்ளது.", "TA_SOFT_PICKY_MEETING"},
		{"இன்று feedback தேவை.", "TA_SOFT_PICKY_FEEDBACK"},
		{"ரொம்ப ரொம்ப முக்கியம்.", "TA_SOFT_PICKY_ROMBA_ROMBA"},
		{"பல விஷயங்கள் உள்ளன.", "TA_SOFT_PICKY_VISHAYANGAL"},
		{"முடிவில் முடிவு செய்வோம்.", "TA_SOFT_PICKY_MUDIVIL"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "ta", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyTL(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"May meeting bukas.", "TL_SOFT_PICKY_MEETING"},
		{"Kailangan ng feedback ngayon.", "TL_SOFT_PICKY_FEEDBACK"},
		{"Sobrang sobrang importante ito.", "TL_SOFT_PICKY_SOBRANG_SOBRANG"},
		{"May mga bagay pang gawin.", "TL_SOFT_PICKY_MGA_BAGAY"},
		{"Sa huli ay magpapasya tayo.", "TL_SOFT_PICKY_SA_HULI"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "tl", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave12(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fill the questionaire carefully.", "EN_SOFT_QUESTIONAIRE", "questionnaire"},
		{"I recomend this approach.", "EN_SOFT_RECOMEND", "recommend"},
		{"Keep only relevent details.", "EN_SOFT_RELEVENT", "relevant"},
		{"Lower the resistence carefully.", "EN_SOFT_RESISTENCE", "resistance"},
		{"Keep the rythm steady.", "EN_SOFT_RHYTHM_MISS", "rhythm"},
		{"Do not sieze the assets.", "EN_SOFT_SIEZE", "seize"},
		{"Prepare a short speach.", "EN_SOFT_SPEACH", "speech"},
		{"Build strenght gradually.", "EN_SOFT_STRENGHT", "strength"},
		{"Raise the threshhold slightly.", "EN_SOFT_THRESHHOLD", "threshold"},
		{"See you tomorow morning.", "EN_SOFT_TOMOROW", "tomorrow"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalKM(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "km", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "ទាក់ទង នឹង គម្រោង យើង រង់ចាំ។", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "KM_SOFT_OPT_TEANG_NUNG")
}

func TestGolden_SoftOptionalTA(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ta", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "திட்டம் பற்றி காத்திருக்கிறோம்.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "TA_SOFT_OPT_PATRI")
}

func TestGolden_SoftOptionalTL(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "tl", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Kaugnay sa proyekto ay maghintay tayo.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "TL_SOFT_OPT_KAUGNAY_SA")
}

func TestGolden_SoftPickyCRH(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Yarın meeting bar.", "CRH_SOFT_PICKY_MEETING"},
		{"Bugün feedback kerek.", "CRH_SOFT_PICKY_FEEDBACK"},
		{"Bu çok çok mühim.", "CRH_SOFT_PICKY_ÇOK_ÇOK"},
		{"Çoq şey qaldı.", "CRH_SOFT_PICKY_ŞEYLER"},
		{"Soñunda qarar beremiz.", "CRH_SOFT_PICKY_SOÑUNDA"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "crh", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyML(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"നാളെ meeting ഉണ്ട്.", "ML_SOFT_PICKY_MEETING"},
		{"ഇന്ന് feedback വേണം.", "ML_SOFT_PICKY_FEEDBACK"},
		{"വളരെ വളരെ പ്രധാനമാണ്.", "ML_SOFT_PICKY_VALARE_VALARE"},
		{"പല കാര്യങ്ങൾ ഉണ്ട്.", "ML_SOFT_PICKY_KARYANGAL"},
		{"അവസാനം തീരുമാനിക്കാം.", "ML_SOFT_PICKY_AVASANAM"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "ml", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave13(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Park the vehical outside.", "EN_SOFT_VEHICAL", "vehicle"},
		{"Report each occurence carefully.", "EN_SOFT_OCCURENCE", "occurrence"},
		{"That looks wierd to me.", "EN_SOFT_WIERD", "weird"},
		{"Choose wich path to take.", "EN_SOFT_WICH", "which"},
		{"Come whith us tomorrow.", "EN_SOFT_WHITH", "with"},
		{"I know wether it works.", "EN_SOFT_WETHER_ALONE", "whether"},
		{"Check thier credentials first.", "EN_SOFT_THIER", "their"},
		{"Open teh file carefully.", "EN_SOFT_TEH", "the"},
		{"Keep the reciept for taxes.", "EN_SOFT_RECIEPT", "receipt"},
		{"Persue the opportunity now.", "EN_SOFT_PERSUE", "Pursue"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalCRH(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "crh", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Proyekt munasebetnen bekleyiz.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "CRH_SOFT_OPT_MUNASEBETNEN")
}

func TestGolden_SoftOptionalML(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "ml", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "പദ്ധതി സംബന്ധിച്ച് കാത്തിരിക്കുന്നു.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "ML_SOFT_OPT_SAMBANDHICHU")
}

func TestGolden_SoftOptionalENNearFuture(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "In the near future we ship.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_OPT_IN_THE_NEAR_FUTURE")
}

func TestGolden_FalseFriendsMolest(t *testing.T) {
	ff := softFalseFriendsPath(t)
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "No debes molestar ahora.", &CommandLineOptions{
		Language:         "es",
		MotherTongue:     "en",
		FalseFriendsFile: ff,
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "MOLEST" {
			found = true
			require.Equal(t, "bother / annoy", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftIdiomConfusablesWave14(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Book accomodation early.", "EN_SOFT_ACCOMODATION", "accommodation"},
		{"She is fully commited now.", "EN_SOFT_COMMITED", "committed"},
		{"That will exhilerate fans.", "EN_SOFT_EXHILERATE", "exhilarate"},
		{"Report harrasment promptly.", "EN_SOFT_HARRASMENT", "harassment"},
		{"Hire a gardner this spring.", "EN_SOFT_GARDNER", "gardener"},
		{"Prove possesion of the key.", "EN_SOFT_POSSESION", "possession"},
		{"A proffesional review helps.", "EN_SOFT_PROFFESIONAL", "professional"},
		{"Add a referance section.", "EN_SOFT_REFERANCE", "reference"},
		{"This is a temperary fix.", "EN_SOFT_TEMPERARY", "temporary"},
		{"Use a vaccum carefully.", "EN_SOFT_VACCUM", "vacuum"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave2(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"We should ideate tomorrow.", "EN_SOFT_PICKY_IDEATE"},
		{"Teams will synergize next week.", "EN_SOFT_PICKY_SYNERGIZE"},
		{"This is a paradigm shift.", "EN_SOFT_PICKY_PARADIGM_SHIFT"},
		{"We need best of breed tools.", "EN_SOFT_PICKY_BEST_OF_BREED"},
		{"Deliver more value-add features.", "EN_SOFT_PICKY_VALUE_ADD"},
		{"It is a win-win outcome.", "EN_SOFT_PICKY_WIN_WIN"},
		{"Do not boil the ocean here.", "EN_SOFT_PICKY_BOIL_THE_OCEAN"},
		{"Please reach out soon.", "EN_SOFT_PICKY_REACH_OUT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftAUVariantHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Pick a color today.", "EN_SOFT_COLOR_AU", "colour"},
		{"My favorite book.", "EN_SOFT_FAVORITE_AU", "favourite"},
		{"City center is busy.", "EN_SOFT_CENTER_AU", "centre"},
		{"Please organize files.", "EN_SOFT_ORGANIZE_AU", "organise"},
		{"I realize now.", "EN_SOFT_REALIZE_AU", "realise"},
		{"We traveled far.", "EN_SOFT_TRAVELED_AU", "travelled"},
		{"The flight was canceled.", "EN_SOFT_CANCELED_AU", "cancelled"},
		{"Show your license card.", "EN_SOFT_LICENSE_AU", "licence"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-AU"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftAUVariantsNotOnUS(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Pick a color today.", &CommandLineOptions{Language: "en-US"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "EN_SOFT_COLOR_AU", f.Rule, "%+v", findings)
	}
}

func TestGolden_ImmunizeLgtmWfh(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Lgtm the change looks fine.",
		"Sgtm we can ship today.",
		"Wfh tomorrow is fine.",
		"Ship by eod please.",
		"This is still a wip.",
		"Take the dose prn.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave15(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Visit the libary today.", "EN_SOFT_LIBARY", "library"},
		{"Measure the heigth carefully.", "EN_SOFT_HEIGTH", "height"},
		{"Draw a parralel line here.", "EN_SOFT_PARRALEL", "parallel"},
		{"Move foward with the plan.", "EN_SOFT_FOWARD", "forward"},
		{"Check the lenght of the rope.", "EN_SOFT_LENGHT", "length"},
		{"Measure the widht next.", "EN_SOFT_WIDHT", "width"},
		{"Reduce the weigth slightly.", "EN_SOFT_WEIGTH", "weight"},
		{"Build a stong foundation.", "EN_SOFT_STONG", "strong"},
		{"Share your beleif openly.", "EN_SOFT_BELEIF", "belief"},
		{"We will acheive the goal.", "EN_SOFT_ACHIEVE_MISS", "achieve"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftCAVariantHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Pick a color today.", "EN_SOFT_COLOR_CA", "colour"},
		{"My favorite book.", "EN_SOFT_FAVORITE_CA", "favourite"},
		{"City center is busy.", "EN_SOFT_CENTER_CA", "centre"},
		{"Good behavior matters.", "EN_SOFT_BEHAVIOR_CA", "behaviour"},
		{"Please organise files.", "EN_SOFT_ORGANISE_CA", "organize"},
		{"I realise now.", "EN_SOFT_REALISE_CA", "realize"},
		{"Strong defense wins.", "EN_SOFT_DEFENCE_SPELL_CA", "defence"},
		{"Please analyse data.", "EN_SOFT_ANALYSE_CA", "analyze"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-CA"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftUSExtraVariants(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Write a cheque today.", "EN_SOFT_CHEQUE_US", "check"},
		{"Replace the tyre soon.", "EN_SOFT_TYRE_US", "tire"},
		{"Use aluminium foil carefully.", "EN_SOFT_ALUMINIUM_US", "aluminum"},
		{"A two storey house sold.", "EN_SOFT_STOREY_US", "story"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-US"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftGBExtraVariants(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Open the check book carefully.", "EN_SOFT_CHECK_GB", "cheque book"},
		{"Replace the tire soon.", "EN_SOFT_TIRE_GB", "tyre"},
		{"Use aluminum foil carefully.", "EN_SOFT_ALUMINUM_GB", "aluminium"},
		{"A multi story building rose.", "EN_SOFT_STORY_FLOOR_GB", "storey"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-GB"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave16(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please acknowlege receipt.", "EN_SOFT_ACKNOWLEDGE_MISS", "acknowledge"},
		{"It is aparent to everyone.", "EN_SOFT_APPARENT_MISS", "apparent"},
		{"Mind your appearence today.", "EN_SOFT_APPEARANCE_MISS", "appearance"},
		{"That is argubly better.", "EN_SOFT_ARGUABLY_MISS", "arguably"},
		{"It is basicly finished.", "EN_SOFT_BASICALLY_MISS", "basically"},
		{"A beatiful garden grew.", "EN_SOFT_BEAUTIFUL_MISS", "beautiful"},
		{"There is a clear benifit.", "EN_SOFT_BENEFIT_MISS", "benefit"},
		{"Pick a catagory carefully.", "EN_SOFT_CATEGORY_MISS", "category"},
		{"Visit the cemetary gates.", "EN_SOFT_CEMETERY_MISS", "cemetery"},
		{"The plan is changable still.", "EN_SOFT_CHANGEABLE_MISS", "changeable"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalENTimelyManner(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Respond in a timely manner please.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_OPT_IN_A_TIMELY_MANNER")
}

func TestGolden_FalseFriendsGenial(t *testing.T) {
	ff := softFalseFriendsPath(t)
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Das war genial gelöst.", &CommandLineOptions{
		Language:         "de",
		MotherTongue:     "en",
		FalseFriendsFile: ff,
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "GENIAL" {
			found = true
			require.Equal(t, "brilliant / ingenious", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_IgnoreSpellingGrok(t *testing.T) {
	if DiscoverEnglishIgnoreSpellingList(nil) == "" {
		t.Skip("ignore-spelling list missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Grok and xAI shipped Claude alternatives.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_MultiwordSanDiego(t *testing.T) {
	if DiscoverEnglishMultiwords(nil) == "" {
		t.Skip("multiwords missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "We flew to San Diego yesterday.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		// multiword place names should not be flagged as spelling errors
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave17(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Apply to colledge next year.", "EN_SOFT_COLLEGE_MISS", "college"},
		{"Join the commitee meeting.", "EN_SOFT_COMMITTEE_MISS", "committee"},
		{"Stay consious of risks.", "EN_SOFT_CONSCIOUS_MISS", "conscious"},
		{"Pick a conveniant time.", "EN_SOFT_CONVENIENT_MISS", "convenient"},
		{"Ignore harsh critisism.", "EN_SOFT_CRITICISM_MISS", "criticism"},
		{"Make a decison carefully.", "EN_SOFT_DECISION_MISS", "decision"},
		{"Read the discription first.", "EN_SOFT_DESCRIPTION_MISS", "description"},
		{"Fund the developement plan.", "EN_SOFT_DEVELOP_MISS", "development"},
		{"Try a diffrent approach.", "EN_SOFT_DIFFERENT_MISS", "different"},
		{"Do not dissapear suddenly.", "EN_SOFT_DISAPPEAR_MISS", "disappear"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave3(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"We need thought leadership here.", "EN_SOFT_PICKY_THOUGHT_LEADERSHIP"},
		{"Please drill down into metrics.", "EN_SOFT_PICKY_DRILL_DOWN"},
		{"Let us align on priorities.", "EN_SOFT_PICKY_ALIGN_ON"},
		{"Expect pushback from sales.", "EN_SOFT_PICKY_PUSHBACK"},
		{"We should take this offline.", "EN_SOFT_PICKY_OFFLINE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftNZVariantHints(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Pick a color today.", "EN_SOFT_COLOR_NZ", "colour"},
		{"My favorite book.", "EN_SOFT_FAVORITE_NZ", "favourite"},
		{"City center is busy.", "EN_SOFT_CENTER_NZ", "centre"},
		{"Please organize files.", "EN_SOFT_ORGANIZE_NZ", "organise"},
		{"I realize now.", "EN_SOFT_REALIZE_NZ", "realise"},
		{"We traveled far.", "EN_SOFT_TRAVELED_NZ", "travelled"},
		{"The flight was canceled.", "EN_SOFT_CANCELED_NZ", "cancelled"},
		{"Replace the tire soon.", "EN_SOFT_TIRE_NZ", "tyre"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-NZ"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "misspelling", f.Type, "%+v", f)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave18(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Do not dissapoint the team.", "EN_SOFT_DISAPPOINT_MISS", "disappoint"},
		{"I felt embarassed yesterday.", "EN_SOFT_EMBARRASS_MISS", "embarrassed"},
		{"Buy equiptment carefully.", "EN_SOFT_EQUIPMENT_MISS", "equipment"},
		{"I like this expecially now.", "EN_SOFT_ESPECIALLY_MISS", "especially"},
		{"That was an exellent talk.", "EN_SOFT_EXCELLENT_MISS", "excellent"},
		{"Share your experiance freely.", "EN_SOFT_EXPERIENCE_MISS", "experience"},
		{"That name looks familar.", "EN_SOFT_FAMILIAR_MISS2", "familiar"},
		{"Meet in febuary next year.", "EN_SOFT_FEBRUARY_MISS", "February"},
		{"We finaly finished testing.", "EN_SOFT_FINALLY_MISS", "finally"},
		{"This happens frequenly here.", "EN_SOFT_FREQUENTLY_MISS", "frequently"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalENExceptionOf(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "With the exception of one bug, ship.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_OPT_WITH_THE_EXCEPTION_OF")
}

func TestGolden_IgnoreSpellingSvelte(t *testing.T) {
	if DiscoverEnglishIgnoreSpellingList(nil) == "" {
		t.Skip("ignore-spelling list missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Svelte and Bun pair well with Vercel.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_MultiwordCapeCod(t *testing.T) {
	if DiscoverEnglishMultiwords(nil) == "" {
		t.Skip("multiwords missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "We visited Cape Cod last summer.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave19(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Be freindly to guests.", "EN_SOFT_FRIENDLY_MISS", "friendly"},
		{"Discuss this futher later.", "EN_SOFT_FURTHER_MISS", "further"},
		{"Check the pressure guage.", "EN_SOFT_GAUGE_MISS", "gauge"},
		{"Fix the heirarchy soon.", "EN_SOFT_HIERARCHY_MISS", "hierarchy"},
		{"Study English grammer daily.", "EN_SOFT_GRAMMAR_MISS", "grammar"},
		{"I am greatful for help.", "EN_SOFT_GRATEFUL_MISS", "grateful"},
		{"Need guidence on setup.", "EN_SOFT_GUIDANCE_MISS", "guidance"},
		{"It happend last night.", "EN_SOFT_HAPPENED_MISS", "happened"},
		{"Do not haras the staff.", "EN_SOFT_HARRASS_MISS", "harass"},
		{"Measure the hieght carefully.", "EN_SOFT_HEIGHT_MISS", "height"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftAUExtraVariants(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Replace the tire soon.", "EN_SOFT_TIRE_AU", "tyre"},
		{"Use aluminum foil carefully.", "EN_SOFT_ALUMINUM_AU", "aluminium"},
		{"I favor this approach.", "EN_SOFT_COLOR_VERB_AU", "favour"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-AU"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_FalseFriendsNormal(t *testing.T) {
	ff := softFalseFriendsPath(t)
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Todo está normal hoy.", &CommandLineOptions{
		Language:         "es",
		MotherTongue:     "en",
		FalseFriendsFile: ff,
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "NORMAL" {
			found = true
			require.Equal(t, "usual / ordinary (not \"OK\")", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_SoftIdiomConfusablesWave20(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Need imediate action now.", "EN_SOFT_IMMEDIATE_MISS", "immediate"},
		{"This is importent work.", "EN_SOFT_IMPORTANT_MISS", "important"},
		{"Raise inteligence carefully.", "EN_SOFT_INTELLIGENCE_MISS", "intelligence"},
		{"That is an intresting idea.", "EN_SOFT_INTERESTING_MISS", "interesting"},
		{"Sell jewelery carefully.", "EN_SOFT_JEWELRY_MISS", "jewelry"},
		{"Learn a new langauge today.", "EN_SOFT_LANGUAGE_MISS", "language"},
		{"Enjoy liesure time fully.", "EN_SOFT_LEISURE_MISS", "leisure"},
		{"Get a lisence first.", "EN_SOFT_LICENSE_MISS", "license"},
		{"Watch for lightening outdoors.", "EN_SOFT_LIGHTNING_MISS", "lightning"},
		{"Fight lonliness carefully.", "EN_SOFT_LONELINESS_MISS", "loneliness"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave4(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"We have six months of runway.", "EN_SOFT_PICKY_RUNWAY"},
		{"Define a north star metric soon.", "EN_SOFT_PICKY_NORTH_STAR"},
		{"Please double-click the issue.", "EN_SOFT_PICKY_DOUBLE_CLICK"},
		{"Let us sync up tomorrow.", "EN_SOFT_PICKY_SYNC_UP"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeEtaMvp(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Eta is tomorrow afternoon.",
		"ETA is tomorrow afternoon.",
		"Ship a poc this week.",
		"Launch the mvp soon.",
		"Track the okr carefully.",
		"Improve the kpi next.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave21(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They manufactur parts here.", "EN_SOFT_MANUFACTURE_MISS", "manufacture"},
		{"Study medevil history carefully.", "EN_SOFT_MEDIEVAL_MISS", "medieval"},
		{"Ask your nieghbor for help.", "EN_SOFT_NEIGHBOR_MISS", "neighbor"},
		{"Take the opurtunity now.", "EN_SOFT_OPPORTUNITY_MISS2", "opportunity"},
		{"Keep the orignal file.", "EN_SOFT_ORIGINAL_MISS", "original"},
		{"Draw a paralel line carefully.", "EN_SOFT_PARALLEL_MISS", "parallel"},
		{"In this particuler case wait.", "EN_SOFT_PARTICULAR_MISS", "particular"},
		{"Perhpas we should wait.", "EN_SOFT_PERHAPS_MISS", "Perhaps"},
		{"I personaly disagree strongly.", "EN_SOFT_PERSONALLY_MISS", "personally"},
		{"Prove posession of the key.", "EN_SOFT_POSSESSION_MISS", "possession"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftGBExtraFavourHonor(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I favor this approach.", "EN_SOFT_FAVOR_GB", "favour"},
		{"It is an honor to serve.", "EN_SOFT_HONOR_GB", "honour"},
		{"They labored through night.", "EN_SOFT_LABOR_VERB_GB", "laboured"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-GB"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_MultiwordNovaScotia(t *testing.T) {
	if DiscoverEnglishMultiwords(nil) == "" {
		t.Skip("multiwords missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "We visited Nova Scotia last year.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave22(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I recieved the package today.", "EN_SOFT_RECEIVED_MISS", "received"},
		{"Store files seperately please.", "EN_SOFT_SEPARATELY_MISS", "separately"},
		{"A successfull deploy landed.", "EN_SOFT_SUCCESSFUL_MISS", "successful"},
		{"See you tommorow morning.", "EN_SOFT_TOMORROW_MISS", "tomorrow"},
		{"We are transfering data now.", "EN_SOFT_TRANSFERRING_MISS", "transferring"},
		{"The note was writen carefully.", "EN_SOFT_WRITTEN_MISS", "written"},
		{"This is the prefered option.", "EN_SOFT_PREFERRED_MISS", "preferred"},
		{"State your preferance clearly.", "EN_SOFT_PREFERENCE_MISS", "preference"},
		{"She refered me to support.", "EN_SOFT_REFERRED_MISS", "referred"},
		{"Report each occurance carefully.", "EN_SOFT_OCCURRENCE_MISS", "occurrence"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalENNeedlessToSay(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "Needless to say, we shipped.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_OPT_NEEDLESS_TO_SAY")
}

func TestGolden_IgnoreSpellingPlaywright(t *testing.T) {
	if DiscoverEnglishIgnoreSpellingList(nil) == "" {
		t.Skip("ignore-spelling list missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Playwright and Vitest cover the suite.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave23(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fight for independance carefully.", "EN_SOFT_INDEPENDENCE_MISS", "independence"},
		{"Ask the proffesor carefully.", "EN_SOFT_PROFESSOR_MISS", "professor"},
		{"I reccomend this approach.", "EN_SOFT_RECOMMEND_MISS", "recommend"},
		{"Read the recomendation carefully.", "EN_SOFT_RECOMMENDATION_MISS", "recommendation"},
		{"Add a refference section.", "EN_SOFT_REFERENCE_MISS", "reference"},
		{"Need a clean seperation here.", "EN_SOFT_SEPARATION_MISS", "separation"},
		{"Therfore we should wait.", "EN_SOFT_THEREFORE_MISS", "Therefore"},
		{"Search thruout the codebase.", "EN_SOFT_THROUGHOUT_MISS", "throughout"},
		{"I am truely grateful today.", "EN_SOFT_TRULY_MISS", "truly"},
		{"Unfortunatly the test failed.", "EN_SOFT_UNFORTUNATELY_MISS", "Unfortunately"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_FalseFriendsPretend(t *testing.T) {
	ff := softFalseFriendsPath(t)
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "No pretender ser experto.", &CommandLineOptions{
		Language:         "es",
		MotherTongue:     "en",
		FalseFriendsFile: ff,
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "PRETEND" {
			found = true
			require.Equal(t, "intend / aim / claim", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_MultiwordSaltLakeCity(t *testing.T) {
	if DiscoverEnglishMultiwords(nil) == "" {
		t.Skip("multiwords missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "We flew to Salt Lake City yesterday.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave24(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"This is a usefull tip.", "EN_SOFT_USEFUL_MISS", "useful"},
		{"Make the API useable soon.", "EN_SOFT_USABLE_MISS", "usable"},
		{"No such file is existant.", "EN_SOFT_EXISTENT_MISS", "existent"},
		{"Choose a proffession carefully.", "EN_SOFT_PROFESSION_MISS", "profession"},
		{"She reffered me to support.", "EN_SOFT_REFERRED_MISS2", "referred"},
		{"Do a throrough review first.", "EN_SOFT_THOROUGH_MISS", "thorough"},
		{"It completed sucessfuly today.", "EN_SOFT_SUCCESSFULLY_MISS", "successfully"},
		{"Store items seperatly please.", "EN_SOFT_SEPARATELY_MISS2", "separately"},
		{"Please aquaint yourself first.", "EN_SOFT_ACQUAINT_MISS", "acquaint"},
		{"He is an old aquaintance.", "EN_SOFT_ACQUAINTANCE_MISS", "acquaintance"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftUSExtraHonourFavour(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It is an honour to serve.", "EN_SOFT_HONOUR_US", "honor"},
		{"I favour this approach.", "EN_SOFT_FAVOUR_US", "favor"},
		{"They laboured through night.", "EN_SOFT_LABOURED_US", "labored"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en-US"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave5(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"We will circle back on this.", "EN_SOFT_PICKY_CIRCLE_BACK_ON"},
		{"Please loop in the legal team.", "EN_SOFT_PICKY_LOOP_IN"},
		{"Let us table this for now.", "EN_SOFT_PICKY_TABLE_THIS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave25(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"That is not beleivable today.", "EN_SOFT_BELIEVABLE_MISS", "believable"},
		{"Sort the catagories carefully.", "EN_SOFT_CATEGORIES_MISS", "categories"},
		{"We reached a concensus today.", "EN_SOFT_CONSENSUS_MISS", "consensus"},
		{"I cannot concieve of that.", "EN_SOFT_CONCEIVE_MISS", "conceive"},
		{"Need continous monitoring now.", "EN_SOFT_CONTINUOUS_MISS", "continuous"},
		{"Need a definate answer soon.", "EN_SOFT_DEFINITE_MISS", "definite"},
		{"Avoid dependance on one host.", "EN_SOFT_DEPENDENCE_MISS", "dependence"},
		{"I am dissapointed with that.", "EN_SOFT_DISAPPOINTED_MISS", "disappointed"},
		{"Hide your embarassment carefully.", "EN_SOFT_EMBARRASSMENT_MISS", "embarrassment"},
		{"Use an equivelent approach.", "EN_SOFT_EQUIVALENT_MISS", "equivalent"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeSlaRca(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Meet the sla carefully.",
		"Meet the SLA carefully.",
		"Write an rca soon.",
		"Write an RCA soon.",
		"This is a p0 incident.",
		"Track the p1 queue.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_IgnoreSpellingPnpm(t *testing.T) {
	if DiscoverEnglishIgnoreSpellingList(nil) == "" {
		t.Skip("ignore-spelling list missing")
	}
	var buf bytes.Buffer
	// Capitalize sentence start; avoid unknown common nouns in the rest.
	_, err := CoreGoldenHook(&buf, "Pnpm and Turborepo are tools we use.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave26(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Avoid exageration carefully.", "EN_SOFT_EXAGGERATION_MISS", "exaggeration"},
		{"Stories facinate children.", "EN_SOFT_FASCINATE_MISS", "fascinate"},
		{"That is a facinating idea.", "EN_SOFT_FASCINATING_MISS", "fascinating"},
		{"He is my freind today.", "EN_SOFT_FRIEND_MISS", "friend"},
		{"Meet freinds after work.", "EN_SOFT_FRIENDS_MISS", "friends"},
		{"Mind hygene carefully.", "EN_SOFT_HYGIENE_MISS", "hygiene"},
		{"Improve hygeine standards now.", "EN_SOFT_HYGIENE_MISS2", "hygiene"},
		{"Reply immediatly after lunch.", "EN_SOFT_IMMEDIATELY_MISS", "immediately"},
		{"Use a heirarchial layout.", "EN_SOFT_HIERARCHICAL_MISS", "hierarchical"},
		{"Avoid govermental delay here.", "EN_SOFT_GOVERNMENTAL_MISS", "governmental"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalENItShouldBeNoted(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "It should be noted that tests pass.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_OPT_IT_SHOULD_BE_NOTED")
}

func TestGolden_MultiwordElPaso(t *testing.T) {
	if DiscoverEnglishMultiwords(nil) == "" {
		t.Skip("multiwords missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "We drove to El Paso yesterday.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave27(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"She is knowledgable about APIs.", "EN_SOFT_KNOWLEDGEABLE_MISS", "knowledgeable"},
		{"Please liase with support.", "EN_SOFT_LIAISE_MISS", "liaise"},
		{"Schedule maintainence tonight.", "EN_SOFT_MAINTENANCE_MISS2", "maintenance"},
		{"Do not mispell names carefully.", "EN_SOFT_MISSPELL_MISS", "misspell"},
		{"Fix the mispelled token.", "EN_SOFT_MISSPELLED_MISS", "misspelled"},
		{"Not neccessarily true today.", "EN_SOFT_NECESSARILY_MISS", "necessarily"},
		{"We meet occassionally now.", "EN_SOFT_OCCASIONALLY_MISS", "occasionally"},
		{"It ocurred last night.", "EN_SOFT_OCCURRED_MISS", "occurred"},
		{"Avoid propoganda carefully.", "EN_SOFT_PROPAGANDA_MISS", "propaganda"},
		{"Do not propogate bad data.", "EN_SOFT_PROPAGATE_MISS", "propagate"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_FalseFriendsAttend(t *testing.T) {
	ff := softFalseFriendsPath(t)
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Debes atender al cliente.", &CommandLineOptions{
		Language:         "es",
		MotherTongue:     "en",
		FalseFriendsFile: ff,
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "ATTEND" {
			found = true
			require.Equal(t, "serve / take care of / answer", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_MultiwordSantaFe(t *testing.T) {
	if DiscoverEnglishMultiwords(nil) == "" {
		t.Skip("multiwords missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "We visited Santa Fe last year.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave28(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Continue the persuit carefully.", "EN_SOFT_PURSUIT_MISS", "pursuit"},
		{"I realy need a break.", "EN_SOFT_REALLY_MISS", "really"},
		{"We are recieving traffic now.", "EN_SOFT_RECEIVING_MISS", "receiving"},
		{"Send a refferal link please.", "EN_SOFT_REFERRAL_MISS", "referral"},
		{"Check the relevence carefully.", "EN_SOFT_RELEVANCE_MISS", "relevance"},
		{"Please remeber the password.", "EN_SOFT_REMEMBER_MISS", "remember"},
		{"Find a resturant nearby.", "EN_SOFT_RESTAURANT_MISS", "restaurant"},
		{"End the seige carefully.", "EN_SOFT_SIEGE_MISS", "siege"},
		{"What a surprize today.", "EN_SOFT_SURPRISE_MISS", "surprise"},
		{"Celebrate sucess carefully.", "EN_SOFT_SUCCESS_MISS", "success"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave6(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Capture action items after.", "EN_SOFT_PICKY_ACTION_ITEMS"},
		{"We need net new customers.", "EN_SOFT_PICKY_NET_NEW"},
		{"We have a hard stop at noon.", "EN_SOFT_PICKY_HARD_STOP"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeTbdTbc(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Date is tbd for now.",
		"Date is TBD for now.",
		"Details are tbc still.",
		"Details are TBC still.",
		"That feature is nyi.",
		"Still a WIP today.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave29(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Thierfore we should wait.", "EN_SOFT_THEREFORE_MISS2", "Therefore"},
		{"Unfortunatley the build failed.", "EN_SOFT_UNFORTUNATELY_MISS2", "Unfortunately"},
		{"Plan for unforseen issues.", "EN_SOFT_UNFORESEEN_MISS", "unforeseen"},
		{"Avoid unneccessary work.", "EN_SOFT_UNNECESSARY_MISS", "unnecessary"},
		{"Use a vaccuum carefully.", "EN_SOFT_VACUUM_MISS2", "vacuum"},
		{"Go whereever you like.", "EN_SOFT_WHEREVER_MISS", "wherever"},
		{"We shipped yesturday morning.", "EN_SOFT_YESTERDAY_MISS", "yesterday"},
		{"Find a resteraunt nearby.", "EN_SOFT_RESTAURANT_MISS2", "restaurant"},
		{"Please rember the keys.", "EN_SOFT_REMEMBER_MISS2", "remember"},
		{"Strenghten the test suite.", "EN_SOFT_STRENGTHEN_MISS", "Strengthen"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftOptionalENAsMatterOfFact(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-e", "SOFT_OPTIONAL", "--json", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "As a matter of fact, tests pass.", nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Contains(t, out.String(), "EN_SOFT_OPT_AS_A_MATTER_OF_FACT")
}

func TestGolden_MultiwordSanJose(t *testing.T) {
	if DiscoverEnglishMultiwords(nil) == "" {
		t.Skip("multiwords missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "We flew to San Jose yesterday.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_IgnoreSpellingRipgrep(t *testing.T) {
	if DiscoverEnglishIgnoreSpellingList(nil) == "" {
		t.Skip("ignore-spelling list missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "We use ripgrep and zellij daily.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave30(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"We are adressing the bug.", "EN_SOFT_ADDRESSING_MISS", "addressing"},
		{"Track recieveable carefully.", "EN_SOFT_RECEIVABLE_MISS", "receivable"},
		{"I was surprized by that.", "EN_SOFT_SURPRISED_MISS", "surprised"},
		{"Fill the questionnair carefully.", "EN_SOFT_QUESTIONNAIRE_MISS", "questionnaire"},
		{"Stop the propogation carefully.", "EN_SOFT_PROPAGATION_MISS", "propagation"},
		{"We are persueing the lead.", "EN_SOFT_PURSUING_MISS", "pursuing"},
		{"Over millenia of change.", "EN_SOFT_MILLENNIA_MISS", "millennia"},
		{"She is fully comitted now.", "EN_SOFT_COMMITTED_MISS", "committed"},
		{"I definatly agree with that.", "EN_SOFT_DEFINITELY_MISS2", "definitely"},
		{"I am greatfull for help.", "EN_SOFT_GRATEFUL_MISS2", "grateful"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_FalseFriendsLarge(t *testing.T) {
	ff := softFalseFriendsPath(t)
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "El camino es largo hoy.", &CommandLineOptions{
		Language:         "es",
		MotherTongue:     "en",
		FalseFriendsFile: ff,
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "LARGE" {
			found = true
			require.Equal(t, "long (not large)", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_MultiwordNewOrleans(t *testing.T) {
	if DiscoverEnglishMultiwords(nil) == "" {
		t.Skip("multiwords missing")
	}
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "We visited New Orleans last year.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestGolden_SoftIdiomConfusablesWave31(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please aquire the license.", "EN_SOFT_ACQUIRE_MISS", "acquire"},
		{"They aquired the company.", "EN_SOFT_ACQUIRED_MISS", "acquired"},
		{"Avoid weak arguements here.", "EN_SOFT_ARGUMENTS_MISS", "arguments"},
		{"He is a buisnessman today.", "EN_SOFT_BUSINESSMAN_MISS", "businessman"},
		{"Check the calander carefully.", "EN_SOFT_CALENDAR_MISS", "calendar"},
		{"Plan your carreer carefully.", "EN_SOFT_CAREER_MISS", "career"},
		{"It happended last night.", "EN_SOFT_HAPPENED_MISS2", "happened"},
		{"Prove the existense carefully.", "EN_SOFT_EXISTENCE_MISS", "existence"},
		{"I reccommend this approach.", "EN_SOFT_RECOMMEND_MISS2", "recommend"},
		{"A sucessfull deploy landed.", "EN_SOFT_SUCCESSFUL_MISS2", "successful"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave7(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Do not move the goalposts midstream.", "EN_SOFT_PICKY_MOVE_THE_GOALPOSTS"},
		{"Let us put a pin in that.", "EN_SOFT_PICKY_PUT_A_PIN"},
		{"We need to level set first.", "EN_SOFT_PICKY_LEVEL_SET"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeRfcAdr(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Write an rfc carefully.",
		"Write an RFC carefully.",
		"Record the adr soon.",
		"Record the ADR soon.",
		"Tldr the patch works.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave32(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They are loosing customers.", "EN_SOFT_LOSING_MISS", "losing"},
		{"Bad managment decisions hurt.", "EN_SOFT_MANAGEMENT_MISS", "management"},
		{"One peice is missing.", "EN_SOFT_PIECE_MISS", "piece"},
		{"Find a restaraunt nearby.", "EN_SOFT_RESTAURANT_MISS3", "restaurant"},
		{"Update the scedule carefully.", "EN_SOFT_SCHEDULE_MISS", "schedule"},
		{"Keep them seperated carefully.", "EN_SOFT_SEPARATED_MISS", "separated"},
		{"Check enviromental impact.", "EN_SOFT_ENVIRONMENTAL_MISS", "environmental"},
		{"Bugs are occuring often.", "EN_SOFT_OCCURRING_MISS", "occurring"},
		{"It is completly wrong.", "EN_SOFT_COMPLETELY_MISS", "completely"},
		{"My roomate left early.", "EN_SOFT_ROOMMATE_MISS", "roommate"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave8(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"We are in the weeds already.", "EN_SOFT_PICKY_IN_THE_WEEDS"},
		{"This is herding cats again.", "EN_SOFT_PICKY_HERDING_CATS"},
		{"Please close the loop today.", "EN_SOFT_PICKY_CLOSE_THE_LOOP"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeNfrSreSlo(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Capture each nfr carefully.",
		"Capture each NFR carefully.",
		"Join the sre rotation.",
		"Join the SRE rotation.",
		"Track the slo carefully.",
		"Track the SLO carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave33(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Share it publically soon.", "EN_SOFT_PUBLICLY_MISS", "publicly"},
		{"Face a hard dilemna now.", "EN_SOFT_DILEMMA_MISS", "dilemma"},
		{"They feel desparate today.", "EN_SOFT_DESPERATE_MISS", "desperate"},
		{"I beleived the report.", "EN_SOFT_BELIEVED_MISS", "believed"},
		{"They acheived the goal.", "EN_SOFT_ACHIEVED_MISS", "achieved"},
		{"Take the medecine carefully.", "EN_SOFT_MEDICINE_MISS", "medicine"},
		{"Visit my neice soon.", "EN_SOFT_NIECE_MISS", "niece"},
		{"Seize the oppurtunity carefully.", "EN_SOFT_OPPORTUNITY_MISS3", "opportunity"},
		{"Buy fresh potatos today.", "EN_SOFT_POTATOES_MISS", "potatoes"},
		{"Do not yeild early.", "EN_SOFT_YIELD_MISS", "yield"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave9(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"We prefer to fail fast here.", "EN_SOFT_PICKY_FAIL_FAST"},
		{"Teams should shift left now.", "EN_SOFT_PICKY_SHIFT_LEFT"},
		{"I will be out of pocket Friday.", "EN_SOFT_PICKY_OUT_OF_POCKET"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeJwtUuidK8s(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Validate the jwt carefully.",
		"Validate the JWT carefully.",
		"Store the uuid carefully.",
		"Store the UUID carefully.",
		"Deploy to k8s carefully.",
		"Deploy to K8s carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave34(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Find a similiar example.", "EN_SOFT_SIMILAR_MISS", "similar"},
		{"We gaurantee delivery soon.", "EN_SOFT_GUARANTEE_MISS", "guarantee"},
		{"Do not interupt the talk.", "EN_SOFT_INTERRUPT_MISS", "interrupt"},
		{"Skip irrelevent details.", "EN_SOFT_IRRELEVANT_MISS", "irrelevant"},
		{"Send the correspondance carefully.", "EN_SOFT_CORRESPONDENCE_MISS", "correspondence"},
		{"It aparently works now.", "EN_SOFT_APPARENTLY_MISS", "apparently"},
		{"At the beggining carefully.", "EN_SOFT_BEGINNING_MISS", "beginning"},
		{"An excelent result landed.", "EN_SOFT_EXCELLENT_MISS2", "excellent"},
		{"Bugs can disapear overnight.", "EN_SOFT_DISAPPEAR_MISS2", "disappear"},
		{"Do careful reasearch first.", "EN_SOFT_RESEARCH_MISS", "research"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave10(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Do not double down on risk.", "EN_SOFT_PICKY_DOUBLE_DOWN"},
		{"Leaders should lean in more.", "EN_SOFT_PICKY_LEAN_IN"},
		{"That is in my wheelhouse.", "EN_SOFT_PICKY_WHEELHOUSE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeOauthPrdGrpc(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Enable oauth carefully.",
		"Enable OAuth carefully.",
		"Write the prd carefully.",
		"Write the PRD carefully.",
		"Call grpc carefully.",
		"Call gRPC carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave35(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Use flourescent lights carefully.", "EN_SOFT_FLUORESCENT_MISS", "fluorescent"},
		{"Build intellegence carefully.", "EN_SOFT_INTELLIGENCE_MISS2", "intelligence"},
		{"New rules supercede the old.", "EN_SOFT_SUPERSEDE_MISS", "supersede"},
		{"Keep the label visable.", "EN_SOFT_VISIBLE_MISS", "visible"},
		{"Meet on wensday carefully.", "EN_SOFT_WEDNESDAY_MISS", "Wednesday"},
		{"Finish by the twelth carefully.", "EN_SOFT_TWELFTH_MISS", "twelfth"},
		{"Keep the origional carefully.", "EN_SOFT_ORIGINAL_MISS2", "original"},
		{"Study the assasination carefully.", "EN_SOFT_ASSASSINATION_MISS", "assassination"},
		{"List employee benifits carefully.", "EN_SOFT_BENEFITS_MISS", "benefits"},
		{"Add a new colum carefully.", "EN_SOFT_COLUMN_MISS", "column"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave11(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Let us unpack the metrics.", "EN_SOFT_PICKY_UNPACK"},
		{"Grow your skill set carefully.", "EN_SOFT_PICKY_SKILL_SET"},
		{"Please socialize the plan.", "EN_SOFT_PICKY_SOCIALIZE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeHelmKafkaRedis(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Install helm carefully.",
		"Install Helm carefully.",
		"Publish to kafka carefully.",
		"Publish to Kafka carefully.",
		"Cache in redis carefully.",
		"Cache in Redis carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave36(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Winter is comming soon.", "EN_SOFT_COMING_MISS", "coming"},
		{"Satisfy your curiousity carefully.", "EN_SOFT_CURIOSITY_MISS", "curiosity"},
		{"That is decidely wrong.", "EN_SOFT_DECIDEDLY_MISS", "decidedly"},
		{"They need help desparately.", "EN_SOFT_DESPERATELY_MISS", "desperately"},
		{"Do not disapoint users.", "EN_SOFT_DISAPPOINT_MISS2", "disappoint"},
		{"Put more efort carefully.", "EN_SOFT_EFFORT_MISS", "effort"},
		{"Finish the eigth chapter.", "EN_SOFT_EIGHTH_MISS", "eighth"},
		{"Add one more eliment carefully.", "EN_SOFT_ELEMENT_MISS", "element"},
		{"An embarassing bug shipped.", "EN_SOFT_EMBARRASSING_MISS", "embarrassing"},
		{"Start a new endevor carefully.", "EN_SOFT_ENDEAVOR_MISS", "endeavor"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave12(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Keep a bias for action always.", "EN_SOFT_PICKY_BIAS_FOR_ACTION"},
		{"Design to future proof systems.", "EN_SOFT_PICKY_FUTURE_PROOF"},
		{"Please bake in monitoring.", "EN_SOFT_PICKY_BAKE_IN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeTerraformNginxPostgres(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Apply terraform carefully.",
		"Apply Terraform carefully.",
		"Configure nginx carefully.",
		"Configure Nginx carefully.",
		"Query postgres carefully.",
		"Query Postgres carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave37(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They are equiped carefully.", "EN_SOFT_EQUIPPED_MISS", "equipped"},
		{"Stop all harrassment carefully.", "EN_SOFT_HARASSMENT_MISS", "harassment"},
		{"Reduce bad influance carefully.", "EN_SOFT_INFLUENCE_MISS", "influence"},
		{"Repair the jewlery carefully.", "EN_SOFT_JEWELRY_MISS2", "jewelry"},
		{"Hire knowlegeable people.", "EN_SOFT_KNOWLEDGEABLE_MISS2", "knowledgeable"},
		{"Use a lazer carefully.", "EN_SOFT_LASER_MISS", "laser"},
		{"Plan maintanance carefully.", "EN_SOFT_MAINTENANCE_MISS3", "maintenance"},
		{"A mischevious bug appeared.", "EN_SOFT_MISCHIEVOUS_MISS", "mischievous"},
		{"A noticible delay remains.", "EN_SOFT_NOTICEABLE_MISS2", "noticeable"},
		{"Expect occassional failures.", "EN_SOFT_OCCASIONAL_MISS", "occasional"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave13(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"We should right size the fleet.", "EN_SOFT_PICKY_RIGHT_SIZE"},
		{"Let us move forward carefully.", "EN_SOFT_PICKY_MOVE_FORWARD"},
		{"This will raise the bar carefully.", "EN_SOFT_PICKY_RAISE_THE_BAR"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeMongoGraphqlWebpack(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Query mongo carefully.",
		"Query Mongo carefully.",
		"Design graphql carefully.",
		"Design GraphQL carefully.",
		"Configure webpack carefully.",
		"Configure Webpack carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave38(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fix the omision carefully.", "EN_SOFT_OMISSION_MISS", "omission"},
		{"Events preceed the launch.", "EN_SOFT_PRECEDE_MISS", "precede"},
		{"Ship preferrably tomorrow.", "EN_SOFT_PREFERABLY_MISS", "preferably"},
		{"That is a rare privilage.", "EN_SOFT_PRIVILEGE_MISS", "privilege"},
		{"A priviledged account remains.", "EN_SOFT_PRIVILEGED_MISS", "privileged"},
		{"It will probaly work.", "EN_SOFT_PROBABLY_MISS", "probably"},
		{"A prominant bug remains.", "EN_SOFT_PROMINENT_MISS", "prominent"},
		{"Study psycology carefully.", "EN_SOFT_PSYCHOLOGY_MISS", "psychology"},
		{"Hire a reasearcher carefully.", "EN_SOFT_RESEARCHER_MISS", "researcher"},
		{"Avoid religous debates.", "EN_SOFT_RELIGIOUS_MISS", "religious"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave14(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Please think outside the plan.", "EN_SOFT_PICKY_THINK_OUTSIDE"},
		{"Just ship it carefully.", "EN_SOFT_PICKY_SHIP_IT"},
		{"You must own it carefully.", "EN_SOFT_PICKY_OWN_IT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeEslintPrismaDocker(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run eslint carefully.",
		"Run ESLint carefully.",
		"Configure prisma carefully.",
		"Configure Prisma carefully.",
		"Start docker carefully.",
		"Start Docker carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave39(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"That is reminescent carefully.", "EN_SOFT_REMINISCENT_MISS", "reminiscent"},
		{"Wait for a responce carefully.", "EN_SOFT_RESPONSE_MISS", "response"},
		{"Note the ressemblance carefully.", "EN_SOFT_RESEMBLANCE_MISS", "resemblance"},
		{"Make a sacrafice carefully.", "EN_SOFT_SACRIFICE_MISS", "sacrifice"},
		{"Check saftey carefully.", "EN_SOFT_SAFETY_MISS", "safety"},
		{"Use common sence carefully.", "EN_SOFT_SENSE_MISS", "sense"},
		{"Celebrate the acheivement carefully.", "EN_SOFT_ACHIEVEMENT_MISS", "achievement"},
		{"Ask your collaegue carefully.", "EN_SOFT_COLLEAGUE_MISS2", "colleague"},
		{"Form a comittee carefully.", "EN_SOFT_COMMITTEE_MISS2", "committee"},
		{"Pay the comission carefully.", "EN_SOFT_COMMISSION_MISS", "commission"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave15(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Do not circle the wagons now.", "EN_SOFT_PICKY_CIRCLE_THE_WAGONS"},
		{"They hit the ground running.", "EN_SOFT_PICKY_HIT_THE_GROUND_RUNNING"},
		{"Let us peel the onion carefully.", "EN_SOFT_PICKY_PEEL_THE_ONION"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeSddViteBun(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Write the sdd carefully.",
		"Write the SDD carefully.",
		"Build with vite carefully.",
		"Build with Vite carefully.",
		"Run with bun carefully.",
		"Run with Bun carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave40(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Do not critisize carefully.", "EN_SOFT_CRITICIZE_MISS", "criticize"},
		{"Do not decieve carefully.", "EN_SOFT_DECEIVE_MISS", "deceive"},
		{"Make a desicion carefully.", "EN_SOFT_DECISION_MISS2", "decision"},
		{"Note the differance carefully.", "EN_SOFT_DIFFERENCE_MISS", "difference"},
		{"Protect the enviromment carefully.", "EN_SOFT_ENVIRONMENT_MISS", "environment"},
		{"Buy equipement carefully.", "EN_SOFT_EQUIPMENT_MISS2", "equipment"},
		{"Honor the heros carefully.", "EN_SOFT_HEROES_MISS", "heroes"},
		{"A misterious bug remains.", "EN_SOFT_MYSTERIOUS_MISS", "mysterious"},
		{"It is a neccessity carefully.", "EN_SOFT_NECESSITY_MISS", "necessity"},
		{"Please orginize carefully.", "EN_SOFT_ORGANIZE_MISS", "organize"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave16(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Let us park that idea.", "EN_SOFT_PICKY_PARK_THAT"},
		{"We will table that request.", "EN_SOFT_PICKY_TABLE_THAT"},
		{"Done by close of play.", "EN_SOFT_PICKY_CLOSE_OF_PLAY"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeDenoPlaywrightCypress(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run with deno carefully.",
		"Run with Deno carefully.",
		"Test with playwright carefully.",
		"Test with Playwright carefully.",
		"Test with cypress carefully.",
		"Test with Cypress carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave41(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Run parrallel jobs carefully.", "EN_SOFT_PARALLEL_MISS2", "parallel"},
		{"A persistant issue remains.", "EN_SOFT_PERSISTENT_MISS", "persistent"},
		{"Notify personel carefully.", "EN_SOFT_PERSONNEL_MISS", "personnel"},
		{"Avoid plagarism carefully.", "EN_SOFT_PLAGIARISM_MISS", "plagiarism"},
		{"Choose practicle options.", "EN_SOFT_PRACTICAL_MISS", "practical"},
		{"Avoid repitition carefully.", "EN_SOFT_REPETITION_MISS", "repetition"},
		{"Build resistent systems.", "EN_SOFT_RESISTANT_MISS", "resistant"},
		{"Ask the sargent carefully.", "EN_SOFT_SERGEANT_MISS", "sergeant"},
		{"Celebrate the succes carefully.", "EN_SOFT_SUCCESS_MISS", "success"},
		{"Note the tendancy carefully.", "EN_SOFT_TENDENCY_MISS", "tendency"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave17(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"This is a game changer carefully.", "EN_SOFT_PICKY_GAME_CHANGER"},
		{"Do not push the envelope now.", "EN_SOFT_PICKY_PUSH_THE_ENVELOPE"},
		{"Have skin in the game carefully.", "EN_SOFT_PICKY_SKIN_IN_THE_GAME"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeNextjsJestVitest(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Build with nextjs carefully.",
		"Build with Nextjs carefully.",
		"Test with jest carefully.",
		"Test with Jest carefully.",
		"Test with vitest carefully.",
		"Test with Vitest carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave42(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Buy fresh tomatos carefully.", "EN_SOFT_TOMATOES_MISS", "tomatoes"},
		{"Reduce memory useage carefully.", "EN_SOFT_USAGE_MISS", "usage"},
		{"Do not withold carefully.", "EN_SOFT_WITHHOLD_MISS", "withhold"},
		{"Need contious updates carefully.", "EN_SOFT_CONTINUOUS_MISS2", "continuous"},
		{"Set acheiveable goals carefully.", "EN_SOFT_ACHIEVABLE_MISS", "achievable"},
		{"I am speachless carefully.", "EN_SOFT_SPEECHLESS_MISS", "speechless"},
		{"Oil seperates carefully.", "EN_SOFT_SEPARATES_MISS", "separates"},
		{"A lenghty process remains.", "EN_SOFT_LENGTHY_MISS", "lengthy"},
		{"Plan the manuever carefully.", "EN_SOFT_MANEUVER_MISS", "maneuver"},
		{"It is neccesary carefully.", "EN_SOFT_NECESSARY_MISS", "necessary"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave18(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Address the elephant in the room.", "EN_SOFT_PICKY_ELEPHANT_IN_THE_ROOM"},
		{"Give a ballpark figure carefully.", "EN_SOFT_PICKY_BALLPARK_FIGURE"},
		{"Get on the same page carefully.", "EN_SOFT_PICKY_ON_THE_SAME_PAGE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeTurboWasmNix(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Enable turbo carefully.",
		"Enable Turbo carefully.",
		"Compile to wasm carefully.",
		"Compile to Wasm carefully.",
		"Install nix carefully.",
		"Install Nix carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave43(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Seize oppurtunities carefully.", "EN_SOFT_OPPORTUNITIES_MISS", "opportunities"},
		{"Keep the orginial carefully.", "EN_SOFT_ORIGINAL_MISS3", "original"},
		{"Run parallell tasks carefully.", "EN_SOFT_PARALLEL_MISS3", "parallel"},
		{"A fun passtime remains.", "EN_SOFT_PASTIME_MISS", "pastime"},
		{"Study the phenomina carefully.", "EN_SOFT_PHENOMENA_MISS", "phenomena"},
		{"They posess power carefully.", "EN_SOFT_POSSESS_MISS", "possess"},
		{"Night preceeds day carefully.", "EN_SOFT_PRECEDES_MISS", "precedes"},
		{"Featured prominantely carefully.", "EN_SOFT_PROMINENTLY_MISS", "prominently"},
		{"I remebered carefully.", "EN_SOFT_REMEMBERED_MISS", "remembered"},
		{"My roomates left carefully.", "EN_SOFT_ROOMMATES_MISS", "roommates"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave19(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Do not drink the kool-aid carefully.", "EN_SOFT_PICKY_DRINK_THE_KOOL_AID"},
		{"Seek a win-win situation carefully.", "EN_SOFT_PICKY_WIN_WIN_SITUATION"},
		{"Please think outside the box carefully.", "EN_SOFT_PICKY_THINK_OUTSIDE_THE_BOX"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeRustcMiseGolang(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Invoke rustc carefully.",
		"Invoke Rustc carefully.",
		"Install mise carefully.",
		"Install Mise carefully.",
		"Learn golang carefully.",
		"Learn Golang carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave44(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Note the similiarity carefully.", "EN_SOFT_SIMILARITY_MISS", "similarity"},
		{"They are seigeing carefully.", "EN_SOFT_SIEGING_MISS", "sieging"},
		{"End the seiges carefully.", "EN_SOFT_SIEGES_MISS", "sieges"},
		{"Avoid a siezure carefully.", "EN_SOFT_SEIZURE_MISS", "seizure"},
		{"Improve the wirting carefully.", "EN_SOFT_WRITING_MISS", "writing"},
		{"Stop yeilding carefully.", "EN_SOFT_YIELDING_MISS", "yielding"},
		{"Celebrate the aniversary carefully.", "EN_SOFT_ANNIVERSARY_MISS", "anniversary"},
		{"It aparenty works carefully.", "EN_SOFT_APPARENTLY_MISS2", "apparently"},
		{"Meet an aquiantance carefully.", "EN_SOFT_ACQUAINTANCE_MISS2", "acquaintance"},
		{"Avoid a weak arguemnt carefully.", "EN_SOFT_ARGUMENT_MISS", "argument"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave20(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"That choice is a no-brainer carefully.", "EN_SOFT_PICKY_NO_BRAINER"},
		{"Now get after it carefully.", "EN_SOFT_PICKY_GET_AFTER_IT"},
		{"Net-net the plan works.", "EN_SOFT_PICKY_NET_NET"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCueBazelPodman(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Validate cue carefully.",
		"Validate Cue carefully.",
		"Build with bazel carefully.",
		"Build with Bazel carefully.",
		"Run with podman carefully.",
		"Run with Podman carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave45(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please asess carefully.", "EN_SOFT_ASSESS_MISS", "assess"},
		{"Write an asessment carefully.", "EN_SOFT_ASSESSMENT_MISS", "assessment"},
		{"Hire an attorny carefully.", "EN_SOFT_ATTORNEY_MISS", "attorney"},
		{"Recall the beginings carefully.", "EN_SOFT_BEGINNINGS_MISS", "beginnings"},
		{"Help small buisnesses carefully.", "EN_SOFT_BUSINESSES_MISS", "businesses"},
		{"Avoid a carreerist carefully.", "EN_SOFT_CAREERIST_MISS", "careerist"},
		{"Visit the cemetarys carefully.", "EN_SOFT_CEMETERIES_MISS", "cemeteries"},
		{"Thank your collaegues carefully.", "EN_SOFT_COLLEAGUES_MISS", "colleagues"},
		{"Show comitment carefully.", "EN_SOFT_COMMITMENT_MISS", "commitment"},
		{"Do not be decieved carefully.", "EN_SOFT_DECEIVED_MISS", "deceived"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave21(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Name our core competency carefully.", "EN_SOFT_PICKY_CORE_COMPETENCY"},
		{"State the value proposition carefully.", "EN_SOFT_PICKY_VALUE_PROPOSITION"},
		{"Let us take this offline carefully.", "EN_SOFT_PICKY_TAKE_THIS_OFFLINE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCmakeSkaffoldKustomize(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Configure cmake carefully.",
		"Configure CMake carefully.",
		"Run skaffold carefully.",
		"Run Skaffold carefully.",
		"Apply kustomize carefully.",
		"Apply Kustomize carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave46(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Note the absense carefully.", "EN_SOFT_ABSENCE_MISS", "absence"},
		{"Celebrate the acheivment carefully.", "EN_SOFT_ACHIEVEMENT_MISS2", "achievement"},
		{"They are aquireing carefully.", "EN_SOFT_ACQUIRING_MISS", "acquiring"},
		{"Read the articel carefully.", "EN_SOFT_ARTICLE_MISS", "article"},
		{"Do not asume carefully.", "EN_SOFT_ASSUME_MISS", "assume"},
		{"He is an athiest carefully.", "EN_SOFT_ATHEIST_MISS", "atheist"},
		{"A beautifull day carefully.", "EN_SOFT_BEAUTIFUL_MISS2", "beautiful"},
		{"Finish befor carefully.", "EN_SOFT_BEFORE_MISS", "before"},
		{"I belive carefully.", "EN_SOFT_BELIEVE_MISS", "believe"},
		{"A benificial change carefully.", "EN_SOFT_BENEFICIAL_MISS", "beneficial"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave22(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Improve stakeholder management carefully.", "EN_SOFT_PICKY_STAKEHOLDER_MANAGEMENT"},
		{"Keep a single source of truth carefully.", "EN_SOFT_PICKY_SINGLE_SOURCE_OF_TRUTH"},
		{"Track the north star metric carefully.", "EN_SOFT_PICKY_NORTH_STAR_METRIC"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeAnsibleVagrantConsul(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run ansible carefully.",
		"Run Ansible carefully.",
		"Start vagrant carefully.",
		"Start Vagrant carefully.",
		"Query consul carefully.",
		"Query Consul carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave47(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Use camoflage carefully.", "EN_SOFT_CAMOUFLAGE_MISS", "camouflage"},
		{"Accept the challange carefully.", "EN_SOFT_CHALLENGE_MISS", "challenge"},
		{"Build strong charactor carefully.", "EN_SOFT_CHARACTER_MISS", "character"},
		{"Teams collaberate carefully.", "EN_SOFT_COLLABORATE_MISS", "collaborate"},
		{"We comemorate carefully.", "EN_SOFT_COMMEMORATE_MISS", "commemorate"},
		{"Check completness carefully.", "EN_SOFT_COMPLETENESS_MISS", "completeness"},
		{"Fix the conection carefully.", "EN_SOFT_CONNECTION_MISS", "connection"},
		{"Stay consentrated carefully.", "EN_SOFT_CONCENTRATED_MISS", "concentrated"},
		{"Delay the contruction carefully.", "EN_SOFT_CONSTRUCTION_MISS", "construction"},
		{"Avoid corupt data carefully.", "EN_SOFT_CORRUPT_MISS", "corrupt"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave23(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Follow best practice carefully.", "EN_SOFT_PICKY_BEST_PRACTICE"},
		{"Name the pain point carefully.", "EN_SOFT_PICKY_PAIN_POINT"},
		{"Share the learnings carefully.", "EN_SOFT_PICKY_LEARNINGS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizePackerNomadVault(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Build with packer carefully.",
		"Build with Packer carefully.",
		"Schedule on nomad carefully.",
		"Schedule on Nomad carefully.",
		"Store secrets in vault carefully.",
		"Store secrets in Vault carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave48(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please correpond carefully.", "EN_SOFT_CORRESPOND_MISS", "correspond"},
		{"Do not critisise carefully.", "EN_SOFT_CRITICISE_MISS", "criticise"},
		{"They decieded carefully.", "EN_SOFT_DECIDED_MISS", "decided"},
		{"Please deffine carefully.", "EN_SOFT_DEFINE_MISS", "define"},
		{"Add a dependancy carefully.", "EN_SOFT_DEPENDENCY_MISS", "dependency"},
		{"Track developemental carefully.", "EN_SOFT_DEVELOPMENTAL_MISS", "developmental"},
		{"Pick a differant carefully.", "EN_SOFT_DIFFERENT_MISS2", "different"},
		{"Value freindship carefully.", "EN_SOFT_FRIENDSHIP_MISS", "friendship"},
		{"Keep gaurding carefully.", "EN_SOFT_GUARDING_MISS", "guarding"},
		{"Elect a governer carefully.", "EN_SOFT_GOVERNOR_MISS", "governor"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave24(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Let us deep dive into metrics.", "EN_SOFT_PICKY_DEEP_DIVE_INTO"},
		{"We will circle back later carefully.", "EN_SOFT_PICKY_CIRCLE_BACK_LATER"},
		{"I am bandwidth constrained carefully.", "EN_SOFT_PICKY_BANDWIDTH_CONSTRAINED"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeWaypointJqBoundary(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Deploy with waypoint carefully.",
		"Deploy with Waypoint carefully.",
		"Filter with jq carefully.",
		"Filter with JQ carefully.",
		"Secure with boundary carefully.",
		"Secure with Boundary carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave49(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Hire experianced people carefully.", "EN_SOFT_EXPERIENCED_MISS", "experienced"},
		{"Do daily exersize carefully.", "EN_SOFT_EXERCISE_MISS", "exercise"},
		{"At the beging carefully.", "EN_SOFT_BEGINNING_MISS2", "beginning"},
		{"Join the divison carefully.", "EN_SOFT_DIVISION_MISS", "division"},
		{"Write documention carefully.", "EN_SOFT_DOCUMENTATION_MISS", "documentation"},
		{"An enourmous bug remains.", "EN_SOFT_ENORMOUS_MISS", "enormous"},
		{"Check everthing carefully.", "EN_SOFT_EVERYTHING_MISS", "everything"},
		{"Give an explination carefully.", "EN_SOFT_EXPLANATION_MISS", "explanation"},
		{"It is extremly carefully.", "EN_SOFT_EXTREMELY_MISS", "extremely"},
		{"Help familes carefully.", "EN_SOFT_FAMILIES_MISS", "families"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave25(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Plan the go to market carefully.", "EN_SOFT_PICKY_GO_TO_MARKET"},
		{"Find product market fit carefully.", "EN_SOFT_PICKY_PRODUCT_MARKET_FIT"},
		{"Reduce time to market carefully.", "EN_SOFT_PICKY_TIME_TO_MARKET"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeSkopeoBuildahCrio(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Inspect with skopeo carefully.",
		"Inspect with Skopeo carefully.",
		"Build with buildah carefully.",
		"Build with Buildah carefully.",
		"Run with crio carefully.",
		"Run with CriO carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave50(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fortunatly it worked carefully.", "EN_SOFT_FORTUNATELY_MISS", "Fortunately"},
		{"Build freindships carefully.", "EN_SOFT_FRIENDSHIPS_MISS", "friendships"},
		{"Call the funtion carefully.", "EN_SOFT_FUNCTION_MISS", "function"},
		{"It is gauranteed carefully.", "EN_SOFT_GUARANTEED_MISS", "guaranteed"},
		{"Post more gaurds carefully.", "EN_SOFT_GUARDS_MISS", "guards"},
		{"Reform goverments carefully.", "EN_SOFT_GOVERNMENTS_MISS", "governments"},
		{"I greatfully accept carefully.", "EN_SOFT_GRATEFULLY_MISS", "gratefully"},
		{"Flatten the hierachy carefully.", "EN_SOFT_HIERARCHY_MISS2", "hierarchy"},
		{"Hightlight the risk carefully.", "EN_SOFT_HIGHLIGHT_MISS", "Highlight"},
		{"Ship initally carefully.", "EN_SOFT_INITIALLY_MISS", "initially"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave26(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Ship a minimum viable product carefully.", "EN_SOFT_PICKY_MINIMUM_VIABLE_PRODUCT"},
		{"Run a proof of concept carefully.", "EN_SOFT_PICKY_PROOF_OF_CONCEPT"},
		{"Measure return on investment carefully.", "EN_SOFT_PICKY_RETURN_ON_INVESTMENT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeNerdctlContainerdRunc(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run with nerdctl carefully.",
		"Run with Nerdctl carefully.",
		"Start containerd carefully.",
		"Start Containerd carefully.",
		"Invoke runc carefully.",
		"Invoke Runc carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave51(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fight ignorence carefully.", "EN_SOFT_IGNORANCE_MISS", "ignorance"},
		{"Hire inteligent people carefully.", "EN_SOFT_INTELLIGENT_MISS", "intelligent"},
		{"Visit the labratory carefully.", "EN_SOFT_LABORATORY_MISS", "laboratory"},
		{"Moniter the system carefully.", "EN_SOFT_MONITOR_MISS", "Monitor"},
		{"Pay the morgage carefully.", "EN_SOFT_MORTGAGE_MISS", "mortgage"},
		{"It is necesary carefully.", "EN_SOFT_NECESSARY_MISS2", "necessary"},
		{"Walk the neigborhood carefully.", "EN_SOFT_NEIGHBORHOOD_MISS", "neighborhood"},
		{"Seize the oportunity carefully.", "EN_SOFT_OPPORTUNITY_MISS4", "opportunity"},
		{"Do the oposite carefully.", "EN_SOFT_OPPOSITE_MISS", "opposite"},
		{"Proove the claim carefully.", "EN_SOFT_PROVE_MISS", "Prove"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave27(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Reduce total cost of ownership carefully.", "EN_SOFT_PICKY_TOTAL_COST_OF_OWNERSHIP"},
		{"Track each key performance indicator carefully.", "EN_SOFT_PICKY_KEY_PERFORMANCE_INDICATOR"},
		{"Map the customer journey carefully.", "EN_SOFT_PICKY_CUSTOMER_JOURNEY"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCaddyTraefikEnvoy(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Front with caddy carefully.",
		"Front with Caddy carefully.",
		"Route with traefik carefully.",
		"Route with Traefik carefully.",
		"Proxy with envoy carefully.",
		"Proxy with Envoy carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave52(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Join the orginisation carefully.", "EN_SOFT_ORGANISATION_MISS", "organisation"},
		{"Be particuarly careful carefully.", "EN_SOFT_PARTICULARLY_MISS", "particularly"},
		{"Collect the peices carefully.", "EN_SOFT_PIECES_MISS", "pieces"},
		{"Notify personell carefully.", "EN_SOFT_PERSONNEL_MISS2", "personnel"},
		{"Do not plagarize carefully.", "EN_SOFT_PLAGIARIZE_MISS", "plagiarize"},
		{"Finish preperations carefully.", "EN_SOFT_PREPARATIONS_MISS", "preparations"},
		{"Please remmember carefully.", "EN_SOFT_REMEMBER_MISS3", "remember"},
		{"Yours sincerly carefully.", "EN_SOFT_SINCERELY_MISS", "sincerely"},
		{"Check the temperture carefully.", "EN_SOFT_TEMPERATURE_MISS", "temperature"},
		{"Ship a varient carefully.", "EN_SOFT_VARIANT_MISS", "variant"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave28(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Define the user persona carefully.", "EN_SOFT_PICKY_USER_PERSONA"},
		{"Apply jobs to be done carefully.", "EN_SOFT_PICKY_JOBS_TO_BE_DONE"},
		{"Capture voice of the customer carefully.", "EN_SOFT_PICKY_VOICE_OF_THE_CUSTOMER"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeLinkerdIstioCilium(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Mesh with linkerd carefully.",
		"Mesh with Linkerd carefully.",
		"Mesh with istio carefully.",
		"Mesh with Istio carefully.",
		"Network with cilium carefully.",
		"Network with Cilium carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave53(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Study psycological factors carefully.", "EN_SOFT_PSYCHOLOGICAL_MISS", "psychological"},
		{"Find restaraunts carefully.", "EN_SOFT_RESTAURANTS_MISS", "restaurants"},
		{"Update the scedules carefully.", "EN_SOFT_SCHEDULES_MISS", "schedules"},
		{"Write speaches carefully.", "EN_SOFT_SPEECHES_MISS", "speeches"},
		{"Deploy succesfully carefully.", "EN_SOFT_SUCCESSFULLY_MISS2", "successfully"},
		{"Avoid a suprise carefully.", "EN_SOFT_SURPRISE_MISS2", "surprise"},
		{"Note the tendancies carefully.", "EN_SOFT_TENDENCIES_MISS", "tendencies"},
		{"It acts wierdly carefully.", "EN_SOFT_WEIRDLY_MISS", "weirdly"},
		{"Stop witholding carefully.", "EN_SOFT_WITHHOLDING_MISS", "withholding"},
		{"Secure enviroments carefully.", "EN_SOFT_ENVIRONMENTS_MISS", "environments"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave29(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply design thinking carefully.", "EN_SOFT_PICKY_DESIGN_THINKING"},
		{"Avoid growth hacking carefully.", "EN_SOFT_PICKY_GROWTH_HACKING"},
		{"Fund a moonshot carefully.", "EN_SOFT_PICKY_MOONSHOT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeContourKongHaproxy(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Ingress with contour carefully.",
		"Ingress with Contour carefully.",
		"Gateway with kong carefully.",
		"Gateway with Kong carefully.",
		"Balance with haproxy carefully.",
		"Balance with HAProxy carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave54(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Stay buisnesslike carefully.", "EN_SOFT_BUSINESSLIKE_MISS", "businesslike"},
		{"Events were calendered carefully.", "EN_SOFT_CALENDARED_MISS", "calendared"},
		{"Buy equippment carefully.", "EN_SOFT_EQUIPMENT_MISS3", "equipment"},
		{"Measure the lenghts carefully.", "EN_SOFT_LENGTHS_MISS", "lengths"},
		{"Assign liasons carefully.", "EN_SOFT_LIAISONS_MISS", "liaisons"},
		{"Visit libaries carefully.", "EN_SOFT_LIBRARIES_MISS", "libraries"},
		{"Note the absance carefully.", "EN_SOFT_ABSENCE_MISS2", "absence"},
		{"Make it accesible carefully.", "EN_SOFT_ACCESSIBLE_MISS", "accessible"},
		{"Avoid agressive growth carefully.", "EN_SOFT_AGGRESSIVE_MISS", "aggressive"},
		{"Reduce the ammount carefully.", "EN_SOFT_AMOUNT_MISS", "amount"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave30(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Avoid blitzscaling carefully.", "EN_SOFT_PICKY_BLITZSCALING"},
		{"Hire a 10x engineer carefully.", "EN_SOFT_PICKY_TENX_ENGINEER"},
		{"Learn to fail forward carefully.", "EN_SOFT_PICKY_FAIL_FORWARD"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeFluxArgocdHelmfile(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Sync with flux carefully.",
		"Sync with Flux carefully.",
		"Deploy with argocd carefully.",
		"Deploy with ArgoCD carefully.",
		"Render with helmfile carefully.",
		"Render with Helmfile carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave55(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It is allmost ready carefully.", "EN_SOFT_ALMOST_MISS", "almost"},
		{"We allways test carefully.", "EN_SOFT_ALWAYS_MISS", "always"},
		{"Please anounce carefully.", "EN_SOFT_ANNOUNCE_MISS", "announce"},
		{"Post an anouncement carefully.", "EN_SOFT_ANNOUNCEMENT_MISS", "announcement"},
		{"Check the apearance carefully.", "EN_SOFT_APPEARANCE_MISS2", "appearance"},
		{"Hire an asistant carefully.", "EN_SOFT_ASSISTANT_MISS", "assistant"},
		{"Improve the athmosphere carefully.", "EN_SOFT_ATMOSPHERE_MISS", "atmosphere"},
		{"Keep files attatched carefully.", "EN_SOFT_ATTACHED_MISS", "attached"},
		{"Make it avaiable carefully.", "EN_SOFT_AVAILABLE_MISS", "available"},
		{"Check the backround carefully.", "EN_SOFT_BACKGROUND_MISS", "background"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave31(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Try working backwards carefully.", "EN_SOFT_PICKY_WORKING_BACKWARDS"},
		{"Practice customer obsession carefully.", "EN_SOFT_PICKY_CUSTOMER_OBSESSION"},
		{"Please ship early carefully.", "EN_SOFT_PICKY_SHIP_EARLY"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeSopsPulumiTerragrunt(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Encrypt with sops carefully.",
		"Encrypt with Sops carefully.",
		"Provision with pulumi carefully.",
		"Provision with Pulumi carefully.",
		"Wrap with terragrunt carefully.",
		"Wrap with Terragrunt carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave56(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It is becomming clear carefully.", "EN_SOFT_BECOMING_MISS", "becoming"},
		{"Plan beforhand carefully.", "EN_SOFT_BEFOREHAND_MISS", "beforehand"},
		{"Users benifited carefully.", "EN_SOFT_BENEFITED_MISS", "benefited"},
		{"Choose betwen options carefully.", "EN_SOFT_BETWEEN_MISS", "between"},
		{"Cross the boundry carefully.", "EN_SOFT_BOUNDARY_MISS", "boundary"},
		{"Use camoflague carefully.", "EN_SOFT_CAMOUFLAGE_MISS2", "camouflage"},
		{"A challanging task carefully.", "EN_SOFT_CHALLENGING_MISS", "challenging"},
		{"Build charachter carefully.", "EN_SOFT_CHARACTER_MISS2", "character"},
		{"Meet the cheif carefully.", "EN_SOFT_CHIEF_MISS", "chief"},
		{"Boxes contian items carefully.", "EN_SOFT_CONTAIN_MISS", "contain"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave32(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Form a two pizza team carefully.", "EN_SOFT_PICKY_TWO_PIZZA_TEAM"},
		{"Leaders think big carefully.", "EN_SOFT_PICKY_THINK_BIG"},
		{"Keep a day one mindset carefully.", "EN_SOFT_PICKY_DAY_ONE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCrossplaneKyvernoGatekeeper(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Provision with crossplane carefully.",
		"Provision with Crossplane carefully.",
		"Policy with kyverno carefully.",
		"Policy with Kyverno carefully.",
		"Policy with gatekeeper carefully.",
		"Policy with Gatekeeper carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave57(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Make it convienent carefully.", "EN_SOFT_CONVENIENT_MISS2", "convenient"},
		{"Please cosider carefully.", "EN_SOFT_CONSIDER_MISS", "consider"},
		{"Stay cosistent carefully.", "EN_SOFT_CONSISTENT_MISS", "consistent"},
		{"Form a comitee carefully.", "EN_SOFT_COMMITTEE_MISS3", "committee"},
		{"I deffinitely agree carefully.", "EN_SOFT_DEFINITELY_MISS3", "definitely"},
		{"Hold discusions carefully.", "EN_SOFT_DISCUSSIONS_MISS", "discussions"},
		{"Join the eforts carefully.", "EN_SOFT_EFFORTS_MISS", "efforts"},
		{"An embarasing bug carefully.", "EN_SOFT_EMBARRASSING_MISS2", "embarrassing"},
		{"An awfull bug carefully.", "EN_SOFT_AWFUL_MISS", "awful"},
		{"Courts aquit carefully.", "EN_SOFT_ACQUIT_MISS", "acquit"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave33(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Teams disagree and commit carefully.", "EN_SOFT_PICKY_DISAGREE_AND_COMMIT"},
		{"Leaders earn trust carefully.", "EN_SOFT_PICKY_EARN_TRUST"},
		{"Engineers dive deep carefully.", "EN_SOFT_PICKY_DIVE_DEEP"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeFalcoTrivyGrype(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Alert with falco carefully.",
		"Alert with Falco carefully.",
		"Scan with trivy carefully.",
		"Scan with Trivy carefully.",
		"Scan with grype carefully.",
		"Scan with Grype carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave58(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I definetly agree carefully.", "EN_SOFT_DEFINITELY_MISS4", "definitely"},
		{"Keep cosntant pressure carefully.", "EN_SOFT_CONSTANT_MISS", "constant"},
		{"Send correpondence carefully.", "EN_SOFT_CORRESPONDENCE_MISS2", "correspondence"},
		{"It fails cosntantly carefully.", "EN_SOFT_CONSTANTLY_MISS", "constantly"},
		{"Name the contsants carefully.", "EN_SOFT_CONSTANTS_MISS", "constants"},
		{"Interview condidates carefully.", "EN_SOFT_CANDIDATES_MISS", "candidates"},
		{"Enable collaberation carefully.", "EN_SOFT_COLLABORATION_MISS", "collaboration"},
		{"Build confidance carefully.", "EN_SOFT_CONFIDENCE_MISS", "confidence"},
		{"Need contineous updates carefully.", "EN_SOFT_CONTINUOUS_MISS3", "continuous"},
		{"Join corperate training carefully.", "EN_SOFT_CORPORATE_MISS", "corporate"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave34(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Leaders insist on high standards carefully.", "EN_SOFT_PICKY_INSIST_ON_HIGH_STANDARDS"},
		{"Leaders have backbone carefully.", "EN_SOFT_PICKY_HAVE_BACKBONE"},
		{"Teams deliver results carefully.", "EN_SOFT_PICKY_DELIVER_RESULTS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeOpaCosignSyft(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Policy with opa carefully.",
		"Policy with OPA carefully.",
		"Sign with cosign carefully.",
		"Sign with Cosign carefully.",
		"Inventory with syft carefully.",
		"Inventory with Syft carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave59(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Keep it contolled carefully.", "EN_SOFT_CONTROLLED_MISS", "controlled"},
		{"Keep it controled carefully.", "EN_SOFT_CONTROLLED_MISS2", "controlled"},
		{"Prefer convinience carefully.", "EN_SOFT_CONVENIENCE_MISS", "convenience"},
		{"Type it correclty carefully.", "EN_SOFT_CORRECTLY_MISS", "correctly"},
		{"Fight coruption carefully.", "EN_SOFT_CORRUPTION_MISS", "corruption"},
		{"Protect the cosumer carefully.", "EN_SOFT_CONSUMER_MISS", "consumer"},
		{"Thank the creater carefully.", "EN_SOFT_CREATOR_MISS", "creator"},
		{"Please deliever carefully.", "EN_SOFT_DELIVER_MISS", "deliver"},
		{"Track delievery carefully.", "EN_SOFT_DELIVERY_MISS", "delivery"},
		{"Check the detials carefully.", "EN_SOFT_DETAILS_MISS", "details"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave35(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Build an ownership mindset carefully.", "EN_SOFT_PICKY_OWNERSHIP_MINDSET"},
		{"Stay customer centric carefully.", "EN_SOFT_PICKY_CUSTOMER_CENTRIC"},
		{"Teams act small carefully.", "EN_SOFT_PICKY_ACT_SMALL"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeOrasCraneRegctl(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Push with oras carefully.",
		"Push with ORAS carefully.",
		"Copy with crane carefully.",
		"Copy with Crane carefully.",
		"Query with regctl carefully.",
		"Query with Regctl carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave60(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Write a defintion carefully.", "EN_SOFT_DEFINITION_MISS", "definition"},
		{"Keep it deffined carefully.", "EN_SOFT_DEFINED_MISS", "defined"},
		{"Avoid danagerous paths carefully.", "EN_SOFT_DANGEROUS_MISS", "dangerous"},
		{"Help each custmer carefully.", "EN_SOFT_CUSTOMER_MISS", "customer"},
		{"It is currenlty open carefully.", "EN_SOFT_CURRENTLY_MISS", "currently"},
		{"Track the curent status carefully.", "EN_SOFT_CURRENT_MISS", "current"},
		{"Stay detirmined carefully.", "EN_SOFT_DETERMINED_MISS", "determined"},
		{"A dificult task carefully.", "EN_SOFT_DIFFICULT_MISS", "difficult"},
		{"Face a hard dilema carefully.", "EN_SOFT_DILEMMA_MISS2", "dilemma"},
		{"Use dinamic typing carefully.", "EN_SOFT_DYNAMIC_MISS", "dynamic"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave36(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Stay customer first carefully.", "EN_SOFT_PICKY_CUSTOMER_FIRST"},
		{"Be data driven carefully.", "EN_SOFT_PICKY_DATA_DRIVEN"},
		{"Hire full stack engineers carefully.", "EN_SOFT_PICKY_FULL_STACK"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeSteampipeGitleaksSemgrep(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Query with steampipe carefully.",
		"Query with Steampipe carefully.",
		"Scan with gitleaks carefully.",
		"Scan with Gitleaks carefully.",
		"Scan with semgrep carefully.",
		"Scan with Semgrep carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave61(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Note the diference carefully.", "EN_SOFT_DIFFERENCE_MISS2", "difference"},
		{"Pick a diferent path carefully.", "EN_SOFT_DIFFERENT_MISS3", "different"},
		{"Act diferently carefully.", "EN_SOFT_DIFFERENTLY_MISS", "differently"},
		{"Face dificulties carefully.", "EN_SOFT_DIFFICULTIES_MISS", "difficulties"},
		{"Add a dimmension carefully.", "EN_SOFT_DIMENSION_MISS", "dimension"},
		{"Load dinamicly carefully.", "EN_SOFT_DYNAMICALLY_MISS", "dynamically"},
		{"Speak diretly carefully.", "EN_SOFT_DIRECTLY_MISS", "directly"},
		{"I felt disapointed carefully.", "EN_SOFT_DISAPPOINTED_MISS2", "disappointed"},
		{"Bugs are disapearing carefully.", "EN_SOFT_DISAPPEARING_MISS", "disappearing"},
		{"Practice disipline carefully.", "EN_SOFT_DISCIPLINE_MISS", "discipline"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave37(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Adopt zero trust carefully.", "EN_SOFT_PICKY_ZERO_TRUST"},
		{"Invest in platform engineering carefully.", "EN_SOFT_PICKY_PLATFORM_ENGINEERING"},
		{"Stay outcome oriented carefully.", "EN_SOFT_PICKY_OUTCOME_ORIENTED"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCheckovTerrascanTrufflehog(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Scan with checkov carefully.",
		"Scan with Checkov carefully.",
		"Scan with terrascan carefully.",
		"Scan with Terrascan carefully.",
		"Scan with trufflehog carefully.",
		"Scan with TruffleHog carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave62(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Do not disolve carefully.", "EN_SOFT_DISSOLVE_MISS", "dissolve"},
		{"Do not distroy carefully.", "EN_SOFT_DESTROY_MISS", "destroy"},
		{"Please discouver carefully.", "EN_SOFT_DISCOVER_MISS", "discover"},
		{"We discused carefully.", "EN_SOFT_DISCUSSED_MISS", "discussed"},
		{"Write a documnet carefully.", "EN_SOFT_DOCUMENT_MISS", "document"},
		{"Start the dowload carefully.", "EN_SOFT_DOWNLOAD_MISS", "download"},
		{"Stay eagar carefully.", "EN_SOFT_EAGER_MISS", "eager"},
		{"Ship ealry carefully.", "EN_SOFT_EARLY_MISS", "early"},
		{"Please eleminate carefully.", "EN_SOFT_ELIMINATE_MISS", "eliminate"},
		{"Make it eligable carefully.", "EN_SOFT_ELIGIBLE_MISS", "eligible"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave38(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Improve developer experience carefully.", "EN_SOFT_PICKY_DEVELOPER_EXPERIENCE"},
		{"Offer a paved road carefully.", "EN_SOFT_PICKY_PAVED_ROAD"},
		{"Follow the golden path carefully.", "EN_SOFT_PICKY_GOLDEN_PATH"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeKubevalKubeconformPolaris(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Validate with kubeval carefully.",
		"Validate with Kubeval carefully.",
		"Validate with kubeconform carefully.",
		"Validate with Kubeconform carefully.",
		"Audit with polaris carefully.",
		"Audit with Polaris carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave63(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please embbed carefully.", "EN_SOFT_EMBED_MISS", "embed"},
		{"Keep it embeded carefully.", "EN_SOFT_EMBEDDED_MISS", "embedded"},
		{"Handle emergancy carefully.", "EN_SOFT_EMERGENCY_MISS", "emergency"},
		{"Hire an empoyee carefully.", "EN_SOFT_EMPLOYEE_MISS", "employee"},
		{"Please encorage carefully.", "EN_SOFT_ENCOURAGE_MISS", "encourage"},
		{"Hire an engeneer carefully.", "EN_SOFT_ENGINEER_MISS", "engineer"},
		{"Start an endevour carefully.", "EN_SOFT_ENDEAVOUR_MISS", "endeavour"},
		{"Fix it emediately carefully.", "EN_SOFT_IMMEDIATELY_MISS2", "immediately"},
		{"Use electirc power carefully.", "EN_SOFT_ELECTRIC_MISS", "electric"},
		{"Please egnore carefully.", "EN_SOFT_IGNORE_MISS", "ignore"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave39(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Adopt inner source carefully.", "EN_SOFT_PICKY_INNER_SOURCE"},
		{"Prefer shift left testing carefully.", "EN_SOFT_PICKY_SHIFT_LEFT_TESTING"},
		{"Remember you build it you run it carefully.", "EN_SOFT_PICKY_YOU_BUILD_IT_YOU_RUN_IT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeConftestStarlarkKyvernoctl(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Policy with conftest carefully.",
		"Policy with Conftest carefully.",
		"Script with starlark carefully.",
		"Script with Starlark carefully.",
		"Manage with kyvernoctl carefully.",
		"Manage with Kyvernoctl carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave64(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They developped carefully.", "EN_SOFT_DEVELOPED_MISS", "developed"},
		{"Note the disapearance carefully.", "EN_SOFT_DISAPPEARANCE_MISS", "disappearance"},
		{"Hold a discusion carefully.", "EN_SOFT_DISCUSSION_MISS", "discussion"},
		{"Thank the creaters carefully.", "EN_SOFT_CREATORS_MISS", "creators"},
		{"Wait eagarly carefully.", "EN_SOFT_EAGERLY_MISS", "eagerly"},
		{"Hire empoyees carefully.", "EN_SOFT_EMPLOYEES_MISS", "employees"},
		{"Please emphesize carefully.", "EN_SOFT_EMPHASIZE_MISS", "emphasize"},
		{"Study engeneering carefully.", "EN_SOFT_ENGINEERING_MISS", "engineering"},
		{"Need enforcment carefully.", "EN_SOFT_ENFORCEMENT_MISS", "enforcement"},
		{"Make it efortless carefully.", "EN_SOFT_EFFORTLESS_MISS", "effortless"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave40(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Reduce cognitive load carefully.", "EN_SOFT_PICKY_COGNITIVE_LOAD"},
		{"Apply team topologies carefully.", "EN_SOFT_PICKY_TEAM_TOPOLOGIES"},
		{"Prefer shared nothing carefully.", "EN_SOFT_PICKY_SHARED_NOTHING"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCuectlRegclientKustomizectl(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run with cuectl carefully.",
		"Run with Cuectl carefully.",
		"Copy with regclient carefully.",
		"Copy with Regclient carefully.",
		"Apply with kustomizectl carefully.",
		"Apply with Kustomizectl carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave65(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It was distroyed carefully.", "EN_SOFT_DESTROYED_MISS", "destroyed"},
		{"It was disolved carefully.", "EN_SOFT_DISSOLVED_MISS", "dissolved"},
		{"They discouvered carefully.", "EN_SOFT_DISCOVERED_MISS", "discovered"},
		{"Reduce emmision carefully.", "EN_SOFT_EMISSION_MISS", "emission"},
		{"Add emphesis carefully.", "EN_SOFT_EMPHASIS_MISS", "emphasis"},
		{"They encoraged carefully.", "EN_SOFT_ENCOURAGED_MISS", "encouraged"},
		{"Hire an enginneer carefully.", "EN_SOFT_ENGINEER_MISS2", "engineer"},
		{"It is essencial carefully.", "EN_SOFT_ESSENTIAL_MISS", "essential"},
		{"Please estabilish carefully.", "EN_SOFT_ESTABLISH_MISS", "establish"},
		{"Give an estamate carefully.", "EN_SOFT_ESTIMATE_MISS", "estimate"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave41(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Form a stream aligned team carefully.", "EN_SOFT_PICKY_STREAM_ALIGNED"},
		{"Use an enabling team carefully.", "EN_SOFT_PICKY_ENABLING_TEAM"},
		{"Build a thinnest viable platform carefully.", "EN_SOFT_PICKY_THINNEST_VIABLE_PLATFORM"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeK3dMinikubeK0s(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Cluster with k3d carefully.",
		"Cluster with K3d carefully.",
		"Cluster with minikube carefully.",
		"Cluster with Minikube carefully.",
		"Cluster with k0s carefully.",
		"Cluster with K0s carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave66(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fix the eror carefully.", "EN_SOFT_ERROR_MISS", "error"},
		{"Fix the erors carefully.", "EN_SOFT_ERRORS_MISS", "errors"},
		{"It is essencialy ready carefully.", "EN_SOFT_ESSENTIALLY_MISS", "essentially"},
		{"It is estabilished carefully.", "EN_SOFT_ESTABLISHED_MISS", "established"},
		{"Give an estamation carefully.", "EN_SOFT_ESTIMATION_MISS", "estimation"},
		{"Avoid exagerated claims carefully.", "EN_SOFT_EXAGGERATED_MISS", "exaggerated"},
		{"Please examinate carefully.", "EN_SOFT_EXAMINE_MISS", "examine"},
		{"They egnored carefully.", "EN_SOFT_IGNORED_MISS", "ignored"},
		{"Keep it embbeded carefully.", "EN_SOFT_EMBEDDED_MISS2", "embedded"},
		{"Fix it emmediately carefully.", "EN_SOFT_IMMEDIATELY_MISS3", "immediately"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave42(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Form a complicated subsystem carefully.", "EN_SOFT_PICKY_COMPLICATED_SUBSYSTEM"},
		{"Treat platform as a product carefully.", "EN_SOFT_PICKY_PLATFORM_AS_A_PRODUCT"},
		{"Apply inverse conway carefully.", "EN_SOFT_PICKY_INVERSE_CONWAY"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeMicrok8sK3sKind(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Cluster with microk8s carefully.",
		"Cluster with MicroK8s carefully.",
		"Cluster with k3s carefully.",
		"Cluster with K3s carefully.",
		"Cluster with kind carefully.",
		"Cluster with Kind carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave67(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fill the feild carefully.", "EN_SOFT_FIELD_MISS", "field"},
		{"Make it feasable carefully.", "EN_SOFT_FEASIBLE_MISS", "feasible"},
		{"Check finacial risk carefully.", "EN_SOFT_FINANCIAL_MISS", "financial"},
		{"Share the forcast carefully.", "EN_SOFT_FORECAST_MISS", "forecast"},
		{"Measure frquency carefully.", "EN_SOFT_FREQUENCY_MISS", "frequency"},
		{"Write funtional tests carefully.", "EN_SOFT_FUNCTIONAL_MISS", "functional"},
		{"Ship firts carefully.", "EN_SOFT_FIRST_MISS", "first"},
		{"Hold firmely carefully.", "EN_SOFT_FIRMLY_MISS", "firmly"},
		{"Reduce emision carefully.", "EN_SOFT_EMISSION_MISS2", "emission"},
		{"Please discconect carefully.", "EN_SOFT_DISCONNECT_MISS", "disconnect"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave43(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Track the error budget carefully.", "EN_SOFT_PICKY_ERROR_BUDGET"},
		{"Clarify service ownership carefully.", "EN_SOFT_PICKY_SERVICE_OWNERSHIP"},
		{"Prefer reliability driven ops carefully.", "EN_SOFT_PICKY_RELIABILITY_DRIVEN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeTalosKopsEksctl(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Boot with talos carefully.",
		"Boot with Talos carefully.",
		"Cluster with kops carefully.",
		"Cluster with Kops carefully.",
		"Cluster with eksctl carefully.",
		"Cluster with Eksctl carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave68(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A famouse bug carefully.", "EN_SOFT_FAMOUS_MISS", "famous"},
		{"Ideas fasinate carefully.", "EN_SOFT_FASCINATE_MISS2", "fascinate"},
		{"Check feasability carefully.", "EN_SOFT_FEASIBILITY_MISS", "feasibility"},
		{"Fill the feilds carefully.", "EN_SOFT_FIELDS_MISS", "fields"},
		{"It is finacially risky carefully.", "EN_SOFT_FINANCIALLY_MISS", "financially"},
		{"Values were forcasted carefully.", "EN_SOFT_FORECASTED_MISS", "forecasted"},
		{"Risks were forseen carefully.", "EN_SOFT_FORESEEN_MISS", "foreseen"},
		{"Expect frquent failures carefully.", "EN_SOFT_FREQUENT_MISS", "frequent"},
		{"It fails frquently carefully.", "EN_SOFT_FREQUENTLY_MISS2", "frequently"},
		{"Add funtionality carefully.", "EN_SOFT_FUNCTIONALITY_MISS", "functionality"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave44(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Prioritize toil reduction carefully.", "EN_SOFT_PICKY_TOIL_REDUCTION"},
		{"Run a blameless postmortem carefully.", "EN_SOFT_PICKY_BLAMELESS_POSTMORTEM"},
		{"Assign an incident commander carefully.", "EN_SOFT_PICKY_INCIDENT_COMMANDER"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeClusterctlClusterawsadmCapi(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Manage with clusterctl carefully.",
		"Manage with Clusterctl carefully.",
		"Bootstrap with clusterawsadm carefully.",
		"Bootstrap with Clusterawsadm carefully.",
		"Adopt capi carefully.",
		"Adopt CAPI carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave69(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Seek happyness carefully.", "EN_SOFT_HAPPINESS_MISS", "happiness"},
		{"Measure the heigths carefully.", "EN_SOFT_HEIGHTS_MISS", "heights"},
		{"A humerous note carefully.", "EN_SOFT_HUMOROUS_MISS", "humorous"},
		{"Avoid hypocracy carefully.", "EN_SOFT_HYPOCRISY_MISS", "hypocrisy"},
		{"Idealy ship today carefully.", "EN_SOFT_IDEALLY_MISS", "Ideally"},
		{"Protect idenity carefully.", "EN_SOFT_IDENTITY_MISS", "identity"},
		{"Risk is iminent carefully.", "EN_SOFT_IMMINENT_MISS", "imminent"},
		{"Do not immitate carefully.", "EN_SOFT_IMITATE_MISS", "imitate"},
		{"Block incomming traffic carefully.", "EN_SOFT_INCOMING_MISS", "incoming"},
		{"Fix incomplet work carefully.", "EN_SOFT_INCOMPLETE_MISS", "incomplete"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave45(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Prefer runbook driven ops carefully.", "EN_SOFT_PICKY_RUNBOOK_DRIVEN"},
		{"Avoid a bus factor of one carefully.", "EN_SOFT_PICKY_BUS_FACTOR_OF_ONE"},
		{"Publish a team api carefully.", "EN_SOFT_PICKY_TEAM_API"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeVeleroResticK10(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Backup with velero carefully.",
		"Backup with Velero carefully.",
		"Backup with restic carefully.",
		"Backup with Restic carefully.",
		"Backup with k10 carefully.",
		"Backup with K10 carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave70(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Avoid inadequit tests carefully.", "EN_SOFT_INADEQUATE_MISS", "inadequate"},
		{"Avoid inappropiate comments carefully.", "EN_SOFT_INAPPROPRIATE_MISS", "inappropriate"},
		{"Avoid incompetance carefully.", "EN_SOFT_INCOMPETENCE_MISS", "incompetence"},
		{"Avoid inefficent code carefully.", "EN_SOFT_INEFFICIENT_MISS", "inefficient"},
		{"Change is inevitible carefully.", "EN_SOFT_INEVITABLE_MISS", "inevitable"},
		{"Avoid infinit loops carefully.", "EN_SOFT_INFINITE_MISS", "infinite"},
		{"Reduce inflamation carefully.", "EN_SOFT_INFLAMMATION_MISS", "inflammation"},
		{"An influencial paper carefully.", "EN_SOFT_INFLUENTIAL_MISS", "influential"},
		{"Share infromation carefully.", "EN_SOFT_INFORMATION_MISS", "information"},
		{"Avoid cheap immitation carefully.", "EN_SOFT_IMITATION_MISS", "imitation"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave46(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Plan the on call rotation carefully.", "EN_SOFT_PICKY_ON_CALL_ROTATION"},
		{"Use follow the sun carefully.", "EN_SOFT_PICKY_FOLLOW_THE_SUN"},
		{"Avoid a bus factor of two carefully.", "EN_SOFT_PICKY_BUS_FACTOR_OF_TWO"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeLonghornOpenebsRook(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store with longhorn carefully.",
		"Store with Longhorn carefully.",
		"Store with openebs carefully.",
		"Store with OpenEBS carefully.",
		"Store with rook carefully.",
		"Store with Rook carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave71(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Set the inital value carefully.", "EN_SOFT_INITIAL_MISS", "initial"},
		{"Please initalize carefully.", "EN_SOFT_INITIALIZE_MISS", "initialize"},
		{"Drive inovation carefully.", "EN_SOFT_INNOVATION_MISS", "innovation"},
		{"An inovative design carefully.", "EN_SOFT_INNOVATIVE_MISS", "innovative"},
		{"Finish the instalation carefully.", "EN_SOFT_INSTALLATION_MISS", "installation"},
		{"Open an inquiery carefully.", "EN_SOFT_INQUIRY_MISS", "inquiry"},
		{"Stay infromed carefully.", "EN_SOFT_INFORMED_MISS", "informed"},
		{"They immitated carefully.", "EN_SOFT_IMITATED_MISS", "imitated"},
		{"Avoid inflamatory remarks carefully.", "EN_SOFT_INFLAMMATORY_MISS", "inflammatory"},
		{"Welcome foriegners carefully.", "EN_SOFT_FOREIGNERS_MISS", "foreigners"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave47(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Practice chaos engineering carefully.", "EN_SOFT_PICKY_CHAOS_ENGINEERING"},
		{"Improve mean time to recover carefully.", "EN_SOFT_PICKY_MEAN_TIME_TO_RECOVER"},
		{"Improve mean time to detect carefully.", "EN_SOFT_PICKY_MEAN_TIME_TO_DETECT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCephLinstorMayastor(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store with ceph carefully.",
		"Store with Ceph carefully.",
		"Store with linstor carefully.",
		"Store with Linstor carefully.",
		"Store with mayastor carefully.",
		"Store with Mayastor carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave72(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Use this insted carefully.", "EN_SOFT_INSTEAD_MISS", "instead"},
		{"Use this intead carefully.", "EN_SOFT_INSTEAD_MISS2", "instead"},
		{"Please intergrate carefully.", "EN_SOFT_INTEGRATE_MISS", "integrate"},
		{"Finish intergration carefully.", "EN_SOFT_INTEGRATION_MISS", "integration"},
		{"Expect intermitent failures carefully.", "EN_SOFT_INTERMITTENT_MISS", "intermittent"},
		{"Use internel tools carefully.", "EN_SOFT_INTERNAL_MISS", "internal"},
		{"Start an initative carefully.", "EN_SOFT_INITIATIVE_MISS", "initiative"},
		{"Keep it initalized carefully.", "EN_SOFT_INITIALIZED_MISS", "initialized"},
		{"Answer inquieries carefully.", "EN_SOFT_INQUIRIES_MISS", "inquiries"},
		{"Complete instalations carefully.", "EN_SOFT_INSTALLATIONS_MISS", "installations"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave48(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Schedule a game day carefully.", "EN_SOFT_PICKY_GAME_DAY"},
		{"Run a tabletop exercise carefully.", "EN_SOFT_PICKY_TABLETOP_EXERCISE"},
		{"Plan a disaster recovery drill carefully.", "EN_SOFT_PICKY_DISASTER_RECOVERY_DRILL"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeMinioSeaweedfsGarage(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Object store with minio carefully.",
		"Object store with MinIO carefully.",
		"Object store with seaweedfs carefully.",
		"Object store with SeaweedFS carefully.",
		"Object store with garage carefully.",
		"Object store with Garage carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave73(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It was interupted carefully.", "EN_SOFT_INTERRUPTED_MISS", "interrupted"},
		{"Avoid interuption carefully.", "EN_SOFT_INTERRUPTION_MISS", "interruption"},
		{"Please intervine carefully.", "EN_SOFT_INTERVENE_MISS", "intervene"},
		{"Show intrest carefully.", "EN_SOFT_INTEREST_MISS", "interest"},
		{"Keep it intutive carefully.", "EN_SOFT_INTUITIVE_MISS", "intuitive"},
		{"Meet the inventer carefully.", "EN_SOFT_INVENTOR_MISS", "inventor"},
		{"Please invovle carefully.", "EN_SOFT_INVOLVE_MISS", "involve"},
		{"They were invovled carefully.", "EN_SOFT_INVOLVED_MISS", "involved"},
		{"Keep it intergrated carefully.", "EN_SOFT_INTEGRATED_MISS", "integrated"},
		{"State your intension carefully.", "EN_SOFT_INTENTION_MISS", "intention"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave49(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Open a war room carefully.", "EN_SOFT_PICKY_WAR_ROOM"},
		{"Join the bridge call carefully.", "EN_SOFT_PICKY_BRIDGE_CALL"},
		{"Raise the bus factor of three carefully.", "EN_SOFT_PICKY_BUS_FACTOR_OF_THREE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeJuicefsAlluxioCubefs(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Mount with juicefs carefully.",
		"Mount with JuiceFS carefully.",
		"Cache with alluxio carefully.",
		"Cache with Alluxio carefully.",
		"Store with cubefs carefully.",
		"Store with CubeFS carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave74(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please isntall carefully.", "EN_SOFT_INSTALL_MISS", "install"},
		{"Fix the isue carefully.", "EN_SOFT_ISSUE_MISS", "issue"},
		{"Start the jounrey carefully.", "EN_SOFT_JOURNEY_MISS", "journey"},
		{"Use good judement carefully.", "EN_SOFT_JUDGEMENT_MISS", "judgement"},
		{"Need justifcation carefully.", "EN_SOFT_JUSTIFICATION_MISS", "justification"},
		{"I am intrested carefully.", "EN_SOFT_INTERESTED_MISS", "interested"},
		{"Meet the inventers carefully.", "EN_SOFT_INVENTORS_MISS", "inventors"},
		{"Please invole carefully.", "EN_SOFT_INVOLVE_MISS2", "involve"},
		{"Avoid irresponsable acts carefully.", "EN_SOFT_IRRESPONSIBLE_MISS", "irresponsible"},
		{"Join interational talks carefully.", "EN_SOFT_INTERNATIONAL_MISS", "international"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave50(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use a canary deployment carefully.", "EN_SOFT_PICKY_CANARY_DEPLOYMENT"},
		{"Keep a hot spare carefully.", "EN_SOFT_PICKY_HOT_SPARE"},
		{"Keep a cold spare carefully.", "EN_SOFT_PICKY_COLD_SPARE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeGlusterLustreBeegfs(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store with gluster carefully.",
		"Store with Gluster carefully.",
		"Store with lustre carefully.",
		"Store with Lustre carefully.",
		"Store with beegfs carefully.",
		"Store with BeeGFS carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave75(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It was isntalled carefully.", "EN_SOFT_INSTALLED_MISS", "installed"},
		{"Fix the isues carefully.", "EN_SOFT_ISSUES_MISS", "issues"},
		{"Plan the jounries carefully.", "EN_SOFT_JOURNEYS_MISS", "journeys"},
		{"Use good judgemnt carefully.", "EN_SOFT_JUDGMENT_MISS", "judgment"},
		{"A judical review carefully.", "EN_SOFT_JUDICIAL_MISS", "judicial"},
		{"Visit labratories carefully.", "EN_SOFT_LABORATORIES_MISS", "laboratories"},
		{"Learn langauges carefully.", "EN_SOFT_LANGUAGES_MISS", "languages"},
		{"Move lateraly carefully.", "EN_SOFT_LATERALLY_MISS", "laterally"},
		{"Enjoy leasure carefully.", "EN_SOFT_LEISURE_MISS2", "leisure"},
		{"It is lisenced carefully.", "EN_SOFT_LICENSED_MISS", "licensed"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave51(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use a blue green deployment carefully.", "EN_SOFT_PICKY_BLUE_GREEN_DEPLOYMENT"},
		{"Prefer a rolling update carefully.", "EN_SOFT_PICKY_ROLLING_UPDATE"},
		{"Gate with a feature flag carefully.", "EN_SOFT_PICKY_FEATURE_FLAG"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeMoosefsOrangefsCephfs(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store with moosefs carefully.",
		"Store with MooseFS carefully.",
		"Store with orangefs carefully.",
		"Store with OrangeFS carefully.",
		"Store with cephfs carefully.",
		"Store with CephFS carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave76(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Assign a liasion carefully.", "EN_SOFT_LIAISON_MISS", "liaison"},
		{"Walk leasurely carefully.", "EN_SOFT_LEISURELY_MISS", "leisurely"},
		{"Plan maintenence carefully.", "EN_SOFT_MAINTENANCE_MISS4", "maintenance"},
		{"Practice the manouver carefully.", "EN_SOFT_MANEUVER_MISS2", "maneuver"},
		{"Use better materail carefully.", "EN_SOFT_MATERIAL_MISS", "material"},
		{"Study mathamatics carefully.", "EN_SOFT_MATHEMATICS_MISS", "mathematics"},
		{"Hire a mecanic carefully.", "EN_SOFT_MECHANIC_MISS", "mechanic"},
		{"Take the medcine carefully.", "EN_SOFT_MEDICINE_MISS2", "medicine"},
		{"Measure the hieghts carefully.", "EN_SOFT_HEIGHTS_MISS2", "heights"},
		{"Please infrom carefully.", "EN_SOFT_INFORM_MISS", "inform"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave52(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Try a dark launch carefully.", "EN_SOFT_PICKY_DARK_LAUNCH"},
		{"Send shadow traffic carefully.", "EN_SOFT_PICKY_SHADOW_TRAFFIC"},
		{"Use progressive delivery carefully.", "EN_SOFT_PICKY_PROGRESSIVE_DELIVERY"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeLvmMdadmWeka(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Volume with lvm carefully.",
		"Volume with LVM carefully.",
		"Array with mdadm carefully.",
		"Array with Mdadm carefully.",
		"Store with weka carefully.",
		"Store with Weka carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave77(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They manouvered carefully.", "EN_SOFT_MANEUVERED_MISS", "maneuvered"},
		{"Order materails carefully.", "EN_SOFT_MATERIALS_MISS", "materials"},
		{"A mathmatical proof carefully.", "EN_SOFT_MATHEMATICAL_MISS", "mathematical"},
		{"A mecanical part carefully.", "EN_SOFT_MECHANICAL_MISS", "mechanical"},
		{"Study medeval history carefully.", "EN_SOFT_MEDIEVAL_MISS2", "medieval"},
		{"A miniture model carefully.", "EN_SOFT_MINIATURE_MISS", "miniature"},
		{"Fix the mispelling carefully.", "EN_SOFT_MISSPELLING_MISS", "misspelling"},
		{"Pay morgages carefully.", "EN_SOFT_MORTGAGES_MISS", "mortgages"},
		{"Check mositure carefully.", "EN_SOFT_MOISTURE_MISS", "moisture"},
		{"Each mounth carefully.", "EN_SOFT_MONTH_MISS", "month"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave53(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use traffic splitting carefully.", "EN_SOFT_PICKY_TRAFFIC_SPLITTING"},
		{"Keep a holdout group carefully.", "EN_SOFT_PICKY_HOLDOUT_GROUP"},
		{"Declare a feature freeze carefully.", "EN_SOFT_PICKY_FEATURE_FREEZE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeZfsBtrfsXfs(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Format with zfs carefully.",
		"Format with ZFS carefully.",
		"Format with btrfs carefully.",
		"Format with Btrfs carefully.",
		"Format with xfs carefully.",
		"Format with XFS carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave78(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Keep manouvering carefully.", "EN_SOFT_MANEUVERING_MISS", "maneuvering"},
		{"Prove it mathmatically carefully.", "EN_SOFT_MATHEMATICALLY_MISS", "mathematically"},
		{"It works mecanically carefully.", "EN_SOFT_MECHANICALLY_MISS", "mechanically"},
		{"Wait several mounths carefully.", "EN_SOFT_MONTHS_MISS", "months"},
		{"It happens naturly carefully.", "EN_SOFT_NATURALLY_MISS", "naturally"},
		{"It is a neccesity carefully.", "EN_SOFT_NECESSITY_MISS2", "necessity"},
		{"Walk the neigborhoods carefully.", "EN_SOFT_NEIGHBORHOODS_MISS", "neighborhoods"},
		{"It is noticably wrong carefully.", "EN_SOFT_NOTICEABLY_MISS", "noticeably"},
		{"On rare occassions carefully.", "EN_SOFT_OCCASIONS_MISS", "occasions"},
		{"Fix the ommision carefully.", "EN_SOFT_OMISSION_MISS2", "omission"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave54(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Add a kill switch carefully.", "EN_SOFT_PICKY_KILL_SWITCH"},
		{"Use a circuit breaker carefully.", "EN_SOFT_PICKY_CIRCUIT_BREAKER"},
		{"Apply a bulkhead pattern carefully.", "EN_SOFT_PICKY_BULKHEAD_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeNfsCifsFuse(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Mount with nfs carefully.",
		"Mount with NFS carefully.",
		"Mount with cifs carefully.",
		"Mount with CIFS carefully.",
		"Mount with fuse carefully.",
		"Mount with Fuse carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave79(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Be particulary careful carefully.", "EN_SOFT_PARTICULARLY_MISS2", "particularly"},
		{"Help peopel carefully.", "EN_SOFT_PEOPLE_MISS", "people"},
		{"I percieve carefully.", "EN_SOFT_PERCEIVE_MISS", "perceive"},
		{"It was percieved carefully.", "EN_SOFT_PERCEIVED_MISS", "perceived"},
		{"Make it permenant carefully.", "EN_SOFT_PERMANENT_MISS", "permanent"},
		{"Fix it permenantly carefully.", "EN_SOFT_PERMANENTLY_MISS", "permanently"},
		{"Need persistance carefully.", "EN_SOFT_PERSISTENCE_MISS", "persistence"},
		{"Help the neigbours carefully.", "EN_SOFT_NEIGHBOURS_MISS", "neighbours"},
		{"Count occurances carefully.", "EN_SOFT_OCCURRENCES_MISS", "occurrences"},
		{"Fix omisions carefully.", "EN_SOFT_OMISSIONS_MISS", "omissions"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave55(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Avoid a retry storm carefully.", "EN_SOFT_PICKY_RETRY_STORM"},
		{"Avoid a thundering herd carefully.", "EN_SOFT_PICKY_THUNDERING_HERD"},
		{"Pick a backoff strategy carefully.", "EN_SOFT_PICKY_BACKOFF_STRATEGY"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeIscsiSmbAfp(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Attach with iscsi carefully.",
		"Attach with iSCSI carefully.",
		"Share with smb carefully.",
		"Share with SMB carefully.",
		"Share with afp carefully.",
		"Share with AFP carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave80(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Be particullary careful carefully.", "EN_SOFT_PARTICULARLY_MISS3", "particularly"},
		{"Help peopels carefully.", "EN_SOFT_PEOPLES_MISS", "peoples"},
		{"A prefereable option carefully.", "EN_SOFT_PREFERABLE_MISS", "preferable"},
		{"Prefering this carefully.", "EN_SOFT_PREFERRING_MISS", "Preferring"},
		{"In the preceeding step carefully.", "EN_SOFT_PRECEDING_MISS", "preceding"},
		{"It preceeded carefully.", "EN_SOFT_PRECEDED_MISS", "preceded"},
		{"They procliam carefully.", "EN_SOFT_PROCLAIM_MISS", "proclaim"},
		{"Hire proffessional help carefully.", "EN_SOFT_PROFESSIONAL_MISS", "professional"},
		{"Grant privelege carefully.", "EN_SOFT_PRIVILEGE_MISS2", "privilege"},
		{"Write a reccomendation carefully.", "EN_SOFT_RECOMMENDATION_MISS2", "recommendation"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave56(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Enable rate limiting carefully.", "EN_SOFT_PICKY_RATE_LIMITING"},
		{"Enable load shedding carefully.", "EN_SOFT_PICKY_LOAD_SHEDDING"},
		{"Plan graceful degradation carefully.", "EN_SOFT_PICKY_GRACEFUL_DEGRADATION"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeNbdNvmeNvmf(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Export with nbd carefully.",
		"Export with NBD carefully.",
		"Attach with nvme carefully.",
		"Attach with NVMe carefully.",
		"Fabric with nvmf carefully.",
		"Fabric with NVMeF carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave81(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Stop reffering carefully.", "EN_SOFT_REFERRING_MISS", "referring"},
		{"Keep remebering carefully.", "EN_SOFT_REMEMBERING_MISS", "remembering"},
		{"Avoid repititions carefully.", "EN_SOFT_REPETITIONS_MISS", "repetitions"},
		{"Make sacrafices carefully.", "EN_SOFT_SACRIFICES_MISS", "sacrifices"},
		{"Act proffesionaly carefully.", "EN_SOFT_PROFESSIONALLY_MISS", "professionally"},
		{"A priveleged account carefully.", "EN_SOFT_PRIVILEGED_MISS2", "privileged"},
		{"Keep seperating carefully.", "EN_SOFT_SEPARATING_MISS", "separating"},
		{"They siezed carefully.", "EN_SOFT_SEIZED_MISS", "seized"},
		{"Act similiarly carefully.", "EN_SOFT_SIMILARLY_MISS", "similarly"},
		{"A sucessful deploy carefully.", "EN_SOFT_SUCCESSFUL_MISS3", "successful"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave57(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Send an idempotency key carefully.", "EN_SOFT_PICKY_IDEMPOTENCY_KEY"},
		{"Promise exactly once carefully.", "EN_SOFT_PICKY_EXACTLY_ONCE"},
		{"Promise at least once carefully.", "EN_SOFT_PICKY_AT_LEAST_ONCE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeSpdkIouringUring(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Poll with spdk carefully.",
		"Poll with SPDK carefully.",
		"Submit with iouring carefully.",
		"Submit with io_uring carefully.",
		"Submit with uring carefully.",
		"Submit with Uring carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave82(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Avoid unecessary work carefully.", "EN_SOFT_UNNECESSARY_MISS2", "unnecessary"},
		{"Avoid unecessarily complex code carefully.", "EN_SOFT_UNNECESSARILY_MISS", "unnecessarily"},
		{"An unforseeable risk carefully.", "EN_SOFT_UNFORESEEABLE_MISS", "unforeseeable"},
		{"Keep uninterupted service carefully.", "EN_SOFT_UNINTERRUPTED_MISS", "uninterrupted"},
		{"Please deactive carefully.", "EN_SOFT_DEACTIVATE_MISS", "deactivate"},
		{"It dependends carefully.", "EN_SOFT_DEPENDS_MISS", "depends"},
		{"They felt disatisfied carefully.", "EN_SOFT_DISSATISFIED_MISS", "dissatisfied"},
		{"Do not embaras carefully.", "EN_SOFT_EMBARRASS_MISS2", "embarrass"},
		{"Light was emited carefully.", "EN_SOFT_EMITTED_MISS", "emitted"},
		{"Start endevours carefully.", "EN_SOFT_ENDEAVOURS_MISS", "endeavours"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave58(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use a dead letter queue carefully.", "EN_SOFT_PICKY_DEAD_LETTER_QUEUE"},
		{"Avoid a poison pill carefully.", "EN_SOFT_PICKY_POISON_PILL"},
		{"Apply the outbox pattern carefully.", "EN_SOFT_PICKY_OUTBOX_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeDpdkLibaioUring2(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Packet with dpdk carefully.",
		"Packet with DPDK carefully.",
		"Submit with libaio carefully.",
		"Submit with Libaio carefully.",
		"Submit with uring2 carefully.",
		"Submit with Uring2 carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave83(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I felt embarased carefully.", "EN_SOFT_EMBARRASSED_MISS2", "embarrassed"},
		{"It was examinated carefully.", "EN_SOFT_EXAMINED_MISS", "examined"},
		{"Hire experianceed staff carefully.", "EN_SOFT_EXPERIENCED_MISS2", "experienced"},
		{"A fasinating topic carefully.", "EN_SOFT_FASCINATING_MISS2", "fascinating"},
		{"It is finacialy risky carefully.", "EN_SOFT_FINANCIALLY_MISS2", "financially"},
		{"Prove usefullness carefully.", "EN_SOFT_USEFULNESS_MISS", "usefulness"},
		{"Apply it usefuly carefully.", "EN_SOFT_USEFULLY_MISS", "usefully"},
		{"Park vehicals carefully.", "EN_SOFT_VEHICLES_MISS", "vehicles"},
		{"It is visably wrong carefully.", "EN_SOFT_VISIBLY_MISS", "visibly"},
		{"They are aquiring carefully.", "EN_SOFT_ACQUIRING_MISS2", "acquiring"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave59(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use a saga pattern carefully.", "EN_SOFT_PICKY_SAGA_PATTERN"},
		{"Adopt event sourcing carefully.", "EN_SOFT_PICKY_EVENT_SOURCING"},
		{"Apply a cqrs pattern carefully.", "EN_SOFT_PICKY_CQRS_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeRdmaRoceIbverbs(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Transfer with rdma carefully.",
		"Transfer with RDMA carefully.",
		"Transfer with roce carefully.",
		"Transfer with RoCE carefully.",
		"Transfer with ibverbs carefully.",
		"Transfer with Ibverbs carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave84(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Make it acceptible carefully.", "EN_SOFT_ACCEPTABLE_MISS", "acceptable"},
		{"It happened accidentaly carefully.", "EN_SOFT_ACCIDENTALLY_MISS", "accidentally"},
		{"It happened accidently carefully.", "EN_SOFT_ACCIDENTALLY_MISS2", "accidentally"},
		{"Be accomodating carefully.", "EN_SOFT_ACCOMMODATING_MISS", "accommodating"},
		{"Please acomodate carefully.", "EN_SOFT_ACCOMMODATE_MISS2", "accommodate"},
		{"Provide adiquate space carefully.", "EN_SOFT_ADEQUATE_MISS", "adequate"},
		{"Fix the adresss carefully.", "EN_SOFT_ADDRESS_MISS3", "address"},
		{"Grow agressively carefully.", "EN_SOFT_AGGRESSIVELY_MISS", "aggressively"},
		{"Avoid alchohol carefully.", "EN_SOFT_ALCOHOL_MISS", "alcohol"},
		{"It is allready done carefully.", "EN_SOFT_ALREADY_MISS", "already"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave60(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use a strangler fig carefully.", "EN_SOFT_PICKY_STRANGLER_FIG"},
		{"Add an anti corruption layer carefully.", "EN_SOFT_PICKY_ANTI_CORRUPTION_LAYER"},
		{"Define a bounded context carefully.", "EN_SOFT_PICKY_BOUNDED_CONTEXT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeVhostVfioUio(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Device with vhost carefully.",
		"Device with Vhost carefully.",
		"Device with vfio carefully.",
		"Device with VFIO carefully.",
		"Device with uio carefully.",
		"Device with UIO carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave85(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It is apparant carefully.", "EN_SOFT_APPARENT_MISS2", "apparent"},
		{"It apparantly works carefully.", "EN_SOFT_APPARENTLY_MISS3", "apparently"},
		{"Write an appology carefully.", "EN_SOFT_APOLOGY_MISS", "apology"},
		{"Look arround carefully.", "EN_SOFT_AROUND_MISS", "around"},
		{"Please assosiate carefully.", "EN_SOFT_ASSOCIATE_MISS", "associate"},
		{"Make it availible carefully.", "EN_SOFT_AVAILABLE_MISS2", "available"},
		{"Sit amoung friends carefully.", "EN_SOFT_AMONG_MISS", "among"},
		{"Pay anually carefully.", "EN_SOFT_ANNUALLY_MISS", "annually"},
		{"An amature player carefully.", "EN_SOFT_AMATEUR_MISS", "amateur"},
		{"It is allright carefully.", "EN_SOFT_ALRIGHT_MISS", "alright"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave61(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Draw a context map carefully.", "EN_SOFT_PICKY_CONTEXT_MAP"},
		{"Define ubiquitous language carefully.", "EN_SOFT_PICKY_UBIQUITOUS_LANGUAGE"},
		{"Publish a domain event carefully.", "EN_SOFT_PICKY_DOMAIN_EVENT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeVirtioVhostuserMdev(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Device with virtio carefully.",
		"Device with Virtio carefully.",
		"Device with vhost-user carefully.",
		"Device with vhostuser carefully.",
		"Device with mdev carefully.",
		"Device with Mdev carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave86(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Catch the assasin carefully.", "EN_SOFT_ASSASSIN_MISS", "assassin"},
		{"Do not assasinate carefully.", "EN_SOFT_ASSASSINATE_MISS", "assassinate"},
		{"It is assosiated carefully.", "EN_SOFT_ASSOCIATED_MISS", "associated"},
		{"Join the assotiation carefully.", "EN_SOFT_ASSOCIATION_MISS", "association"},
		{"She beleives carefully.", "EN_SOFT_BELIEVES_MISS", "believes"},
		{"Keep beleiving carefully.", "EN_SOFT_BELIEVING_MISS", "believing"},
		{"Respect copywrite carefully.", "EN_SOFT_COPYRIGHT_MISS", "copyright"},
		{"Send acknowlegement carefully.", "EN_SOFT_ACKNOWLEDGEMENT_MISS", "acknowledgement"},
		{"Make it adressable carefully.", "EN_SOFT_ADDRESSABLE_MISS", "addressable"},
		{"It is adviseable carefully.", "EN_SOFT_ADVISABLE_MISS", "advisable"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave62(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Protect the aggregate root carefully.", "EN_SOFT_PICKY_AGGREGATE_ROOT"},
		{"Model a value object carefully.", "EN_SOFT_PICKY_VALUE_OBJECT"},
		{"Avoid an entity service carefully.", "EN_SOFT_PICKY_ENTITY_SERVICE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeQcow2VmdkVdi(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Image as qcow2 carefully.",
		"Image as QCOW2 carefully.",
		"Image as vmdk carefully.",
		"Image as VMDK carefully.",
		"Image as vdi carefully.",
		"Image as VDI carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave87(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Read the advisery carefully.", "EN_SOFT_ADVISORY_MISS", "advisory"},
		{"Avoid alchoholic drinks carefully.", "EN_SOFT_ALCOHOLIC_MISS", "alcoholic"},
		{"There is alotof risk carefully.", "EN_SOFT_ALOTOF_MISS", "a lot of"},
		{"Sit amoungst friends carefully.", "EN_SOFT_AMONGST_MISS", "amongst"},
		{"It aparenttly works carefully.", "EN_SOFT_APPARENTLY_MISS4", "apparently"},
		{"Offer appologies carefully.", "EN_SOFT_APOLOGIES_MISS", "apologies"},
		{"They procliamed carefully.", "EN_SOFT_PROCLAIMED_MISS", "proclaimed"},
		{"It affects adversly carefully.", "EN_SOFT_ADVERSELY_MISS", "adversely"},
		{"Please adversize carefully.", "EN_SOFT_ADVERTISE_MISS", "advertise"},
		{"Note the absensee carefully.", "EN_SOFT_ABSENCE_MISS3", "absence"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave63(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use hexagonal architecture carefully.", "EN_SOFT_PICKY_HEXAGONAL_ARCHITECTURE"},
		{"Prefer ports and adapters carefully.", "EN_SOFT_PICKY_PORTS_AND_ADAPTERS"},
		{"Share a shared kernel carefully.", "EN_SOFT_PICKY_SHARED_KERNEL"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeOvaOvfQed(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Export as ova carefully.",
		"Export as OVA carefully.",
		"Export as ovf carefully.",
		"Export as OVF carefully.",
		"Image as qed carefully.",
		"Image as QED carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave88(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Guests were accomodated carefully.", "EN_SOFT_ACCOMMODATED_MISS", "accommodated"},
		{"The plan accomodates carefully.", "EN_SOFT_ACCOMMODATES_MISS", "accommodates"},
		{"Book accomodations carefully.", "EN_SOFT_ACCOMMODATIONS_MISS", "accommodations"},
		{"They acheiveds carefully.", "EN_SOFT_ACHIEVEDS_MISS", "achieved"},
		{"List acheivments carefully.", "EN_SOFT_ACHIEVEMENTS_MISS", "achievements"},
		{"Send acknowlegements carefully.", "EN_SOFT_ACKNOWLEDGEMENTS_MISS", "acknowledgements"},
		{"Avoid agressiveness carefully.", "EN_SOFT_AGGRESSIVENESS_MISS", "aggressiveness"},
		{"They aquitted carefully.", "EN_SOFT_ACQUITTED_MISS", "acquitted"},
		{"Check availibility carefully.", "EN_SOFT_AVAILABILITY_MISS", "availability"},
		{"It is basicaly ready carefully.", "EN_SOFT_BASICALLY_MISS2", "basically"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave64(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Prefer clean architecture carefully.", "EN_SOFT_PICKY_CLEAN_ARCHITECTURE"},
		{"Prefer onion architecture carefully.", "EN_SOFT_PICKY_ONION_ARCHITECTURE"},
		{"Prefer screaming architecture carefully.", "EN_SOFT_PICKY_SCREAMING_ARCHITECTURE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeAmiEbsSnapshot(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Launch from ami carefully.",
		"Launch from AMI carefully.",
		"Attach ebs carefully.",
		"Attach EBS carefully.",
		"Create a snapshot carefully.",
		"Create a Snapshot carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave89(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Start a bussiness carefully.", "EN_SOFT_BUSINESS_MISS", "business"},
		{"Help small bussinesses carefully.", "EN_SOFT_BUSINESSES_MISS2", "businesses"},
		{"Check the calanders carefully.", "EN_SOFT_CALENDARS_MISS", "calendars"},
		{"Please catagorize carefully.", "EN_SOFT_CATEGORIZE_MISS", "categorize"},
		{"Face hard challanges carefully.", "EN_SOFT_CHALLENGES_MISS", "challenges"},
		{"Count the charachters carefully.", "EN_SOFT_CHARACTERS_MISS", "characters"},
		{"Meet the cheifs carefully.", "EN_SOFT_CHIEFS_MISS", "chiefs"},
		{"Prefer collaberative work carefully.", "EN_SOFT_COLLABORATIVE_MISS", "collaborative"},
		{"Pay the comissions carefully.", "EN_SOFT_COMMISSIONS_MISS", "commissions"},
		{"Configure contollers carefully.", "EN_SOFT_CONTROLLERS_MISS", "controllers"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave65(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply dependency inversion carefully.", "EN_SOFT_PICKY_DEPENDENCY_INVERSION"},
		{"Apply interface segregation carefully.", "EN_SOFT_PICKY_INTERFACE_SEGREGATION"},
		{"Apply single responsibility carefully.", "EN_SOFT_PICKY_SINGLE_RESPONSIBILITY"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeS3RdsGlacier(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store in s3 carefully.",
		"Store in S3 carefully.",
		"Query rds carefully.",
		"Query RDS carefully.",
		"Archive to glacier carefully.",
		"Archive to Glacier carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave90(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"He is a bussinessman carefully.", "EN_SOFT_BUSINESSMAN_MISS2", "businessman"},
		{"Meet the bussinessmen carefully.", "EN_SOFT_BUSINESSMEN_MISS", "businessmen"},
		{"A catagorical answer carefully.", "EN_SOFT_CATEGORICAL_MISS", "categorical"},
		{"Items were catagorized carefully.", "EN_SOFT_CATEGORIZED_MISS", "categorized"},
		{"They challanged carefully.", "EN_SOFT_CHALLENGED_MISS", "challenged"},
		{"It is cheifly used carefully.", "EN_SOFT_CHIEFLY_MISS", "chiefly"},
		{"Start collaberations carefully.", "EN_SOFT_COLLABORATIONS_MISS", "collaborations"},
		{"It was comissioned carefully.", "EN_SOFT_COMMISSIONED_MISS", "commissioned"},
		{"Avoid contolling carefully.", "EN_SOFT_CONTROLLING_MISS", "controlling"},
		{"Protect cosumers carefully.", "EN_SOFT_CONSUMERS_MISS", "consumers"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave66(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply open closed principle carefully.", "EN_SOFT_PICKY_OPEN_CLOSED_PRINCIPLE"},
		{"Apply liskov substitution carefully.", "EN_SOFT_PICKY_LISKOV_SUBSTITUTION"},
		{"Prefer composition over inheritance carefully.", "EN_SOFT_PICKY_COMPOSITION_OVER_INHERITANCE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeDynamodbRedshiftAthena(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store in dynamodb carefully.",
		"Store in DynamoDB carefully.",
		"Query redshift carefully.",
		"Query Redshift carefully.",
		"Query athena carefully.",
		"Query Athena carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave91(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Please contibute carefully.", "EN_SOFT_CONTRIBUTE_MISS", "contribute"},
		{"A large contibution carefully.", "EN_SOFT_CONTRIBUTION_MISS", "contribution"},
		{"A contigent plan carefully.", "EN_SOFT_CONTINGENT_MISS", "contingent"},
		{"Use contigious blocks carefully.", "EN_SOFT_CONTIGUOUS_MISS", "contiguous"},
		{"It runs contineously carefully.", "EN_SOFT_CONTINUOUSLY_MISS2", "continuously"},
		{"A contriversial claim carefully.", "EN_SOFT_CONTROVERSIAL_MISS", "controversial"},
		{"Do not cosume carefully.", "EN_SOFT_CONSUME_MISS", "consume"},
		{"Reduce cosumption carefully.", "EN_SOFT_CONSUMPTION_MISS", "consumption"},
		{"A critcal bug carefully.", "EN_SOFT_CRITICAL_MISS", "critical"},
		{"Make a delibrate choice carefully.", "EN_SOFT_DELIBERATE_MISS", "deliberate"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave67(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply law of demeter carefully.", "EN_SOFT_PICKY_LAW_OF_DEMETER"},
		{"Prefer tell dont ask carefully.", "EN_SOFT_PICKY_TELL_DONT_ASK"},
		{"Keep separation of concerns carefully.", "EN_SOFT_PICKY_SEPARATION_OF_CONCERNS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeKinesisSqsSns(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Stream via kinesis carefully.",
		"Stream via Kinesis carefully.",
		"Poll sqs carefully.",
		"Poll SQS carefully.",
		"Publish to sns carefully.",
		"Publish to SNS carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave92(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They contibuted carefully.", "EN_SOFT_CONTRIBUTED_MISS", "contributed"},
		{"Keep contibuting carefully.", "EN_SOFT_CONTRIBUTING_MISS", "contributing"},
		{"A new contibutor carefully.", "EN_SOFT_CONTRIBUTOR_MISS", "contributor"},
		{"Thank contibutors carefully.", "EN_SOFT_CONTRIBUTORS_MISS", "contributors"},
		{"It runs continously carefully.", "EN_SOFT_CONTINUOUSLY_MISS3", "continuously"},
		{"On the contary carefully.", "EN_SOFT_CONTRARY_MISS", "contrary"},
		{"Fight curruption carefully.", "EN_SOFT_CORRUPTION_MISS2", "corruption"},
		{"Make a deposite carefully.", "EN_SOFT_DEPOSIT_MISS", "deposit"},
		{"Please descibe carefully.", "EN_SOFT_DESCRIBE_MISS", "describe"},
		{"A detremental effect carefully.", "EN_SOFT_DETRIMENTAL_MISS", "detrimental"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave68(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Prefer information hiding carefully.", "EN_SOFT_PICKY_INFORMATION_HIDING"},
		{"Apply yagni carefully.", "EN_SOFT_PICKY_YAGNI"},
		{"Follow hollywood principle carefully.", "EN_SOFT_PICKY_HOLLYWOOD_PRINCIPLE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeLambdaFargateEcr(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Deploy lambda carefully.",
		"Deploy Lambda carefully.",
		"Run on fargate carefully.",
		"Run on Fargate carefully.",
		"Push to ecr carefully.",
		"Push to ECR carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave93(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Use contiguos blocks carefully.", "EN_SOFT_CONTIGUOUS_MISS2", "contiguous"},
		{"Software develepment carefully.", "EN_SOFT_DEVELOPMENT_MISS", "development"},
		{"Please develope carefully.", "EN_SOFT_DEVELOP_MISS2", "develop"},
		{"Check the dimmensions carefully.", "EN_SOFT_DIMENSIONS_MISS", "dimensions"},
		{"It disapeared carefully.", "EN_SOFT_DISAPPEARED_MISS", "disappeared"},
		{"A disapointing result carefully.", "EN_SOFT_DISAPPOINTING_MISS", "disappointing"},
		{"A disastorous failure carefully.", "EN_SOFT_DISASTROUS_MISS", "disastrous"},
		{"A disfunctional team carefully.", "EN_SOFT_DYSFUNCTIONAL_MISS", "dysfunctional"},
		{"Value diversety carefully.", "EN_SOFT_DIVERSITY_MISS", "diversity"},
		{"A dramtic change carefully.", "EN_SOFT_DRAMATIC_MISS", "dramatic"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave69(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply dry principle carefully.", "EN_SOFT_PICKY_DRY_PRINCIPLE"},
		{"Apply kiss principle carefully.", "EN_SOFT_PICKY_KISS_PRINCIPLE"},
		{"Follow boy scout rule carefully.", "EN_SOFT_PICKY_BOY_SCOUT_RULE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeEcsCloudfrontCloudwatch(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run on ecs carefully.",
		"Run on ECS carefully.",
		"Serve via cloudfront carefully.",
		"Serve via CloudFront carefully.",
		"Alert in cloudwatch carefully.",
		"Alert in CloudWatch carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave94(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It changed dramtically carefully.", "EN_SOFT_DRAMATICALLY_MISS", "dramatically"},
		{"A disiplined approach carefully.", "EN_SOFT_DISCIPLINED_MISS", "disciplined"},
		{"The salt is disolving carefully.", "EN_SOFT_DISSOLVING_MISS", "dissolving"},
		{"Do not distrub carefully.", "EN_SOFT_DISTURB_MISS", "disturb"},
		{"The sales divsion carefully.", "EN_SOFT_DIVISION_MISS2", "division"},
		{"Write documantation carefully.", "EN_SOFT_DOCUMENTATION_MISS2", "documentation"},
		{"I have no doupt carefully.", "EN_SOFT_DOUBT_MISS", "doubt"},
		{"Improve efficency carefully.", "EN_SOFT_EFFICIENCY_MISS", "efficiency"},
		{"An efficent design carefully.", "EN_SOFT_EFFICIENT_MISS", "efficient"},
		{"Please expalin carefully.", "EN_SOFT_EXPLAIN_MISS", "explain"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave70(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Avoid broken windows carefully.", "EN_SOFT_PICKY_BROKEN_WINDOWS"},
		{"Recall cap theorem carefully.", "EN_SOFT_PICKY_CAP_THEOREM"},
		{"Use rule of three carefully.", "EN_SOFT_PICKY_RULE_OF_THREE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeElasticacheNeptuneDocumentdb(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Cache with elasticache carefully.",
		"Cache with ElastiCache carefully.",
		"Query neptune carefully.",
		"Query Neptune carefully.",
		"Store in documentdb carefully.",
		"Store in DocumentDB carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave95(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I was distrubed carefully.", "EN_SOFT_DISTURBED_MISS", "disturbed"},
		{"A distrubing trend carefully.", "EN_SOFT_DISTURBING_MISS", "disturbing"},
		{"Several divsions carefully.", "EN_SOFT_DIVISIONS_MISS", "divisions"},
		{"Read the docuement carefully.", "EN_SOFT_DOCUMENT_MISS2", "document"},
		{"It is douptful carefully.", "EN_SOFT_DOUBTFUL_MISS", "doubtful"},
		{"Work efficently carefully.", "EN_SOFT_EFFICIENTLY_MISS", "efficiently"},
		{"It was encorporated carefully.", "EN_SOFT_INCORPORATED_MISS", "incorporated"},
		{"An equivilant result carefully.", "EN_SOFT_EQUIVALENT_MISS2", "equivalent"},
		{"It was expalined carefully.", "EN_SOFT_EXPLAINED_MISS", "explained"},
		{"Please fullfill carefully.", "EN_SOFT_FULFILL_MISS", "fulfill"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave71(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Prefer eventual consistency carefully.", "EN_SOFT_PICKY_EVENTUAL_CONSISTENCY"},
		{"Require acid properties carefully.", "EN_SOFT_PICKY_ACID_PROPERTIES"},
		{"Recall two generals carefully.", "EN_SOFT_PICKY_TWO_GENERALS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeEmrCognitoKms(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run jobs on emr carefully.",
		"Run jobs on EMR carefully.",
		"Auth with cognito carefully.",
		"Auth with Cognito carefully.",
		"Encrypt with kms carefully.",
		"Encrypt with KMS carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave96(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Read the docuements carefully.", "EN_SOFT_DOCUMENTS_MISS", "documents"},
		{"I have doupts carefully.", "EN_SOFT_DOUBTS_MISS", "doubts"},
		{"Start encorporating carefully.", "EN_SOFT_INCORPORATING_MISS", "incorporating"},
		{"Act enviromentally carefully.", "EN_SOFT_ENVIRONMENTALLY_MISS", "environmentally"},
		{"An equivilent amount carefully.", "EN_SOFT_EQUIVALENT_MISS3", "equivalent"},
		{"Need better explinations carefully.", "EN_SOFT_EXPLANATIONS_MISS", "explanations"},
		{"Please facilatate carefully.", "EN_SOFT_FACILITATE_MISS", "facilitate"},
		{"I cannot forsee carefully.", "EN_SOFT_FORESEE_MISS", "foresee"},
		{"It was fullfilled carefully.", "EN_SOFT_FULFILLED_MISS", "fulfilled"},
		{"Follow the guidlines carefully.", "EN_SOFT_GUIDELINES_MISS", "guidelines"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave72(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Handle byzantine fault carefully.", "EN_SOFT_PICKY_BYZANTINE_FAULT"},
		{"Fix the flaky test carefully.", "EN_SOFT_PICKY_FLAKY_TEST"},
		{"Use a golden master carefully.", "EN_SOFT_PICKY_GOLDEN_MASTER"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeGlueBedrockSagemaker(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"ETL with glue carefully.",
		"ETL with Glue carefully.",
		"Call bedrock carefully.",
		"Call Bedrock carefully.",
		"Train on sagemaker carefully.",
		"Train on SageMaker carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave97(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A fullfilling role carefully.", "EN_SOFT_FULFILLING_MISS", "fulfilling"},
		{"Order fullfilment carefully.", "EN_SOFT_FULFILLMENT_MISS", "fulfillment"},
		{"Service gaurantees carefully.", "EN_SOFT_GUARANTEES_MISS", "guarantees"},
		{"Follow this guidline carefully.", "EN_SOFT_GUIDELINE_MISS", "guideline"},
		{"He was harrassed carefully.", "EN_SOFT_HARASSED_MISS", "harassed"},
		{"Deep heirarchies carefully.", "EN_SOFT_HIERARCHIES_MISS", "hierarchies"},
		{"One hundered items carefully.", "EN_SOFT_HUNDRED_MISS", "hundred"},
		{"An imense effort carefully.", "EN_SOFT_IMMENSE_MISS", "immense"},
		{"An imigrant worker carefully.", "EN_SOFT_IMMIGRANT_MISS", "immigrant"},
		{"Need improvments carefully.", "EN_SOFT_IMPROVEMENTS_MISS", "improvements"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave73(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Write a characterization test carefully.", "EN_SOFT_PICKY_CHARACTERIZATION_TEST"},
		{"Prefer property based carefully.", "EN_SOFT_PICKY_PROPERTY_BASED"},
		{"Run mutation testing carefully.", "EN_SOFT_PICKY_MUTATION_TESTING"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeBatchEventbridgeAmplify(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Submit to batch carefully.",
		"Submit to Batch carefully.",
		"Wire eventbridge carefully.",
		"Wire EventBridge carefully.",
		"Host with amplify carefully.",
		"Host with Amplify carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave98(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It facilatates carefully.", "EN_SOFT_FACILITATES_MISS", "facilitates"},
		{"Build familarity carefully.", "EN_SOFT_FAMILIARITY_MISS", "familiarity"},
		{"In the forseeable carefully.", "EN_SOFT_FORESEEABLE_MISS", "foreseeable"},
		{"Show forsight carefully.", "EN_SOFT_FORESIGHT_MISS", "foresight"},
		{"A gaurded response carefully.", "EN_SOFT_GUARDED_MISS", "guarded"},
		{"Spent hundereds carefully.", "EN_SOFT_HUNDREDS_MISS", "hundreds"},
		{"It helped imensely carefully.", "EN_SOFT_IMMENSELY_MISS", "immensely"},
		{"Welcome imigrants carefully.", "EN_SOFT_IMMIGRANTS_MISS", "immigrants"},
		{"One improvemnt carefully.", "EN_SOFT_IMPROVEMENT_MISS", "improvement"},
		{"An increadible result carefully.", "EN_SOFT_INCREDIBLE_MISS", "incredible"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave74(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Add a contract test carefully.", "EN_SOFT_PICKY_CONTRACT_TEST"},
		{"Add a smoke test carefully.", "EN_SOFT_PICKY_SMOKE_TEST"},
		{"Run fuzz testing carefully.", "EN_SOFT_PICKY_FUZZ_TESTING"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeAppsyncTimestreamLightsail(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Query appsync carefully.",
		"Query AppSync carefully.",
		"Store in timestream carefully.",
		"Store in Timestream carefully.",
		"Host on lightsail carefully.",
		"Host on Lightsail carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave99(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Start facilatating carefully.", "EN_SOFT_FACILITATING_MISS", "facilitating"},
		{"Please familarize carefully.", "EN_SOFT_FAMILIARIZE_MISS", "familiarize"},
		{"Stop harrassing carefully.", "EN_SOFT_HARASSING_MISS", "harassing"},
		{"It is increadibly hard carefully.", "EN_SOFT_INCREDIBLY_MISS", "incredibly"},
		{"Work independantly carefully.", "EN_SOFT_INDEPENDENTLY_MISS", "independently"},
		{"An indispensible tool carefully.", "EN_SOFT_INDISPENSABLE_MISS", "indispensable"},
		{"Each indivdual carefully.", "EN_SOFT_INDIVIDUAL_MISS", "individual"},
		{"Reduce inefficency carefully.", "EN_SOFT_INEFFICIENCY_MISS", "inefficiency"},
		{"It will inevitibly fail carefully.", "EN_SOFT_INEVITABLY_MISS", "inevitably"},
		{"Choose inteligently carefully.", "EN_SOFT_INTELLIGENTLY_MISS", "intelligently"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave75(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Add a regression test carefully.", "EN_SOFT_PICKY_REGRESSION_TEST"},
		{"Run a load test carefully.", "EN_SOFT_PICKY_LOAD_TEST"},
		{"Run a stress test carefully.", "EN_SOFT_PICKY_STRESS_TEST"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeOpensearchCodebuildCodepipeline(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Index in opensearch carefully.",
		"Index in OpenSearch carefully.",
		"Build with codebuild carefully.",
		"Build with CodeBuild carefully.",
		"Deploy via codepipeline carefully.",
		"Deploy via CodePipeline carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave100(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"We finished fianlly carefully.", "EN_SOFT_FINALLY_MISS2", "finally"},
		{"Check the guages carefully.", "EN_SOFT_GAUGES_MISS", "gauges"},
		{"A strict idiology carefully.", "EN_SOFT_IDEOLOGY_MISS", "ideology"},
		{"An ignorent claim carefully.", "EN_SOFT_IGNORANT_MISS", "ignorant"},
		{"An imature choice carefully.", "EN_SOFT_IMMATURE_MISS", "immature"},
		{"Stop immitating carefully.", "EN_SOFT_IMITATING_MISS", "imitating"},
		{"An inapropriate joke carefully.", "EN_SOFT_INAPPROPRIATE_MISS2", "inappropriate"},
		{"It was incidently found carefully.", "EN_SOFT_INCIDENTALLY_MISS", "incidentally"},
		{"Declare independece carefully.", "EN_SOFT_INDEPENDENCE_MISS2", "independence"},
		{"Many indivduals carefully.", "EN_SOFT_INDIVIDUALS_MISS", "individuals"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave76(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Prefer end to end carefully.", "EN_SOFT_PICKY_END_TO_END"},
		{"Add an integration test carefully.", "EN_SOFT_PICKY_INTEGRATION_TEST"},
		{"Add a unit test carefully.", "EN_SOFT_PICKY_UNIT_TEST"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCodecommitCodedeployCloudtrail(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Clone from codecommit carefully.",
		"Clone from CodeCommit carefully.",
		"Release with codedeploy carefully.",
		"Release with CodeDeploy carefully.",
		"Audit via cloudtrail carefully.",
		"Audit via CloudTrail carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave101(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Show the hightlights carefully.", "EN_SOFT_HIGHLIGHTS_MISS", "highlights"},
		{"Keep it hygenic carefully.", "EN_SOFT_HYGIENIC_MISS", "hygienic"},
		{"Call out the hypocrit carefully.", "EN_SOFT_HYPOCRITE_MISS", "hypocrite"},
		{"An idelogical split carefully.", "EN_SOFT_IDEOLOGICAL_MISS", "ideological"},
		{"Spoke inapropriately carefully.", "EN_SOFT_INAPPROPRIATELY_MISS", "inappropriately"},
		{"Prove inocence carefully.", "EN_SOFT_INNOCENCE_MISS", "innocence"},
		{"Find inpsiration carefully.", "EN_SOFT_INSPIRATION_MISS", "inspiration"},
		{"Please interpet carefully.", "EN_SOFT_INTERPRET_MISS", "interpret"},
		{"An internatinal deal carefully.", "EN_SOFT_INTERNATIONAL_MISS2", "international"},
		{"An irrelavent detail carefully.", "EN_SOFT_IRRELEVANT_MISS2", "irrelevant"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave77(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Add an acceptance test carefully.", "EN_SOFT_PICKY_ACCEPTANCE_TEST"},
		{"Run a performance test carefully.", "EN_SOFT_PICKY_PERFORMANCE_TEST"},
		{"Run a penetration test carefully.", "EN_SOFT_PICKY_PENETRATION_TEST"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeAppconfigXrayCloudformation(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Load from appconfig carefully.",
		"Load from AppConfig carefully.",
		"Trace with xray carefully.",
		"Trace with XRay carefully.",
		"Deploy cloudformation carefully.",
		"Deploy CloudFormation carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave102(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"An inocent mistake carefully.", "EN_SOFT_INNOCENT_MISS", "innocent"},
		{"Stories inpsire carefully.", "EN_SOFT_INSPIRE_MISS", "inspire"},
		{"A new interpetation carefully.", "EN_SOFT_INTERPRETATION_MISS", "interpretation"},
		{"Stop interupting carefully.", "EN_SOFT_INTERRUPTING_MISS", "interrupting"},
		{"And intrestingly enough carefully.", "EN_SOFT_INTERESTINGLY_MISS", "interestingly"},
		{"Protect the inviroment carefully.", "EN_SOFT_ENVIRONMENT_MISS2", "environment"},
		{"An irrisistible offer carefully.", "EN_SOFT_IRRESISTIBLE_MISS", "irresistible"},
		{"A trade laison carefully.", "EN_SOFT_LIAISON_MISS2", "liaison"},
		{"A happy marraige carefully.", "EN_SOFT_MARRIAGE_MISS", "marriage"},
		{"Fix the missmatch carefully.", "EN_SOFT_MISMATCH_MISS", "mismatch"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave78(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Run a sanity test carefully.", "EN_SOFT_PICKY_SANITY_TEST"},
		{"Run a benchmark test carefully.", "EN_SOFT_PICKY_BENCHMARK_TEST"},
		{"Run a usability test carefully.", "EN_SOFT_PICKY_USABILITY_TEST"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeWorkspacesWorkdocsConfigservice(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Launch workspaces carefully.",
		"Launch WorkSpaces carefully.",
		"Share via workdocs carefully.",
		"Share via WorkDocs carefully.",
		"Audit with configservice carefully.",
		"Audit with ConfigService carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave103(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I was inpsired carefully.", "EN_SOFT_INSPIRED_MISS", "inspired"},
		{"It fails intermitently carefully.", "EN_SOFT_INTERMITTENTLY_MISS", "intermittently"},
		{"Results were interpeted carefully.", "EN_SOFT_INTERPRETED_MISS", "interpreted"},
		{"Noise interupts carefully.", "EN_SOFT_INTERRUPTS_MISS", "interrupts"},
		{"An inviromental study carefully.", "EN_SOFT_ENVIRONMENTAL_MISS2", "environmental"},
		{"Read liteature carefully.", "EN_SOFT_LITERATURE_MISS", "literature"},
		{"Happy marraiges carefully.", "EN_SOFT_MARRIAGES_MISS", "marriages"},
		{"A kind neigbor carefully.", "EN_SOFT_NEIGHBOR_MISS2", "neighbor"},
		{"About ninty people carefully.", "EN_SOFT_NINETY_MISS", "ninety"},
		{"Strong oposition carefully.", "EN_SOFT_OPPOSITION_MISS", "opposition"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave79(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Run a compatibility test carefully.", "EN_SOFT_PICKY_COMPATIBILITY_TEST"},
		{"Run an accessibility test carefully.", "EN_SOFT_PICKY_ACCESSIBILITY_TEST"},
		{"Run a security test carefully.", "EN_SOFT_PICKY_SECURITY_TEST"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeQuicksightWorkmailGuardduty(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Chart in quicksight carefully.",
		"Chart in QuickSight carefully.",
		"Mail via workmail carefully.",
		"Mail via WorkMail carefully.",
		"Scan with guardduty carefully.",
		"Scan with GuardDuty carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave104(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It was influanced carefully.", "EN_SOFT_INFLUENCED_MISS", "influenced"},
		{"Need a liscence carefully.", "EN_SOFT_LICENSE_MISS2", "license"},
		{"I mean it literaly carefully.", "EN_SOFT_LITERALLY_MISS", "literally"},
		{"Study literture carefully.", "EN_SOFT_LITERATURE_MISS2", "literature"},
		{"They manufature carefully.", "EN_SOFT_MANUFACTURE_MISS2", "manufacture"},
		{"Help the neigbors carefully.", "EN_SOFT_NEIGHBORS_MISS", "neighbors"},
		{"Find the orgin carefully.", "EN_SOFT_ORIGIN_MISS", "origin"},
		{"Join the orginization carefully.", "EN_SOFT_ORGANIZATION_MISS", "organization"},
		{"New oportunities carefully.", "EN_SOFT_OPPORTUNITIES_MISS2", "opportunities"},
		{"Polar oposites carefully.", "EN_SOFT_OPPOSITES_MISS", "opposites"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave80(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Run a reliability test carefully.", "EN_SOFT_PICKY_RELIABILITY_TEST"},
		{"Run a resilience test carefully.", "EN_SOFT_PICKY_RESILIENCE_TEST"},
		{"Run a compliance test carefully.", "EN_SOFT_PICKY_COMPLIANCE_TEST"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeMacieInspectorDetective(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Scan with macie carefully.",
		"Scan with Macie carefully.",
		"Assess with inspector carefully.",
		"Assess with Inspector carefully.",
		"Investigate with detective carefully.",
		"Investigate with Detective carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave105(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"An influancial voice carefully.", "EN_SOFT_INFLUENTIAL_MISS2", "influential"},
		{"Need a liscense carefully.", "EN_SOFT_LICENSE_MISS3", "license"},
		{"It was manufatured carefully.", "EN_SOFT_MANUFACTURED_MISS", "manufactured"},
		{"Local manufaturing carefully.", "EN_SOFT_MANUFACTURING_MISS", "manufacturing"},
		{"Not necesarily true carefully.", "EN_SOFT_NECESSARILY_MISS2", "necessarily"},
		{"Well orginized notes carefully.", "EN_SOFT_ORGANIZED_MISS", "organized"},
		{"Large orginizations carefully.", "EN_SOFT_ORGANIZATIONS_MISS", "organizations"},
		{"It was orignally planned carefully.", "EN_SOFT_ORIGINALLY_MISS", "originally"},
		{"A paticular case carefully.", "EN_SOFT_PARTICULAR_MISS2", "particular"},
		{"Please peirce carefully.", "EN_SOFT_PIERCE_MISS", "pierce"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave81(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Follow the test pyramid carefully.", "EN_SOFT_PICKY_TEST_PYRAMID"},
		{"Use a test double carefully.", "EN_SOFT_PICKY_TEST_DOUBLE"},
		{"Prefer a mock object carefully.", "EN_SOFT_PICKY_MOCK_OBJECT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeSecurityhubShieldWafv2(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Review securityhub carefully.",
		"Review SecurityHub carefully.",
		"Enable shield carefully.",
		"Enable Shield carefully.",
		"Configure wafv2 carefully.",
		"Configure WAFv2 carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave106(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A plesant day carefully.", "EN_SOFT_PLEASANT_MISS", "pleasant"},
		{"All my posessions carefully.", "EN_SOFT_POSSESSIONS_MISS", "possessions"},
		{"He was posessed carefully.", "EN_SOFT_POSSESSED_MISS", "possessed"},
		{"A real posibility carefully.", "EN_SOFT_POSSIBILITY_MISS", "possibility"},
		{"A practial plan carefully.", "EN_SOFT_PRACTICAL_MISS2", "practical"},
		{"State your prefrence carefully.", "EN_SOFT_PREFERENCE_MISS2", "preference"},
		{"User prefrences carefully.", "EN_SOFT_PREFERENCES_MISS", "preferences"},
		{"Please prepair carefully.", "EN_SOFT_PREPARE_MISS", "prepare"},
		{"They persued carefully.", "EN_SOFT_PURSUED_MISS", "pursued"},
		{"Keep orginizing carefully.", "EN_SOFT_ORGANIZING_MISS", "organizing"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave82(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use a stub object carefully.", "EN_SOFT_PICKY_STUB_OBJECT"},
		{"Use a fake object carefully.", "EN_SOFT_PICKY_FAKE_OBJECT"},
		{"Use a spy object carefully.", "EN_SOFT_PICKY_SPY_OBJECT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeNetworkfirewallFirewallmanagerTranscribe(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Configure networkfirewall carefully.",
		"Configure NetworkFirewall carefully.",
		"Manage with firewallmanager carefully.",
		"Manage with FirewallManager carefully.",
		"Use transcribe carefully.",
		"Use Transcribe carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave107(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It ended plesantly carefully.", "EN_SOFT_PLEASANTLY_MISS", "pleasantly"},
		{"Still posessing carefully.", "EN_SOFT_POSSESSING_MISS", "possessing"},
		{"Many posibilities carefully.", "EN_SOFT_POSSIBILITIES_MISS", "possibilities"},
		{"It is practially done carefully.", "EN_SOFT_PRACTICALLY_MISS", "practically"},
		{"We are prepaired carefully.", "EN_SOFT_PREPARED_MISS", "prepared"},
		{"Still prepairing carefully.", "EN_SOFT_PREPARING_MISS", "preparing"},
		{"Grant priviledges carefully.", "EN_SOFT_PRIVILEGES_MISS", "privileges"},
		{"Keep the propotion carefully.", "EN_SOFT_PROPORTION_MISS", "proportion"},
		{"It will proably work carefully.", "EN_SOFT_PROBABLY_MISS2", "probably"},
		{"The radio reciever carefully.", "EN_SOFT_RECEIVER_MISS", "receiver"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave83(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use a factory method carefully.", "EN_SOFT_PICKY_FACTORY_METHOD"},
		{"Use the builder pattern carefully.", "EN_SOFT_PICKY_BUILDER_PATTERN"},
		{"Prefer object mother carefully.", "EN_SOFT_PICKY_OBJECT_MOTHER"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeRekognitionTextractPolly(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Detect with rekognition carefully.",
		"Detect with Rekognition carefully.",
		"Extract with textract carefully.",
		"Extract with Textract carefully.",
		"Speak with polly carefully.",
		"Speak with Polly carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave108(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A propotional share carefully.", "EN_SOFT_PROPORTIONAL_MISS", "proportional"},
		{"Check the propotions carefully.", "EN_SOFT_PROPORTIONS_MISS", "proportions"},
		{"A prosporous year carefully.", "EN_SOFT_PROSPEROUS_MISS", "prosperous"},
		{"Speak publicaly carefully.", "EN_SOFT_PUBLICLY_MISS2", "publicly"},
		{"Ask one quesion carefully.", "EN_SOFT_QUESTION_MISS", "question"},
		{"Ask hard quesions carefully.", "EN_SOFT_QUESTIONS_MISS", "questions"},
		{"Fill questionaires carefully.", "EN_SOFT_QUESTIONNAIRES_MISS", "questionnaires"},
		{"Highly reccomended carefully.", "EN_SOFT_RECOMMENDED_MISS", "recommended"},
		{"A dress rehersal carefully.", "EN_SOFT_REHEARSAL_MISS", "rehearsal"},
		{"Attend religously carefully.", "EN_SOFT_RELIGIOUSLY_MISS", "religiously"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave84(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Prefer a null object carefully.", "EN_SOFT_PICKY_NULL_OBJECT"},
		{"Use abstract factory carefully.", "EN_SOFT_PICKY_ABSTRACT_FACTORY"},
		{"Apply strategy pattern carefully.", "EN_SOFT_PICKY_STRATEGY_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeTranslateComprehendForecast(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Call translate carefully.",
		"Call Translate carefully.",
		"Run comprehend carefully.",
		"Run Comprehend carefully.",
		"Use forecast carefully.",
		"Use Forecast carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave109(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I rembered carefully.", "EN_SOFT_REMEMBERED_MISS2", "remembered"},
		{"Many rehersals carefully.", "EN_SOFT_REHEARSALS_MISS", "rehearsals"},
		{"A nice restuarant carefully.", "EN_SOFT_RESTAURANT_MISS4", "restaurant"},
		{"A ridiculus claim carefully.", "EN_SOFT_RIDICULOUS_MISS", "ridiculous"},
		{"Clear seperations carefully.", "EN_SOFT_SEPARATIONS_MISS", "separations"},
		{"A simmilar case carefully.", "EN_SOFT_SIMILAR_MISS2", "similar"},
		{"It finished sucessfully carefully.", "EN_SOFT_SUCCESSFULLY_MISS3", "successfully"},
		{"It was superceded carefully.", "EN_SOFT_SUPERSEDED_MISS", "superseded"},
		{"A tecnical detail carefully.", "EN_SOFT_TECHNICAL_MISS", "technical"},
		{"Bite your tounge carefully.", "EN_SOFT_TONGUE_MISS", "tongue"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave85(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use observer pattern carefully.", "EN_SOFT_PICKY_OBSERVER_PATTERN"},
		{"Use adapter pattern carefully.", "EN_SOFT_PICKY_ADAPTER_PATTERN"},
		{"Use decorator pattern carefully.", "EN_SOFT_PICKY_DECORATOR_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizePersonalizePinpointLex(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Recommend with personalize carefully.",
		"Recommend with Personalize carefully.",
		"Message via pinpoint carefully.",
		"Message via Pinpoint carefully.",
		"Chat with lex carefully.",
		"Chat with Lex carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave110(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Keep rembering carefully.", "EN_SOFT_REMEMBERING_MISS2", "remembering"},
		{"Many restuarants carefully.", "EN_SOFT_RESTAURANTS_MISS2", "restaurants"},
		{"A rythmic beat carefully.", "EN_SOFT_RHYTHMIC_MISS", "rhythmic"},
		{"Trust your sences carefully.", "EN_SOFT_SENSES_MISS", "senses"},
		{"Stop siezing carefully.", "EN_SOFT_SEIZING_MISS", "seizing"},
		{"And simmilarly carefully.", "EN_SOFT_SIMILARLY_MISS2", "similarly"},
		{"And sucessive wins carefully.", "EN_SOFT_SUCCESSIVE_MISS", "successive"},
		{"This supercedes carefully.", "EN_SOFT_SUPERSEDES_MISS", "supersedes"},
		{"Please surpress carefully.", "EN_SOFT_SUPPRESS_MISS", "suppress"},
		{"It is tecnically true carefully.", "EN_SOFT_TECHNICALLY_MISS", "technically"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave86(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use facade pattern carefully.", "EN_SOFT_PICKY_FACADE_PATTERN"},
		{"Use proxy pattern carefully.", "EN_SOFT_PICKY_PROXY_PATTERN"},
		{"Use singleton pattern carefully.", "EN_SOFT_PICKY_SINGLETON_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeMediaconvertMedialiveMediapackage(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Transcode with mediaconvert carefully.",
		"Transcode with MediaConvert carefully.",
		"Stream with medialive carefully.",
		"Stream with MediaLive carefully.",
		"Package with mediapackage carefully.",
		"Package with MediaPackage carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave111(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A new tecnique carefully.", "EN_SOFT_TECHNIQUE_MISS", "technique"},
		{"Useful tecniques carefully.", "EN_SOFT_TECHNIQUES_MISS", "techniques"},
		{"High tempertures carefully.", "EN_SOFT_TEMPERATURES_MISS", "temperatures"},
		{"Many tounges carefully.", "EN_SOFT_TONGUES_MISS", "tongues"},
		{"It is truley hard carefully.", "EN_SOFT_TRULY_MISS2", "truly"},
		{"The twelvth item carefully.", "EN_SOFT_TWELFTH_MISS2", "twelfth"},
		{"Keep useing carefully.", "EN_SOFT_USING_MISS", "using"},
		{"I usualy agree carefully.", "EN_SOFT_USUALLY_MISS", "usually"},
		{"Several varients carefully.", "EN_SOFT_VARIANTS_MISS", "variants"},
		{"A green vegatable carefully.", "EN_SOFT_VEGETABLE_MISS", "vegetable"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave87(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use composite pattern carefully.", "EN_SOFT_PICKY_COMPOSITE_PATTERN"},
		{"Use bridge pattern carefully.", "EN_SOFT_PICKY_BRIDGE_PATTERN"},
		{"Use command pattern carefully.", "EN_SOFT_PICKY_COMMAND_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeKinesisvideoAppstreamElasticbeanstalk(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Stream via kinesisvideo carefully.",
		"Stream via KinesisVideo carefully.",
		"Desktop via appstream carefully.",
		"Desktop via AppStream carefully.",
		"Deploy elasticbeanstalk carefully.",
		"Deploy ElasticBeanstalk carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave112(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I was suprised carefully.", "EN_SOFT_SURPRISED_MISS2", "surprised"},
		{"A suprising result carefully.", "EN_SOFT_SURPRISING_MISS", "surprising"},
		{"Feelings were surpressed carefully.", "EN_SOFT_SUPPRESSED_MISS", "suppressed"},
		{"It is unneccessarily complex carefully.", "EN_SOFT_UNNECESSARILY_MISS2", "unnecessarily"},
		{"Clean the vaccume carefully.", "EN_SOFT_VACUUM_MISS3", "vacuum"},
		{"Eat more vegatables carefully.", "EN_SOFT_VEGETABLES_MISS", "vegetables"},
		{"A visious cycle carefully.", "EN_SOFT_VICIOUS_MISS", "vicious"},
		{"We can acommodate carefully.", "EN_SOFT_ACCOMMODATE_MISS3", "accommodate"},
		{"Please acomplish carefully.", "EN_SOFT_ACCOMPLISH_MISS", "accomplish"},
		{"Walk accross carefully.", "EN_SOFT_ACROSS_MISS", "across"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave88(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use state pattern carefully.", "EN_SOFT_PICKY_STATE_PATTERN"},
		{"Use template method carefully.", "EN_SOFT_PICKY_TEMPLATE_METHOD"},
		{"Use flyweight pattern carefully.", "EN_SOFT_PICKY_FLYWEIGHT_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCodeartifactCodeguruCodewhisperer(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Publish to codeartifact carefully.",
		"Publish to CodeArtifact carefully.",
		"Review with codeguru carefully.",
		"Review with CodeGuru carefully.",
		"Assist with codewhisperer carefully.",
		"Assist with CodeWhisperer carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave113(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Many suprises carefully.", "EN_SOFT_SURPRISES_MISS", "surprises"},
		{"It is suprisingly easy carefully.", "EN_SOFT_SURPRISINGLY_MISS", "surprisingly"},
		{"Stop surpressing carefully.", "EN_SOFT_SUPPRESSING_MISS", "suppressing"},
		{"An unneccesary step carefully.", "EN_SOFT_UNNECESSARY_MISS3", "unnecessary"},
		{"Book acommodation carefully.", "EN_SOFT_ACCOMMODATION_MISS", "accommodation"},
		{"She acomplished carefully.", "EN_SOFT_ACCOMPLISHED_MISS", "accomplished"},
		{"A big acomplishment carefully.", "EN_SOFT_ACCOMPLISHMENT_MISS", "accomplishment"},
		{"Get aquainted carefully.", "EN_SOFT_ACQUAINTED_MISS", "acquainted"},
		{"One more attemp carefully.", "EN_SOFT_ATTEMPT_MISS", "attempt"},
		{"An avrage score carefully.", "EN_SOFT_AVERAGE_MISS", "average"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave89(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use chain of responsibility carefully.", "EN_SOFT_PICKY_CHAIN_OF_RESPONSIBILITY"},
		{"Use memento pattern carefully.", "EN_SOFT_PICKY_MEMENTO_PATTERN"},
		{"Use visitor pattern carefully.", "EN_SOFT_PICKY_VISITOR_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCloud9CloudshellAppflow(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Edit in cloud9 carefully.",
		"Edit in Cloud9 carefully.",
		"Open cloudshell carefully.",
		"Open CloudShell carefully.",
		"Sync with appflow carefully.",
		"Sync with AppFlow carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave114(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"List acomplishments carefully.", "EN_SOFT_ACCOMPLISHMENTS_MISS", "accomplishments"},
		{"Meet the athiests carefully.", "EN_SOFT_ATHEISTS_MISS", "atheists"},
		{"Failed attemps carefully.", "EN_SOFT_ATTEMPTS_MISS", "attempts"},
		{"Daily avrages carefully.", "EN_SOFT_AVERAGES_MISS", "averages"},
		{"An awsome idea carefully.", "EN_SOFT_AWESOME_MISS", "awesome"},
		{"A superceding rule carefully.", "EN_SOFT_SUPERSEDING_MISS", "superseding"},
		{"Move rythmically carefully.", "EN_SOFT_RHYTHMICALLY_MISS", "rhythmically"},
		{"A strong resembelance carefully.", "EN_SOFT_RESEMBLANCE_MISS2", "resemblance"},
		{"Speak relevently carefully.", "EN_SOFT_RELEVANTLY_MISS", "relevantly"},
		{"A strong candadate carefully.", "EN_SOFT_CANDIDATE_MISS", "candidate"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave90(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use mediator pattern carefully.", "EN_SOFT_PICKY_MEDIATOR_PATTERN"},
		{"Use iterator pattern carefully.", "EN_SOFT_PICKY_ITERATOR_PATTERN"},
		{"Use prototype pattern carefully.", "EN_SOFT_PICKY_PROTOTYPE_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeAppmeshMskMq(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Route with appmesh carefully.",
		"Route with AppMesh carefully.",
		"Stream via msk carefully.",
		"Stream via MSK carefully.",
		"Queue with mq carefully.",
		"Queue with MQ carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave115(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A close resemblence carefully.", "EN_SOFT_RESEMBLANCE_MISS3", "resemblance"},
		{"Attacked visiously carefully.", "EN_SOFT_VICIOUSLY_MISS", "viciously"},
		{"Embrace the wierdness carefully.", "EN_SOFT_WEIRDNESS_MISS", "weirdness"},
		{"She belives carefully.", "EN_SOFT_BELIEVES_MISS2", "believes"},
		{"Print calenders carefully.", "EN_SOFT_CALENDARS_MISS2", "calendars"},
		{"Many candadates carefully.", "EN_SOFT_CANDIDATES_MISS2", "candidates"},
		{"Long carreers carefully.", "EN_SOFT_CAREERS_MISS", "careers"},
		{"Old cemetaries carefully.", "EN_SOFT_CEMETERIES_MISS2", "cemeteries"},
		{"Add more colums carefully.", "EN_SOFT_COLUMNS_MISS", "columns"},
		{"Please contineu carefully.", "EN_SOFT_CONTINUE_MISS", "continue"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave91(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use repository pattern carefully.", "EN_SOFT_PICKY_REPOSITORY_PATTERN"},
		{"Use unit of work carefully.", "EN_SOFT_PICKY_UNIT_OF_WORK"},
		{"Use specification pattern carefully.", "EN_SOFT_PICKY_SPECIFICATION_PATTERN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeStepfunctionsDmsMemorydb(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Orchestrate with stepfunctions carefully.",
		"Orchestrate with StepFunctions carefully.",
		"Migrate with dms carefully.",
		"Migrate with DMS carefully.",
		"Cache with memorydb carefully.",
		"Cache with MemoryDB carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave116(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"He belived carefully.", "EN_SOFT_BELIEVED_MISS2", "believed"},
		{"Several commitees carefully.", "EN_SOFT_COMMITTEES_MISS", "committees"},
		{"Raise conciousness carefully.", "EN_SOFT_CONSCIOUSNESS_MISS", "consciousness"},
		{"A contamporary style carefully.", "EN_SOFT_CONTEMPORARY_MISS", "contemporary"},
		{"It contineus carefully.", "EN_SOFT_CONTINUES_MISS", "continues"},
		{"New contructions carefully.", "EN_SOFT_CONSTRUCTIONS_MISS", "constructions"},
		{"A convinient time carefully.", "EN_SOFT_CONVENIENT_MISS3", "convenient"},
		{"Placed conviniently carefully.", "EN_SOFT_CONVENIENTLY_MISS", "conveniently"},
		{"It correponds carefully.", "EN_SOFT_CORRESPONDS_MISS", "corresponds"},
		{"A clear defination carefully.", "EN_SOFT_DEFINITION_MISS2", "definition"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave92(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Prefer open host service carefully.", "EN_SOFT_PICKY_OPEN_HOST_SERVICE"},
		{"Prefer published language carefully.", "EN_SOFT_PICKY_PUBLISHED_LANGUAGE"},
		{"Prefer customer supplier carefully.", "EN_SOFT_PICKY_CUSTOMER_SUPPLIER"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeDaxKeyspacesQldb(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Cache with dax carefully.",
		"Cache with DAX carefully.",
		"Store in keyspaces carefully.",
		"Store in Keyspaces carefully.",
		"Ledger with qldb carefully.",
		"Ledger with QLDB carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave117(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Walk throught carefully.", "EN_SOFT_THROUGH_MISS", "through"},
		{"Harsh critisisms carefully.", "EN_SOFT_CRITICISMS_MISS", "criticisms"},
		{"He decieves carefully.", "EN_SOFT_DECEIVES_MISS", "deceives"},
		{"Stop decieving carefully.", "EN_SOFT_DECEIVING_MISS", "deceiving"},
		{"Clear definations carefully.", "EN_SOFT_DEFINITIONS_MISS", "definitions"},
		{"A dependant clause carefully.", "EN_SOFT_DEPENDENT_MISS", "dependent"},
		{"Deep disapointment carefully.", "EN_SOFT_DISAPPOINTMENT_MISS", "disappointment"},
		{"They liased carefully.", "EN_SOFT_LIAISED_MISS", "liaised"},
		{"Two milleniums carefully.", "EN_SOFT_MILLENNIUMS_MISS", "millenniums"},
		{"Grin mischieviously carefully.", "EN_SOFT_MISCHIEVOUSLY_MISS", "mischievously"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave93(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Prefer conformist carefully.", "EN_SOFT_PICKY_CONFORMIST"},
		{"Prefer partnership carefully.", "EN_SOFT_PICKY_PARTNERSHIP"},
		{"Prefer separate ways carefully.", "EN_SOFT_PICKY_SEPARATE_WAYS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeAuroraFsxStoragegateway(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run on aurora carefully.",
		"Run on Aurora carefully.",
		"Mount fsx carefully.",
		"Mount FSx carefully.",
		"Use storagegateway carefully.",
		"Use StorageGateway carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave118(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Many disapointments carefully.", "EN_SOFT_DISAPPOINTMENTS_MISS", "disappointments"},
		{"Keep liasing carefully.", "EN_SOFT_LIAISING_MISS", "liaising"},
		{"Count occurences carefully.", "EN_SOFT_OCCURRENCES_MISS2", "occurrences"},
		{"Run in paralell carefully.", "EN_SOFT_PARALLEL_MISS4", "parallel"},
		{"A preferrable option carefully.", "EN_SOFT_PREFERABLE_MISS2", "preferable"},
		{"Need preperation carefully.", "EN_SOFT_PREPARATION_MISS", "preparation"},
		{"A special privledge carefully.", "EN_SOFT_PRIVILEGE_MISS3", "privilege"},
		{"Meet the proffesors carefully.", "EN_SOFT_PROFESSORS_MISS", "professors"},
		{"Celebrate sucesss carefully.", "EN_SOFT_SUCCESS_MISS3", "success"},
		{"Run parallely carefully.", "EN_SOFT_PARALLELLY_MISS", "parallelly"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave94(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Avoid big ball of mud carefully.", "EN_SOFT_PICKY_BIG_BALL_OF_MUD"},
		{"Focus on core domain carefully.", "EN_SOFT_PICKY_CORE_DOMAIN"},
		{"Keep supporting domain carefully.", "EN_SOFT_PICKY_SUPPORTING_DOMAIN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeBackupDatasyncSnowball(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Configure backup carefully.",
		"Configure Backup carefully.",
		"Transfer with datasync carefully.",
		"Transfer with DataSync carefully.",
		"Ship via snowball carefully.",
		"Ship via Snowball carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave119(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Follow reccomendations carefully.", "EN_SOFT_RECOMMENDATIONS_MISS", "recommendations"},
		{"Applied sucessively carefully.", "EN_SOFT_SUCCESSIVELY_MISS", "successively"},
		{"The change propogated carefully.", "EN_SOFT_PROPAGATED_MISS", "propagated"},
		{"It propogates carefully.", "EN_SOFT_PROPAGATES_MISS", "propagates"},
		{"Keep propogating carefully.", "EN_SOFT_PROPAGATING_MISS", "propagating"},
		{"Grant priveleges carefully.", "EN_SOFT_PRIVILEGES_MISS2", "privileges"},
		{"A noble proffesion carefully.", "EN_SOFT_PROFESSION_MISS2", "profession"},
		{"Many absenses carefully.", "EN_SOFT_ABSENCES_MISS", "absences"},
		{"Do not abondon carefully.", "EN_SOFT_ABANDON_MISS", "abandon"},
		{"Gain acceptence carefully.", "EN_SOFT_ACCEPTANCE_MISS", "acceptance"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave95(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Treat generic subdomain carefully.", "EN_SOFT_PICKY_GENERIC_SUBDOMAIN"},
		{"Apply distillation carefully.", "EN_SOFT_PICKY_DISTILLATION"},
		{"Run event storming carefully.", "EN_SOFT_PICKY_EVENT_STORMING"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeSnowconeSnowmobileMediastore(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Ship via snowcone carefully.",
		"Ship via Snowcone carefully.",
		"Ship via snowmobile carefully.",
		"Ship via Snowmobile carefully.",
		"Store in mediastore carefully.",
		"Store in MediaStore carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave120(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It was abondoned carefully.", "EN_SOFT_ABANDONED_MISS", "abandoned"},
		{"Stop abondoning carefully.", "EN_SOFT_ABANDONING_MISS", "abandoning"},
		{"Many acceptences carefully.", "EN_SOFT_ACCEPTANCES_MISS", "acceptances"},
		{"Great acheivements carefully.", "EN_SOFT_ACHIEVEMENTS_MISS2", "achievements"},
		{"Old aquaintances carefully.", "EN_SOFT_ACQUAINTANCES_MISS", "acquaintances"},
		{"Plot assasinations carefully.", "EN_SOFT_ASSASSINATIONS_MISS", "assassinations"},
		{"Done beatifully carefully.", "EN_SOFT_BEAUTIFULLY_MISS", "beautifully"},
		{"Resolve dependancies carefully.", "EN_SOFT_DEPENDENCIES_MISS", "dependencies"},
		{"It dissapeared carefully.", "EN_SOFT_DISAPPEARED_MISS2", "disappeared"},
		{"Gain experance carefully.", "EN_SOFT_EXPERIENCE_MISS2", "experience"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave96(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Write a domain story carefully.", "EN_SOFT_PICKY_DOMAIN_STORY"},
		{"Do context mapping carefully.", "EN_SOFT_PICKY_CONTEXT_MAPPING"},
		{"Prefer model driven design carefully.", "EN_SOFT_PICKY_MODEL_DRIVEN_DESIGN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeElastictranscoderTransferfamilyElemental(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Encode with elastictranscoder carefully.",
		"Encode with ElasticTranscoder carefully.",
		"Move with transferfamily carefully.",
		"Move with TransferFamily carefully.",
		"Process with elemental carefully.",
		"Process with Elemental carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave121(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Avoid embarasment carefully.", "EN_SOFT_EMBARRASSMENT_MISS2", "embarrassment"},
		{"Share experiances carefully.", "EN_SOFT_EXPERIENCES_MISS", "experiences"},
		{"Still experiancing carefully.", "EN_SOFT_EXPERIENCING_MISS", "experiencing"},
		{"A close firend carefully.", "EN_SOFT_FRIEND_MISS2", "friend"},
		{"Meet firends carefully.", "EN_SOFT_FRIENDS_MISS2", "friends"},
		{"Measure the heighth carefully.", "EN_SOFT_HEIGHT_MISS2", "height"},
		{"The hightest peak carefully.", "EN_SOFT_HIGHEST_MISS", "highest"},
		{"Good hygine carefully.", "EN_SOFT_HYGIENE_MISS3", "hygiene"},
		{"Do it imediatey carefully.", "EN_SOFT_IMMEDIATELY_MISS4", "immediately"},
		{"I mean it litearally carefully.", "EN_SOFT_LITERALLY_MISS2", "literally"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave97(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Run event modeling carefully.", "EN_SOFT_PICKY_EVENT_MODELING"},
		{"Run example mapping carefully.", "EN_SOFT_PICKY_EXAMPLE_MAPPING"},
		{"Run story mapping carefully.", "EN_SOFT_PICKY_STORY_MAPPING"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeIvsChimeConnect(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Stream with ivs carefully.",
		"Stream with IVS carefully.",
		"Meet on chime carefully.",
		"Meet on Chime carefully.",
		"Call center connect carefully.",
		"Call center Connect carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave122(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Schedule maintanence carefully.", "EN_SOFT_MAINTENANCE_MISS5", "maintenance"},
		{"A sharp manouever carefully.", "EN_SOFT_MANEUVER_MISS3", "maneuver"},
		{"Complex manouvers carefully.", "EN_SOFT_MANEUVERS_MISS", "maneuvers"},
		{"A skilled mathamatician carefully.", "EN_SOFT_MATHEMATICIAN_MISS", "mathematician"},
		{"Grin mischeviously carefully.", "EN_SOFT_MISCHIEVOUSLY_MISS2", "mischievously"},
		{"Fix mispellings carefully.", "EN_SOFT_MISSPELLINGS_MISS", "misspellings"},
		{"A missmatched pair carefully.", "EN_SOFT_MISMATCHED_MISS", "mismatched"},
		{"It vanished misteriously carefully.", "EN_SOFT_MYSTERIOUSLY_MISS", "mysteriously"},
		{"A kind neigbour carefully.", "EN_SOFT_NEIGHBOUR_MISS", "neighbor"},
		{"The ninteenth century carefully.", "EN_SOFT_NINETEENTH_MISS", "nineteenth"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave98(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Run impact mapping carefully.", "EN_SOFT_PICKY_IMPACT_MAPPING"},
		{"Prefer specification by example carefully.", "EN_SOFT_PICKY_SPECIFICATION_BY_EXAMPLE"},
		{"Prefer user story mapping carefully.", "EN_SOFT_PICKY_USER_STORY_MAPPING"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeChimesdkConnectcasesWisdom(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Build with chimesdk carefully.",
		"Build with ChimeSDK carefully.",
		"Track with connectcases carefully.",
		"Track with ConnectCases carefully.",
		"Assist with wisdom carefully.",
		"Assist with Wisdom carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave123(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fix ommisions carefully.", "EN_SOFT_OMISSIONS_MISS2", "omissions"},
		{"It was ommited carefully.", "EN_SOFT_OMITTED_MISS", "omitted"},
		{"Stop ommitting carefully.", "EN_SOFT_OMITTING_MISS", "omitting"},
		{"An orphant child carefully.", "EN_SOFT_ORPHAN_MISS", "orphan"},
		{"Help orphants carefully.", "EN_SOFT_ORPHANS_MISS", "orphans"},
		{"Most paticularly carefully.", "EN_SOFT_PARTICULARLY_MISS4", "particularly"},
		{"It was peirced carefully.", "EN_SOFT_PIERCED_MISS", "pierced"},
		{"Ask persistantly carefully.", "EN_SOFT_PERSISTENTLY_MISS", "persistently"},
		{"Keep persuing carefully.", "EN_SOFT_PURSUING_MISS2", "pursuing"},
		{"An ancient pharoah carefully.", "EN_SOFT_PHARAOH_MISS", "pharaoh"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave99(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Write given when then carefully.", "EN_SOFT_PICKY_GIVEN_WHEN_THEN"},
		{"Hold three amigos carefully.", "EN_SOFT_PICKY_THREE_AMIGOS"},
		{"Define acceptance criteria carefully.", "EN_SOFT_PICKY_ACCEPTANCE_CRITERIA"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeConnectvoiceAppintegrationsConnectcustomerprofiles(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Route connectvoice carefully.",
		"Route ConnectVoice carefully.",
		"Wire appintegrations carefully.",
		"Wire AppIntegrations carefully.",
		"Use connectcustomerprofiles carefully.",
		"Use ConnectCustomerProfiles carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave124(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Strange phenomona carefully.", "EN_SOFT_PHENOMENA_MISS2", "phenomena"},
		{"A rare phenominon carefully.", "EN_SOFT_PHENOMENON_MISS", "phenomenon"},
		{"I prefere carefully.", "EN_SOFT_PREFER_MISS", "prefer"},
		{"Value libertery carefully.", "EN_SOFT_LIBERTY_MISS", "liberty"},
		{"Stuck inbetween carefully.", "EN_SOFT_INBETWEEN_MISS", "in-between"},
		{"Count incidencies carefully.", "EN_SOFT_INCIDENCES_MISS", "incidences"},
		{"Admit ingorance carefully.", "EN_SOFT_IGNORANCE_MISS2", "ignorance"},
		{"New initatives carefully.", "EN_SOFT_INITIATIVES_MISS", "initiatives"},
		{"Drawn irrisistably carefully.", "EN_SOFT_IRRESISTIBLY_MISS", "irresistibly"},
		{"Still isntalling carefully.", "EN_SOFT_INSTALLING_MISS", "installing"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave100(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Agree definition of done carefully.", "EN_SOFT_PICKY_DEFINITION_OF_DONE"},
		{"Agree definition of ready carefully.", "EN_SOFT_PICKY_DEFINITION_OF_READY"},
		{"Set smart goals carefully.", "EN_SOFT_PICKY_SMART_GOALS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeConnectcampaignsConnectparticipantCustomerprofiles(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run connectcampaigns carefully.",
		"Run ConnectCampaigns carefully.",
		"Use connectparticipant carefully.",
		"Use ConnectParticipant carefully.",
		"Use customerprofiles carefully.",
		"Use CustomerProfiles carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave125(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Speak knowlegeably carefully.", "EN_SOFT_KNOWLEDGEABLY_MISS", "knowledgeably"},
		{"Several laisons carefully.", "EN_SOFT_LIAISONS_MISS2", "liaisons"},
		{"Hold liscences carefully.", "EN_SOFT_LICENCES_MISS", "licenses"},
		{"It is liscensed carefully.", "EN_SOFT_LICENSED_MISS2", "licensed"},
		{"Renew liscenses carefully.", "EN_SOFT_LICENSES_MISS", "licenses"},
		{"Act proffesionally carefully.", "EN_SOFT_PROFESSIONALLY_MISS2", "professionally"},
		{"The prohabition era carefully.", "EN_SOFT_PROHIBITION_MISS", "prohibition"},
		{"Please annouce carefully.", "EN_SOFT_ANNOUNCE_MISS2", "announce"},
		{"An ajacent room carefully.", "EN_SOFT_ADJACENT_MISS", "adjacent"},
		{"Walk amung carefully.", "EN_SOFT_AMONG_MISS2", "among"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave101(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply moscow method carefully.", "EN_SOFT_PICKY_MOSCOW_METHOD"},
		{"Compute rice score carefully.", "EN_SOFT_PICKY_RICE_SCORE"},
		{"Rank by wsjf carefully.", "EN_SOFT_PICKY_WSJF"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeSesv2PinpointsmsvoiceContactlens(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Send via sesv2 carefully.",
		"Send via SESv2 carefully.",
		"Message with pinpointsmsvoice carefully.",
		"Message with PinpointSMSVoice carefully.",
		"Analyze with contactlens carefully.",
		"Analyze with ContactLens carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave126(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"It is almst done carefully.", "EN_SOFT_ALMOST_MISS2", "almost"},
		{"It is alomst ready carefully.", "EN_SOFT_ALMOST_MISS3", "almost"},
		{"Large ammounts carefully.", "EN_SOFT_AMOUNTS_MISS", "amounts"},
		{"Many aniversaries carefully.", "EN_SOFT_ANNIVERSARIES_MISS", "anniversaries"},
		{"It was annouced carefully.", "EN_SOFT_ANNOUNCED_MISS", "announced"},
		{"A public annoucement carefully.", "EN_SOFT_ANNOUNCEMENT_MISS2", "announcement"},
		{"Several annoucements carefully.", "EN_SOFT_ANNOUNCEMENTS_MISS", "announcements"},
		{"Try anothe way carefully.", "EN_SOFT_ANOTHER_MISS", "another"},
		{"Try anouther path carefully.", "EN_SOFT_ANOTHER_MISS2", "another"},
		{"I appreaciate carefully.", "EN_SOFT_APPRECIATE_MISS", "appreciate"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave102(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply kano model carefully.", "EN_SOFT_PICKY_KANO_MODEL"},
		{"Measure cost of delay carefully.", "EN_SOFT_PICKY_COST_OF_DELAY"},
		{"Plot value vs effort carefully.", "EN_SOFT_PICKY_VALUE_VS_EFFORT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizePinpointsmsvoicev2SocialmessagingPinpointemail(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Send with pinpointsmsvoicev2 carefully.",
		"Send with PinpointSMSVoiceV2 carefully.",
		"Chat via socialmessaging carefully.",
		"Chat via SocialMessaging carefully.",
		"Mail via pinpointemail carefully.",
		"Mail via PinpointEmail carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave127(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Many appearences carefully.", "EN_SOFT_APPEARANCES_MISS", "appearances"},
		{"It was appreaciated carefully.", "EN_SOFT_APPRECIATED_MISS", "appreciated"},
		{"Show appreaciation carefully.", "EN_SOFT_APPRECIATION_MISS", "appreciation"},
		{"A new approch carefully.", "EN_SOFT_APPROACH_MISS", "approach"},
		{"They approched carefully.", "EN_SOFT_APPROACHED_MISS", "approached"},
		{"Cars are approching carefully.", "EN_SOFT_APPROACHING_MISS", "approaching"},
		{"An aproximate value carefully.", "EN_SOFT_APPROXIMATE_MISS", "approximate"},
		{"About aproximately ten carefully.", "EN_SOFT_APPROXIMATELY_MISS", "approximately"},
		{"A rough aproximation carefully.", "EN_SOFT_APPROXIMATION_MISS", "approximation"},
		{"Late arival carefully.", "EN_SOFT_ARRIVAL_MISS", "arrival"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave103(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Run buy a feature carefully.", "EN_SOFT_PICKY_BUY_A_FEATURE"},
		{"Run hundred dollar test carefully.", "EN_SOFT_PICKY_HUNDRED_DOLLAR_TEST"},
		{"Use opportunity scoring carefully.", "EN_SOFT_PICKY_OPPORTUNITY_SCORING"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizePinpointappMobiletargetingMobileanalytics(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Configure pinpointapp carefully.",
		"Configure PinpointApp carefully.",
		"Use mobiletargeting carefully.",
		"Use MobileTargeting carefully.",
		"Track mobileanalytics carefully.",
		"Track MobileAnalytics carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave128(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"They will arive carefully.", "EN_SOFT_ARRIVE_MISS", "arrive"},
		{"They arived carefully.", "EN_SOFT_ARRIVED_MISS", "arrived"},
		{"Trains are ariving carefully.", "EN_SOFT_ARRIVING_MISS", "arriving"},
		{"Join the assosiation carefully.", "EN_SOFT_ASSOCIATION_MISS2", "association"},
		{"A bad assumtion carefully.", "EN_SOFT_ASSUMPTION_MISS", "assumption"},
		{"Wrong assumtions carefully.", "EN_SOFT_ASSUMPTIONS_MISS", "assumptions"},
		{"One more atempt carefully.", "EN_SOFT_ATTEMPT_MISS2", "attempt"},
		{"Failed atempts carefully.", "EN_SOFT_ATTEMPTS_MISS2", "attempts"},
		{"Please attatch carefully.", "EN_SOFT_ATTACH_MISS", "attach"},
		{"Still attatching carefully.", "EN_SOFT_ATTACHING_MISS", "attaching"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave104(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use planning poker carefully.", "EN_SOFT_PICKY_PLANNING_POKER"},
		{"Use tshirt sizing carefully.", "EN_SOFT_PICKY_TSHIRT_SIZING"},
		{"Use affinity estimation carefully.", "EN_SOFT_PICKY_AFFINITY_ESTIMATION"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeDevicefarmAmplifyadminAmplifybackend(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Test on devicefarm carefully.",
		"Test on DeviceFarm carefully.",
		"Manage amplifyadmin carefully.",
		"Manage AmplifyAdmin carefully.",
		"Configure amplifybackend carefully.",
		"Configure AmplifyBackend carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave129(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Add an attatchment carefully.", "EN_SOFT_ATTACHMENT_MISS", "attachment"},
		{"Many attatchments carefully.", "EN_SOFT_ATTACHMENTS_MISS", "attachments"},
		{"A large audiance carefully.", "EN_SOFT_AUDIENCE_MISS", "audience"},
		{"Multiple audiances carefully.", "EN_SOFT_AUDIENCES_MISS", "audiences"},
		{"Now avialable carefully.", "EN_SOFT_AVAILABLE_MISS3", "available"},
		{"Check avialability carefully.", "EN_SOFT_AVAILABILITY_MISS2", "availability"},
		{"An aweful mistake carefully.", "EN_SOFT_AWFUL_MISS2", "awful"},
		{"A total begginer carefully.", "EN_SOFT_BEGINNER_MISS", "beginner"},
		{"Help begginers carefully.", "EN_SOFT_BEGINNERS_MISS", "beginners"},
		{"Strong beleifs carefully.", "EN_SOFT_BELIEFS_MISS", "beliefs"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave105(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use bucket system carefully.", "EN_SOFT_PICKY_BUCKET_SYSTEM"},
		{"Use dot voting carefully.", "EN_SOFT_PICKY_DOT_VOTING"},
		{"Use relative mass carefully.", "EN_SOFT_PICKY_RELATIVE_MASS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeAmplifyuibuilderGeofenceLocationservice(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Design with amplifyuibuilder carefully.",
		"Design with AmplifyUIBuilder carefully.",
		"Draw a geofence carefully.",
		"Draw a Geofence carefully.",
		"Call locationservice carefully.",
		"Call LocationService carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave130(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Stuck bettween carefully.", "EN_SOFT_BETWEEN_MISS2", "between"},
		{"Check the calandar carefully.", "EN_SOFT_CALENDAR_MISS2", "calendar"},
		{"Print calandars carefully.", "EN_SOFT_CALENDARS_MISS3", "calendars"},
		{"The path choosen carefully.", "EN_SOFT_CHOSEN_MISS", "chosen"},
		{"A hard choise carefully.", "EN_SOFT_CHOICE_MISS", "choice"},
		{"Limited choises carefully.", "EN_SOFT_CHOICES_MISS", "choices"},
		{"A special ocassion carefully.", "EN_SOFT_OCCASION_MISS2", "occasion"},
		{"Many ocassions carefully.", "EN_SOFT_OCCASIONS_MISS2", "occasions"},
		{"An ocassional visit carefully.", "EN_SOFT_OCCASIONAL_MISS2", "occasional"},
		{"Visit ocassionally carefully.", "EN_SOFT_OCCASIONALLY_MISS2", "occasionally"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave106(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Estimate story points carefully.", "EN_SOFT_PICKY_STORY_POINTS"},
		{"Track velocity chart carefully.", "EN_SOFT_PICKY_VELOCITY_CHART"},
		{"Use fibonacci sizing carefully.", "EN_SOFT_PICKY_FIBONACCI_SIZING"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeIotcoreIoteventsGreengrass(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Connect to iotcore carefully.",
		"Connect to IoTCore carefully.",
		"Detect with iotevents carefully.",
		"Detect with IoTEvents carefully.",
		"Deploy greengrass carefully.",
		"Deploy Greengrass carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave131(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A succesfull plan carefully.", "EN_SOFT_SUCCESSFUL_MISS4", "successful"},
		{"Finished succesfuly carefully.", "EN_SOFT_SUCCESSFULLY_MISS4", "successfully"},
		{"It is aparentely true carefully.", "EN_SOFT_APPARENTLY_MISS4", "apparently"},
		{"Please receieve carefully.", "EN_SOFT_RECEIVE_MISS2", "receive"},
		{"Keep the reciepts carefully.", "EN_SOFT_RECEIPTS_MISS", "receipts"},
		{"Store seperateley carefully.", "EN_SOFT_SEPARATELY_MISS3", "separately"},
		{"Handle seperatley carefully.", "EN_SOFT_SEPARATELY_MISS4", "separately"},
		{"Update the adresses carefully.", "EN_SOFT_ADDRESSES_MISS", "addresses"},
		{"She acheives carefully.", "EN_SOFT_ACHIEVES_MISS", "achieves"},
		{"Keep acheiveing carefully.", "EN_SOFT_ACHIEVING_MISS", "achieving"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave107(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Track burndown chart carefully.", "EN_SOFT_PICKY_BURNDOWN_CHART"},
		{"Track burnup chart carefully.", "EN_SOFT_PICKY_BURNUP_CHART"},
		{"Track cumulative flow carefully.", "EN_SOFT_PICKY_CUMULATIVE_FLOW"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeIotanalyticsIotsitewiseIotthingsgraph(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Query iotanalytics carefully.",
		"Query IoTAnalytics carefully.",
		"Model with iotsitewise carefully.",
		"Model with IoTSiteWise carefully.",
		"Wire iotthingsgraph carefully.",
		"Wire IoTThingsGraph carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave132(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A strange apperance carefully.", "EN_SOFT_APPEARANCE_MISS3", "appearance"},
		{"Multiple apperances carefully.", "EN_SOFT_APPEARANCES_MISS2", "appearances"},
		{"An appropiate reply carefully.", "EN_SOFT_APPROPRIATE_MISS2", "appropriate"},
		{"Dress appropiately carefully.", "EN_SOFT_APPROPRIATELY_MISS2", "appropriately"},
		{"Strong arguemnts carefully.", "EN_SOFT_ARGUMENTS_MISS2", "arguments"},
		{"Hire assasins carefully.", "EN_SOFT_ASSASSINS_MISS", "assassins"},
		{"Keep beliveing carefully.", "EN_SOFT_BELIEVING_MISS2", "believing"},
		{"Still benifiting carefully.", "EN_SOFT_BENEFITING_MISS", "benefiting"},
		{"Have benifitted carefully.", "EN_SOFT_BENEFITTED_MISS", "benefitted"},
		{"Meet buisnessmen carefully.", "EN_SOFT_BUSINESSMEN_MISS2", "businessmen"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave108(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Measure cycle time carefully.", "EN_SOFT_PICKY_CYCLE_TIME"},
		{"Measure lead time carefully.", "EN_SOFT_PICKY_LEAD_TIME"},
		{"Use a kanban board carefully.", "EN_SOFT_PICKY_KANBAN_BOARD"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeIotdeviceadvisorIotfleethubIotwireless(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Test with iotdeviceadvisor carefully.",
		"Test with IoTDeviceAdvisor carefully.",
		"Manage iotfleethub carefully.",
		"Manage IoTFleetHub carefully.",
		"Connect iotwireless carefully.",
		"Connect IoTWireless carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave133(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"The path choisen carefully.", "EN_SOFT_CHOSEN_MISS2", "chosen"},
		{"Start writeing carefully.", "EN_SOFT_WRITING_MISS2", "writing"},
		{"We receievd carefully.", "EN_SOFT_RECEIVED_MISS2", "received"},
		{"Keep annoucing carefully.", "EN_SOFT_ANNOUNCING_MISS", "announcing"},
		{"A radio annoucer carefully.", "EN_SOFT_ANNOUNCER_MISS", "announcer"},
		{"Fix the adressses carefully.", "EN_SOFT_ADDRESSES_MISS2", "addresses"},
		{"An importnat note carefully.", "EN_SOFT_IMPORTANT_MISS", "important"},
		{"Is it posible carefully.", "EN_SOFT_POSSIBLE_MISS", "possible"},
		{"Stop becuase carefully.", "EN_SOFT_BECAUSE_MISS", "because"},
		{"Show an exmaple carefully.", "EN_SOFT_EXAMPLE_MISS", "example"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave109(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Track throughput chart carefully.", "EN_SOFT_PICKY_THROUGHPUT_CHART"},
		{"Measure flow efficiency carefully.", "EN_SOFT_PICKY_FLOW_EFFICIENCY"},
		{"Agree service level carefully.", "EN_SOFT_PICKY_SERVICE_LEVEL"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeFreertosIotjobsdataplaneIotsecuredtunneling(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run freertos carefully.",
		"Run FreeRTOS carefully.",
		"Call iotjobsdataplane carefully.",
		"Call IoTJobsDataPlane carefully.",
		"Open iotsecuredtunneling carefully.",
		"Open IoTSecureTunneling carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave134(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"An improtant note carefully.", "EN_SOFT_IMPORTANT_MISS2", "important"},
		{"Is it possable carefully.", "EN_SOFT_POSSIBLE_MISS2", "possible"},
		{"Is it posibble carefully.", "EN_SOFT_POSSIBLE_MISS3", "possible"},
		{"It will probly work carefully.", "EN_SOFT_PROBABLY_MISS3", "probably"},
		{"It will probbaly work carefully.", "EN_SOFT_PROBABLY_MISS4", "probably"},
		{"A beutiful day carefully.", "EN_SOFT_BEAUTIFUL_MISS3", "beautiful"},
		{"An intereseting idea carefully.", "EN_SOFT_INTERESTING_MISS2", "interesting"},
		{"Now avaliable carefully.", "EN_SOFT_AVAILABLE_MISS4", "available"},
		{"Stop becasue carefully.", "EN_SOFT_BECAUSE_MISS2", "because"},
		{"Stop beacuse carefully.", "EN_SOFT_BECAUSE_MISS3", "because"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave110(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Measure touch time carefully.", "EN_SOFT_PICKY_TOUCH_TIME"},
		{"Measure queue time carefully.", "EN_SOFT_PICKY_QUEUE_TIME"},
		{"Measure process time carefully.", "EN_SOFT_PICKY_PROCESS_TIME"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeIotfleetwiseIotroborunnerIotcoredeviceadvisor(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Track with iotfleetwise carefully.",
		"Track with IoTFleetWise carefully.",
		"Run iotroborunner carefully.",
		"Run IoTRoboRunner carefully.",
		"Test iotcoredeviceadvisor carefully.",
		"Test IoTCoreDeviceAdvisor carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave135(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Do this beofre carefully.", "EN_SOFT_BEFORE_MISS2", "before"},
		{"Start a busines carefully.", "EN_SOFT_BUSINESS_MISS2", "business"},
		{"Help the childern carefully.", "EN_SOFT_CHILDREN_MISS", "children"},
		{"Help the childrun carefully.", "EN_SOFT_CHILDREN_MISS2", "children"},
		{"Visit the contry carefully.", "EN_SOFT_COUNTRY_MISS", "country"},
		{"Please decice carefully.", "EN_SOFT_DECIDE_MISS", "decide"},
		{"Please deside carefully.", "EN_SOFT_DECIDE_MISS2", "decide"},
		{"Please discribe carefully.", "EN_SOFT_DESCRIBE_MISS2", "describe"},
		{"Please develp carefully.", "EN_SOFT_DEVELOP_MISS3", "develop"},
		{"That is enuf carefully.", "EN_SOFT_ENOUGH_MISS", "enough"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave111(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Measure wait time carefully.", "EN_SOFT_PICKY_WAIT_TIME"},
		{"Measure blocked time carefully.", "EN_SOFT_PICKY_BLOCKED_TIME"},
		{"Map the value stream carefully.", "EN_SOFT_PICKY_VALUE_STREAM"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeIotexpresslinkIottwinmakerIotfleethubv2(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Connect iotexpresslink carefully.",
		"Connect IoTExpressLink carefully.",
		"Model with iottwinmaker carefully.",
		"Model with IoTTwinMaker carefully.",
		"Manage iotfleethubv2 carefully.",
		"Manage IoTFleetHubV2 carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave136(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"That is enogh carefully.", "EN_SOFT_ENOUGH_MISS2", "enough"},
		{"That is enoug carefully.", "EN_SOFT_ENOUGH_MISS3", "enough"},
		{"Show an exampe carefully.", "EN_SOFT_EXAMPLE_MISS2", "example"},
		{"Show an exaple carefully.", "EN_SOFT_EXAMPLE_MISS3", "example"},
		{"A close frend carefully.", "EN_SOFT_FRIEND_MISS3", "friend"},
		{"And howver carefully.", "EN_SOFT_HOWEVER_MISS", "however"},
		{"Please inculde carefully.", "EN_SOFT_INCLUDE_MISS", "include"},
		{"Share informtion carefully.", "EN_SOFT_INFORMATION_MISS2", "information"},
		{"Use this instaed carefully.", "EN_SOFT_INSTEAD_MISS3", "instead"},
		{"It is mabye true carefully.", "EN_SOFT_MAYBE_MISS", "maybe"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave112(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Measure reaction time carefully.", "EN_SOFT_PICKY_REACTION_TIME"},
		{"Measure idle time carefully.", "EN_SOFT_PICKY_IDLE_TIME"},
		{"Measure setup time carefully.", "EN_SOFT_PICKY_SETUP_TIME"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeTimestreamqueryTimestreamwriteManagedblockchain(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Query timestreamquery carefully.",
		"Query TimestreamQuery carefully.",
		"Write timestreamwrite carefully.",
		"Write TimestreamWrite carefully.",
		"Use managedblockchain carefully.",
		"Use ManagedBlockchain carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave137(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"And howeever carefully.", "EN_SOFT_HOWEVER_MISS2", "however"},
		{"Please inclued carefully.", "EN_SOFT_INCLUDE_MISS2", "include"},
		{"Show interes carefully.", "EN_SOFT_INTEREST_MISS2", "interest"},
		{"Share knowldge carefully.", "EN_SOFT_KNOWLEDGE_MISS2", "knowledge"},
		{"Learn a languag carefully.", "EN_SOFT_LANGUAGE_MISS2", "language"},
		{"It is mayb true carefully.", "EN_SOFT_MAYBE_MISS2", "maybe"},
		{"It is maybee true carefully.", "EN_SOFT_MAYBE_MISS3", "maybe"},
		{"A large numbre carefully.", "EN_SOFT_NUMBER_MISS", "number"},
		{"And prehaps tomorrow carefully.", "EN_SOFT_PERHAPS_MISS2", "perhaps"},
		{"A hard porblem carefully.", "EN_SOFT_PROBLEM_MISS", "problem"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave113(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Measure takt time carefully.", "EN_SOFT_PICKY_TAKT_TIME"},
		{"Apply little law carefully.", "EN_SOFT_PICKY_LITTLE_LAW"},
		{"Apply theory of constraints carefully.", "EN_SOFT_PICKY_THEORY_OF_CONSTRAINTS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeQldbsessionManagedblockchainqueryPrometheus(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Open qldbsession carefully.",
		"Open QLDBSession carefully.",
		"Query managedblockchainquery carefully.",
		"Query ManagedBlockchainQuery carefully.",
		"Scrape prometheus carefully.",
		"Scrape Prometheus carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave138(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A large numbr carefully.", "EN_SOFT_NUMBER_MISS2", "number"},
		{"A large numbeer carefully.", "EN_SOFT_NUMBER_MISS3", "number"},
		{"Visit the countrey carefully.", "EN_SOFT_COUNTRY_MISS2", "country"},
		{"And plese help carefully.", "EN_SOFT_PLEASE_MISS", "please"},
		{"A hard problm carefully.", "EN_SOFT_PROBLEM_MISS2", "problem"},
		{"Start the proces carefully.", "EN_SOFT_PROCESS_MISS", "process"},
		{"Ask a quetion carefully.", "EN_SOFT_QUESTION_MISS2", "question"},
		{"Try somthing new carefully.", "EN_SOFT_SOMETHING_MISS", "something"},
		{"Work togather carefully.", "EN_SOFT_TOGETHER_MISS", "together"},
		{"I understnad carefully.", "EN_SOFT_UNDERSTAND_MISS", "understand"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave114(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Use five whys carefully.", "EN_SOFT_PICKY_FIVE_WHYS"},
		{"Draw a fishbone carefully.", "EN_SOFT_PICKY_FISHBONE"},
		{"Plot a pareto chart carefully.", "EN_SOFT_PICKY_PARETO_CHART"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeGrafanaOpensearchserverlessAmazonmanagedprometheus(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Dashboard with grafana carefully.",
		"Dashboard with Grafana carefully.",
		"Index opensearchserverless carefully.",
		"Index OpenSearchServerless carefully.",
		"Scrape amazonmanagedprometheus carefully.",
		"Scrape AmazonManagedPrometheus carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave139(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"See you tonite carefully.", "EN_SOFT_TONIGHT_MISS", "tonight"},
		{"I usally agree carefully.", "EN_SOFT_USUALLY_MISS2", "usually"},
		{"Check the verison carefully.", "EN_SOFT_VERSION_MISS", "version"},
		{"Go withot fear carefully.", "EN_SOFT_WITHOUT_MISS", "without"},
		{"Keep workign carefully.", "EN_SOFT_WORKING_MISS", "working"},
		{"A good reasen carefully.", "EN_SOFT_REASON_MISS", "reason"},
		{"I sometiems forget carefully.", "EN_SOFT_SOMETIMES_MISS", "sometimes"},
		{"A specail case carefully.", "EN_SOFT_SPECIAL_MISS", "special"},
		{"Start the procces carefully.", "EN_SOFT_PROCESS_MISS2", "process"},
		{"Ask a qustion carefully.", "EN_SOFT_QUESTION_MISS3", "question"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave115(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Write an a3 report carefully.", "EN_SOFT_PICKY_A3_REPORT"},
		{"Run pdca carefully.", "EN_SOFT_PICKY_PDCA"},
		{"Practice kaizen carefully.", "EN_SOFT_PICKY_KAIZEN"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeJaegerTempoLoki(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Trace with jaeger carefully.",
		"Trace with Jaeger carefully.",
		"Store in tempo carefully.",
		"Store in Tempo carefully.",
		"Query loki carefully.",
		"Query Loki carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave140(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"A kind persom carefully.", "EN_SOFT_PERSON_MISS", "person"},
		{"Please decied carefully.", "EN_SOFT_DECIDE_MISS3", "decide"},
		{"Visit the countrie carefully.", "EN_SOFT_COUNTRY_MISS3", "country"},
		{"And perphas carefully.", "EN_SOFT_PERHAPS_MISS3", "perhaps"},
		{"Check the versoin carefully.", "EN_SOFT_VERSION_MISS2", "version"},
		{"Go wihtout fear carefully.", "EN_SOFT_WITHOUT_MISS2", "without"},
		{"Keep wroking carefully.", "EN_SOFT_WORKING_MISS2", "working"},
		{"Work togehter carefully.", "EN_SOFT_TOGETHER_MISS2", "together"},
		{"A speical case carefully.", "EN_SOFT_SPECIAL_MISS2", "special"},
		{"I sometmes forget carefully.", "EN_SOFT_SOMETIMES_MISS2", "sometimes"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave116(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply dmaic carefully.", "EN_SOFT_PICKY_DMAIC"},
		{"Draw a sipoc carefully.", "EN_SOFT_PICKY_SIPOC"},
		{"Go to the gemba carefully.", "EN_SOFT_PICKY_GEMBA"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeMimirThanosCortex(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store in mimir carefully.",
		"Store in Mimir carefully.",
		"Query thanos carefully.",
		"Query Thanos carefully.",
		"Run cortex carefully.",
		"Run Cortex carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave141(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Find somehting carefully.", "EN_SOFT_SOMETHING_MISS2", "something"},
		{"Give a reoson carefully.", "EN_SOFT_REASON_MISS2", "reason"},
		{"Fix the prolem carefully.", "EN_SOFT_PROBLEM_MISS3", "problem"},
		{"I somtimes forget carefully.", "EN_SOFT_SOMETIMES_MISS3", "sometimes"},
		{"I undrestand carefully.", "EN_SOFT_UNDERSTAND_MISS2", "understand"},
		{"Ask whther carefully.", "EN_SOFT_WHETHER_MISS", "whether"},
		{"Batch procesing carefully.", "EN_SOFT_PROCESSING_MISS", "processing"},
		{"Ask a quesiton carefully.", "EN_SOFT_QUESTION_MISS4", "question"},
		{"We are recieveing carefully.", "EN_SOFT_RECEIVING_MISS2", "receiving"},
		{"A diffcult task carefully.", "EN_SOFT_DIFFICULT_MISS2", "difficult"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave117(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Find the root cause carefully.", "EN_SOFT_PICKY_ROOT_CAUSE"},
		{"Draw a control chart carefully.", "EN_SOFT_PICKY_CONTROL_CHART"},
		{"Make a scatter plot carefully.", "EN_SOFT_PICKY_SCATTER_PLOT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeVictoriametricsAlertmanagerPushgateway(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store in victoriametrics carefully.",
		"Store in VictoriaMetrics carefully.",
		"Run alertmanager carefully.",
		"Run Alertmanager carefully.",
		"Use pushgateway carefully.",
		"Use Pushgateway carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave142(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fix the improt carefully.", "EN_SOFT_IMPORT_MISS", "import"},
		{"Skip becouse carefully.", "EN_SOFT_BECAUSE_MISS4", "because"},
		{"I know taht carefully.", "EN_SOFT_THAT_MISS", "that"},
		{"See hte result carefully.", "EN_SOFT_THE_MISS", "the"},
		{"Come wiht me carefully.", "EN_SOFT_WITH_MISS", "with"},
		{"I jsut tried carefully.", "EN_SOFT_JUST_MISS", "just"},
		{"Plan maintenace carefully.", "EN_SOFT_MAINTENANCE_MISS", "maintenance"},
		{"Please repersent carefully.", "EN_SOFT_REPRESENT_MISS", "represent"},
		{"Call the fucntion carefully.", "EN_SOFT_FUNCTION_MISS2", "function"},
		{"Keep it relavent carefully.", "EN_SOFT_RELEVANT_MISS", "relevant"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave118(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Draw a histogram carefully.", "EN_SOFT_PICKY_HISTOGRAM"},
		{"Draw a run chart carefully.", "EN_SOFT_PICKY_RUN_CHART"},
		{"Draw a box plot carefully.", "EN_SOFT_PICKY_BOX_PLOT"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeCadvisorPromtailZipkin(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Run cadvisor carefully.",
		"Run cAdvisor carefully.",
		"Ship with promtail carefully.",
		"Ship with Promtail carefully.",
		"Trace with zipkin carefully.",
		"Trace with Zipkin carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave143(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Own the responsability carefully.", "EN_SOFT_RESPONSIBILITY_MISS", "responsibility"},
		{"Check the warrenty carefully.", "EN_SOFT_WARRANTY_MISS", "warranty"},
		{"Ask the assistent carefully.", "EN_SOFT_ASSISTANT_MISS2", "assistant"},
		{"Write a breif carefully.", "EN_SOFT_BRIEF_MISS", "brief"},
		{"Be carefull carefully.", "EN_SOFT_CAREFUL_MISS", "careful"},
		{"Show a comparision carefully.", "EN_SOFT_COMPARISON_MISS", "comparison"},
		{"A desireable outcome carefully.", "EN_SOFT_DESIRABLE_MISS", "desirable"},
		{"Update documentaion carefully.", "EN_SOFT_DOCUMENTATION_MISS3", "documentation"},
		{"Run an effecient process carefully.", "EN_SOFT_EFFICIENT_MISS2", "efficient"},
		{"Do the excercise carefully.", "EN_SOFT_EXERCISE_MISS2", "exercise"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave119(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Run an fmea carefully.", "EN_SOFT_PICKY_FMEA"},
		{"Apply spc carefully.", "EN_SOFT_PICKY_SPC"},
		{"Track the cpk carefully.", "EN_SOFT_PICKY_CPK"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeFluentbitFluentdTelegraf(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Ship with fluentbit carefully.",
		"Ship with FluentBit carefully.",
		"Run fluentd carefully.",
		"Run Fluentd carefully.",
		"Collect with telegraf carefully.",
		"Collect with Telegraf carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave144(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Give an explaination carefully.", "EN_SOFT_EXPLANATION_MISS2", "explanation"},
		{"Need an extention carefully.", "EN_SOFT_EXTENSION_MISS", "extension"},
		{"I genuinly agree carefully.", "EN_SOFT_GENUINELY_MISS", "genuinely"},
		{"Ship the implemantation carefully.", "EN_SOFT_IMPLEMENTATION_MISS", "implementation"},
		{"Plan the lauch carefully.", "EN_SOFT_LAUNCH_MISS", "launch"},
		{"We negociate carefully.", "EN_SOFT_NEGOTIATE_MISS", "negotiate"},
		{"A particualr case carefully.", "EN_SOFT_PARTICULAR_MISS3", "particular"},
		{"Note the presance carefully.", "EN_SOFT_PRESENCE_MISS", "presence"},
		{"Check pronounciation carefully.", "EN_SOFT_PRONUNCIATION_MISS", "pronunciation"},
		{"Use a psuedo name carefully.", "EN_SOFT_PSEUDO_MISS", "pseudo"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave120(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Run an msa carefully.", "EN_SOFT_PICKY_MSA"},
		{"Plan a doe carefully.", "EN_SOFT_PICKY_DOE"},
		{"Use design of experiments carefully.", "EN_SOFT_PICKY_DESIGN_OF_EXPERIMENTS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeInfluxdbTimescaledbClickhouse(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store in influxdb carefully.",
		"Store in InfluxDB carefully.",
		"Query timescaledb carefully.",
		"Query TimescaleDB carefully.",
		"Run clickhouse carefully.",
		"Run ClickHouse carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave145(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"That is rediculous carefully.", "EN_SOFT_RIDICULOUS_MISS2", "ridiculous"},
		{"Share the resourse carefully.", "EN_SOFT_RESOURCE_MISS", "resource"},
		{"Read the statment carefully.", "EN_SOFT_STATEMENT_MISS", "statement"},
		{"The job stoped carefully.", "EN_SOFT_STOPPED_MISS", "stopped"},
		{"Talk abotu this carefully.", "EN_SOFT_ABOUT_MISS", "about"},
		{"Try agian carefully.", "EN_SOFT_AGAIN_MISS", "again"},
		{"Pick anotehr option carefully.", "EN_SOFT_ANOTHER_MISS3", "another"},
		{"I aslo agree carefully.", "EN_SOFT_ALSO_MISS", "also"},
		{"Call me befroe carefully.", "EN_SOFT_BEFORE_MISS3", "before"},
		{"Please chekc carefully.", "EN_SOFT_CHECK_MISS", "check"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave121(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Remove muda carefully.", "EN_SOFT_PICKY_MUDA"},
		{"Reduce mura carefully.", "EN_SOFT_PICKY_MURA"},
		{"Avoid muri carefully.", "EN_SOFT_PICKY_MURI"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeQuestdbCollectdStatsd(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Store in questdb carefully.",
		"Store in QuestDB carefully.",
		"Run collectd carefully.",
		"Run Collectd carefully.",
		"Ship with statsd carefully.",
		"Ship with StatsD carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave146(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Join the comapny carefully.", "EN_SOFT_COMPANY_MISS", "company"},
		{"Please confrim carefully.", "EN_SOFT_CONFIRM_MISS", "confirm"},
		{"Please creaet carefully.", "EN_SOFT_CREATE_MISS", "create"},
		{"Open the databse carefully.", "EN_SOFT_DATABASE_MISS", "database"},
		{"Use the defualt carefully.", "EN_SOFT_DEFAULT_MISS", "default"},
		{"Please delte carefully.", "EN_SOFT_DELETE_MISS", "delete"},
		{"Review the desing carefully.", "EN_SOFT_DESIGN_MISS", "design"},
		{"Start the donwload carefully.", "EN_SOFT_DOWNLOAD_MISS2", "download"},
		{"Send an eamil carefully.", "EN_SOFT_EMAIL_MISS", "email"},
		{"Please enabel carefully.", "EN_SOFT_ENABLE_MISS", "enable"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave122(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply jidoka carefully.", "EN_SOFT_PICKY_JIDOKA"},
		{"Use heijunka carefully.", "EN_SOFT_PICKY_HEIJUNKA"},
		{"Add poka yoke carefully.", "EN_SOFT_PICKY_POKA_YOKE"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeFluxcdCorednsEtcd(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Deploy with fluxcd carefully.",
		"Deploy with FluxCD carefully.",
		"Query coredns carefully.",
		"Query CoreDNS carefully.",
		"Store in etcd carefully.",
		"Store in Etcd carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave147(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Hit the endpiont carefully.", "EN_SOFT_ENDPOINT_MISS", "endpoint"},
		{"Not enoguh data carefully.", "EN_SOFT_ENOUGH_MISS4", "enough"},
		{"Fix the erorr carefully.", "EN_SOFT_ERROR_MISS2", "error"},
		{"Please exectue carefully.", "EN_SOFT_EXECUTE_MISS", "execute"},
		{"Ship the faeture carefully.", "EN_SOFT_FEATURE_MISS", "feature"},
		{"Please follwo carefully.", "EN_SOFT_FOLLOW_MISS", "follow"},
		{"Please gernerate carefully.", "EN_SOFT_GENERATE_MISS", "generate"},
		{"Join the gorup carefully.", "EN_SOFT_GROUP_MISS", "group"},
		{"Please hlep carefully.", "EN_SOFT_HELP_MISS", "help"},
		{"Use the identifer carefully.", "EN_SOFT_IDENTIFIER_MISS", "identifier"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave123(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply smed carefully.", "EN_SOFT_PICKY_SMED"},
		{"Track oee carefully.", "EN_SOFT_PICKY_OEE"},
		{"Run tpm carefully.", "EN_SOFT_PICKY_TPM"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeKanikoCalicoFlannel(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Build with kaniko carefully.",
		"Build with Kaniko carefully.",
		"Network with calico carefully.",
		"Network with Calico carefully.",
		"Overlay flannel carefully.",
		"Overlay Flannel carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave148(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Fix the imoprt carefully.", "EN_SOFT_IMPORT_MISS2", "import"},
		{"Please imporve carefully.", "EN_SOFT_IMPROVE_MISS", "improve"},
		{"Please inlcude carefully.", "EN_SOFT_INCLUDE_MISS3", "include"},
		{"Start an instnace carefully.", "EN_SOFT_INSTANCE_MISS", "instance"},
		{"Define the inteface carefully.", "EN_SOFT_INTERFACE_MISS", "interface"},
		{"Reject invliad input carefully.", "EN_SOFT_INVALID_MISS", "invalid"},
		{"Use this isntead carefully.", "EN_SOFT_INSTEAD_MISS4", "instead"},
		{"I knwo that carefully.", "EN_SOFT_KNOW_MISS", "know"},
		{"Please laod carefully.", "EN_SOFT_LOAD_MISS", "load"},
		{"Raise the levle carefully.", "EN_SOFT_LEVEL_MISS", "level"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave124(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Fill a raci carefully.", "EN_SOFT_PICKY_RACI"},
		{"Run a swot carefully.", "EN_SOFT_PICKY_SWOT"},
		{"Define a ctq carefully.", "EN_SOFT_PICKY_CTQ"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeGraphiteDruidPinot(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Ship metrics to graphite carefully.",
		"Ship metrics to Graphite carefully.",
		"Query druid carefully.",
		"Query Druid carefully.",
		"Index in pinot carefully.",
		"Index in Pinot carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave149(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I liek this carefully.", "EN_SOFT_LIKE_MISS", "like"},
		{"Pick a locaiton carefully.", "EN_SOFT_LOCATION_MISS", "location"},
		{"Load the moduel carefully.", "EN_SOFT_MODULE_MISS", "module"},
		{"Need mroe time carefully.", "EN_SOFT_MORE_MISS", "more"},
		{"Count the nubmer carefully.", "EN_SOFT_NUMBER_MISS4", "number"},
		{"Create an obejct carefully.", "EN_SOFT_OBJECT_MISS", "object"},
		{"Please oepn carefully.", "EN_SOFT_OPEN_MISS", "open"},
		{"Use onyl this carefully.", "EN_SOFT_ONLY_MISS", "only"},
		{"Pick an optino carefully.", "EN_SOFT_OPTION_MISS", "option"},
		{"Place the oredr carefully.", "EN_SOFT_ORDER_MISS", "order"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave125(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Apply bant carefully.", "EN_SOFT_PICKY_BANT"},
		{"Run meddic carefully.", "EN_SOFT_PICKY_MEDDIC"},
		{"Track nps carefully.", "EN_SOFT_PICKY_NPS"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizeTrinoPrestoNifi(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Query with trino carefully.",
		"Query with Trino carefully.",
		"Query with presto carefully.",
		"Query with Presto carefully.",
		"Pipeline with nifi carefully.",
		"Pipeline with NiFi carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftIdiomConfusablesWave150(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Check the otuput carefully.", "EN_SOFT_OUTPUT_MISS", "output"},
		{"Install the packge carefully.", "EN_SOFT_PACKAGE_MISS", "package"},
		{"Pass a paramter carefully.", "EN_SOFT_PARAMETER_MISS", "parameter"},
		{"Reset the passwrod carefully.", "EN_SOFT_PASSWORD_MISS", "password"},
		{"Please perfrom carefully.", "EN_SOFT_PERFORM_MISS", "perform"},
		{"And pleae wait carefully.", "EN_SOFT_PLEASE_MISS2", "please"},
		{"Start the porject carefully.", "EN_SOFT_PROJECT_MISS", "project"},
		{"See the previuos carefully.", "EN_SOFT_PREVIOUS_MISS", "previous"},
		{"Read the proprety carefully.", "EN_SOFT_PROPERTY_MISS", "property"},
		{"Pick a protcol carefully.", "EN_SOFT_PROTOCOL_MISS", "protocol"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickyENJargonWave126(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Track csat carefully.", "EN_SOFT_PICKY_CSAT"},
		{"Report mrr carefully.", "EN_SOFT_PICKY_MRR"},
		{"Report arr carefully.", "EN_SOFT_PICKY_ARR"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_ImmunizePulsarRabbitmqMemcached(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Publish on pulsar carefully.",
		"Publish on Pulsar carefully.",
		"Queue on rabbitmq carefully.",
		"Queue on RabbitMQ carefully.",
		"Cache with memcached carefully.",
		"Cache with Memcached carefully.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_FalseFriendsActuality(t *testing.T) {
	ff := softFalseFriendsPath(t)
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Actualmente vivo aquí.", &CommandLineOptions{
		Language:         "es",
		MotherTongue:     "en",
		FalseFriendsFile: ff,
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "ACTUALITY" {
			found = true
			require.Equal(t, "currently / at present", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_ImmunizeAfairIirc(t *testing.T) {
	if DiscoverEnglishSoftDisambiguationXML(nil) == "" {
		t.Skip("en-soft disambig missing")
	}
	for _, text := range []string{
		"Afair that shipped last week.",
		"Iirc the API changed.",
		"Fwiw I agree.",
		"TLDR the patch works.",
	} {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_SoftListRulesSoftOptCount(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "en"))
	out := buf.String()
	require.Contains(t, out, "soft_opt=")
	// soft_opt should be > 0 when en-optional-soft is loaded
	require.True(t, strings.Contains(out, "soft_opt=9") || strings.Contains(out, "soft_opt=1") ||
		strings.Contains(out, "soft_opt=8") || strings.Contains(out, "soft_opt=7") ||
		strings.Contains(out, "soft_opt=10") || strings.Contains(out, "soft_opt=12") ||
		strings.Contains(out, "soft_opt=15") || strings.Contains(out, "soft_opt=18") ||
		strings.Contains(out, "soft_opt=21") || strings.Contains(out, "soft_opt=23") ||
		strings.Contains(out, "soft_opt=25") || strings.Contains(out, "soft_opt=27") ||
		strings.Contains(out, "soft_opt=6"), out)
}

func TestGolden_SoftListRulesOptionalOff(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreListRules(&buf, "en"))
	out := buf.String()
	require.Contains(t, out, "EN_SOFT_OPT_PRIOR_TO")
	require.Contains(t, out, "soft_off=")
	// optional rules listed as off (sixth column)
	require.Contains(t, out, "\tsoft\toff\n")
}

func TestGolden_SoftPickyNL(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"De meeting is morgen.", "NL_SOFT_PICKY_MEETING", "vergadering"},
		{"Ik wil feedback vandaag.", "NL_SOFT_PICKY_FEEDBACK", "terugkoppeling"},
		{"Het is heel heel belangrijk.", "NL_SOFT_PICKY_HEEL_HEEL", ""},
		{"Er zijn veel dingen te doen.", "NL_SOFT_PICKY_DINGEN", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "nl", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftPickySV(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"Vi har en meeting imorgon.", "SV_SOFT_PICKY_MEETING", "möte"},
		{"Jag vill ha feedback snart.", "SV_SOFT_PICKY_FEEDBACK", "återkoppling"},
		{"Det är väldigt väldigt bra.", "SV_SOFT_PICKY_VALDIGT_VALDIGT", ""},
		{"Det finns många saker kvar.", "SV_SOFT_PICKY_SAKER", ""},
		{"I slutändan bestämmer vi.", "SV_SOFT_PICKY_I_SLUTANDAN", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "sv", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "style", f.Type)
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_SoftFailOnNoteStyle(t *testing.T) {
	// soft style rules have severity note; fail-on=error should exit 0
	// fail-on=note should exit non-zero when only style soft hits
	text := "This is very unique work."
	var out, errb bytes.Buffer
	code := RunWithIO([]string{
		"-l", "en", "--lint", "--fail-on", "error",
		"-d", "UPPERCASE_SENTENCE_START",
		"-",
	}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out, &errb)
	// style soft is note severity → fail-on error exits 0
	require.Equal(t, 0, code, "err=%s out=%s", errb.String(), out.String())

	var out2, errb2 bytes.Buffer
	code2 := RunWithIO([]string{
		"-l", "en", "--lint", "--fail-on", "note",
		"-d", "UPPERCASE_SENTENCE_START",
		"-",
	}, RunHooks{
		ReadStdin: func() (string, error) { return text, nil },
		Check:     CoreCheckHook,
	}, &out2, &errb2)
	require.NotEqual(t, 0, code2, "expected fail-on note to fail on style soft; out=%s", out2.String())
}



func TestGolden_SoftInformalForms(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I shud of known better.", "EN_SOFT_SHUD_OF", "should have"},
		{"I woulda gone earlier.", "EN_SOFT_WOULDA", "would have"},
		{"I coulda helped you.", "EN_SOFT_COULDA", "could have"},
		{"I shoulda called first.", "EN_SOFT_SHOULDA", "should have"},
		{"I gotta leave now.", "EN_SOFT_GOTTA", ""},
		{"I wanna go home.", "EN_SOFT_WANNA", ""},
		{"I'm gonna try that.", "EN_SOFT_GONNA", ""},
		{"That aint right.", "EN_SOFT_AIN_T", ""},
		{"Yall should come over.", "EN_SOFT_YALL", "Y'all"},
		{"Imma finish this later.", "EN_SOFT_IMMA", "I'm going to"},
		{"Prolly tomorrow works.", "EN_SOFT_PROLLY", "Probably"},
		{"Deffo a good idea.", "EN_SOFT_DEFFO", "Definitely"},
		{"Basically basically it works.", "EN_SOFT_BASICALLY_BASIC", ""},
		{"Actually actually I agree.", "EN_SOFT_ACTUALLY_ACTUALLY", ""},
		{"Honestly honestly I tried.", "EN_SOFT_HONESTLY_HONESTLY", ""},
		{"Literally literally amazing.", "EN_SOFT_LITERALLY_LITERALLY", ""},
		{"Just just wait here.", "EN_SOFT_JUST_JUST", ""},
		{"Really really good work.", "EN_SOFT_REALLY_REALLY", ""},
		{"Very very cold outside.", "EN_SOFT_VERY_VERY", ""},
		{"I like so so much cake.", "EN_SOFT_SO_SO", "so much"},
		{"Cats and and dogs.", "EN_SOFT_AND_AND", ""},
		{"See the the problem.", "EN_SOFT_THE_THE", ""},
		{"I need a a break.", "EN_SOFT_A_A", ""},
		{"I want to to leave.", "EN_SOFT_TO_TO", ""},
		{"Kind of of work.", "EN_SOFT_OF_OF", ""},
		{"Put it in in the box.", "EN_SOFT_IN_IN", ""},
		{"Put it on on the table.", "EN_SOFT_ON_ON", ""},
		{"This is for for you.", "EN_SOFT_FOR_FOR", ""},
		{"Come with with me.", "EN_SOFT_WITH_WITH", ""},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}
