package server

// SynthesizeFn is a pluggable synthesizer backend (languageCode, lemma, postag, regexp).
type SynthesizeFn func(languageCode, lemma, postag string, postagRegexp bool) ([]string, error)

// RemoteSynthesizer ports org.languagetool.server.RemoteSynthesizer core (without gRPC).
type RemoteSynthesizer struct {
	Synthesize SynthesizeFn
}

func NewRemoteSynthesizer(fn SynthesizeFn) *RemoteSynthesizer {
	return &RemoteSynthesizer{Synthesize: fn}
}

// SynthesizeForms synthesizes forms and removes duplicates.
func (r *RemoteSynthesizer) SynthesizeForms(languageCode, lemma, postag string, postagRegexp bool) ([]string, error) {
	if r == nil || r.Synthesize == nil {
		return nil, nil
	}
	forms, err := r.Synthesize(languageCode, lemma, postag, postagRegexp)
	if err != nil {
		return nil, err
	}
	return removeDuplicateStrings(forms), nil
}

func removeDuplicateStrings(forms []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range forms {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
