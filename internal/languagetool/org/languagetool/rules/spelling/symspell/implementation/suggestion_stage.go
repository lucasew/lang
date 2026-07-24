package implementation

// SuggestionStage ports SymSpell SuggestionStage for staged dictionary building.
type SuggestionStage struct {
	Deletes map[int]*StageEntry
	Nodes   *ChunkArray[StageNode]
}

// StageNode ports SuggestionStage.Node.
type StageNode struct {
	Suggestion string
	Next       int
}

// StageEntry ports SuggestionStage.Entry.
type StageEntry struct {
	Count int
	First int
}

func NewSuggestionStage(initialCapacity int) *SuggestionStage {
	if initialCapacity <= 0 {
		initialCapacity = 16
	}
	return &SuggestionStage{
		Deletes: map[int]*StageEntry{},
		Nodes:   NewChunkArray[StageNode](initialCapacity * 2),
	}
}

func (s *SuggestionStage) DeleteCount() int {
	if s == nil {
		return 0
	}
	return len(s.Deletes)
}

func (s *SuggestionStage) NodeCount() int {
	if s == nil || s.Nodes == nil {
		return 0
	}
	return s.Nodes.Count
}

func (s *SuggestionStage) Clear() {
	if s == nil {
		return
	}
	s.Deletes = map[int]*StageEntry{}
	if s.Nodes != nil {
		s.Nodes.Clear()
	}
}

// Add stages a suggestion under a delete-hash key.
func (s *SuggestionStage) Add(deleteHash int, suggestion string) {
	entry, ok := s.Deletes[deleteHash]
	if !ok {
		entry = &StageEntry{Count: 0, First: -1}
		s.Deletes[deleteHash] = entry
	}
	next := entry.First
	entry.Count++
	entry.First = s.Nodes.Count
	s.Nodes.Add(StageNode{Suggestion: suggestion, Next: next})
}

// CommitTo flushes staged data into permanentDeletes (deleteHash → suggestions).
func (s *SuggestionStage) CommitTo(permanentDeletes map[int][]string) {
	if s == nil || permanentDeletes == nil {
		return
	}
	for key, value := range s.Deletes {
		suggestions := permanentDeletes[key]
		i := len(suggestions)
		// grow
		newSug := make([]string, i+value.Count)
		copy(newSug, suggestions)
		permanentDeletes[key] = newSug
		suggestions = newSug
		next := value.First
		for next >= 0 {
			node := s.Nodes.Get(next)
			suggestions[i] = node.Suggestion
			next = node.Next
			i++
		}
	}
}
