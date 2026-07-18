package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_UpstreamEN23Misses(t *testing.T) {
	cases := []struct{ rule, text string }{
		{"WRB_THERE_THEY_RE", "Wherever there going, I will follow them."},
		{"APPLY_FOR", "You try and apply for another University!"},
		{"TAKING_CASE_OF_IT", "We need to take case of it."},
		{"WORSE_WORST", "Worse came to worse."},
		{"COME_THROUGH", "They came throw the door."},
		{"GOING_TO_VACATION", "She is going to vacation."},
		{"TO_ON_A_TRIP", "Yes, I went to a trip."},
		{"COME_TO_PLANE", "I came to plane."},
		{"PRP_HAFT", "They will haft to go on."},
		{"OBJECTIVE_CASE", "Come with I."},
		{"NIT_NOT", "I could nit do it."},
		{"WITHE_WITH", "I backed the project withe my personal money."},
		{"HER_HEAR", "I'd like to her the truth."},
		{"MAKE_AN_ATTEMPT", "We should make an effort to win."},
		{"EXCEPTION_PREPOSITION_THE_RULE", "Graphite is an exception of the rule."},
		{"MUCH_NEEDED_HYPHEN", "The film gave a much needed boost to the country's tourist industry."},
		{"TAKE_THE_REIGNS", "It's time to take the reigns of leadership."},
		{"PARTICIPATE_TO_IN", "They participate to many activities."},
		{"BE_WILL", "How is would this approach be useful?"},
		{"BIS_BUS", "I will take the bis to the central station."},
		{"EAT_ANTIBIOTICS", "I ate medicine for 2 weeks after my operation."},
		{"LUNCH_TO_FOR", "The lunch to the guests is ready."},
		{"CRAZY_ON_WITH", "What makes me crazy on this dwelling is that some time we are higher than the clouds."},
	}
	var miss []string
	for _, tc := range cases {
		var buf bytes.Buffer
		_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
		require.NoError(t, err)
		var findings []Finding
		require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
		found := false
		for _, f := range findings {
			if f.Rule == tc.rule {
				found = true
				break
			}
		}
		if !found {
			miss = append(miss, tc.rule)
			t.Logf("MISS %s", tc.rule)
		} else {
			t.Logf("OK   %s", tc.rule)
		}
	}
	require.Empty(t, miss, "still missing: %v", miss)
}
