package languagetool

import (
	"sort"
	"strings"
)

// LevelToneTagCacheKey ports org.languagetool.LevelToneTagCacheKey.
type LevelToneTagCacheKey struct {
	Level    Level
	ToneTags []ToneTag // sorted unique
}

func NewLevelToneTagCacheKey(level Level, toneTags []ToneTag) LevelToneTagCacheKey {
	seen := map[ToneTag]struct{}{}
	var tags []ToneTag
	for _, t := range toneTags {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		tags = append(tags, t)
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })
	return LevelToneTagCacheKey{Level: level, ToneTags: tags}
}

func (k LevelToneTagCacheKey) Equal(o LevelToneTagCacheKey) bool {
	if k.Level != o.Level || len(k.ToneTags) != len(o.ToneTags) {
		return false
	}
	for i := range k.ToneTags {
		if k.ToneTags[i] != o.ToneTags[i] {
			return false
		}
	}
	return true
}

func (k LevelToneTagCacheKey) String() string {
	parts := make([]string, len(k.ToneTags))
	for i, t := range k.ToneTags {
		parts[i] = string(t)
	}
	return string(k.Level) + "|" + strings.Join(parts, ",")
}
