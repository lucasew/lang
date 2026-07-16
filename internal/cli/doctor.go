package cli

import (
	"fmt"

	"github.com/lucasew/lang/internal/engine"
	"github.com/lucasew/lang/internal/pipeline"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Show data path, languages, and pipeline stage status (dev helper)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := engine.New(dataDirFlag(cmd))
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "data_root\t%s\n", c.DataRoot())
			fmt.Fprintf(out, "languages\t%d\n", len(c.Languages()))
			fmt.Fprintf(out, "english_tagger\t%v\n", c.HasEnglishTagger())
			fmt.Fprintf(out, "english_speller\t%v\n", c.HasEnglishSpeller())
			fmt.Fprintln(out, "pipeline_stages:")
			implemented := map[string]string{
				pipeline.StageSentenceSplit: "srx (segment.srx)",
				pipeline.StageTokenize:      "WordTokenizer",
				pipeline.StageTag:           "morfologik english.dict (en)",
				pipeline.StageDisambiguate:  "xml rules subset (en)",
				pipeline.StageRules:         "pattern XML + whitespace/word-repeat + speller",
				pipeline.StageFilters:       "default=off, antipattern, chunk skipped",
				pipeline.StageSuggestions:   "static suggestions",
			}
			for _, s := range pipeline.AllStages {
				fmt.Fprintf(out, "  %s\t%s\n", s, implemented[s])
			}
			return nil
		},
	}
}
