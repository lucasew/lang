package implementation

// SuggestionStage ports org.languagetool.rules.spelling.symspell.implementation.SuggestionStage.
// An intentionally opaque class used to temporarily stage dictionary data during the
// adding of many words (significant time savings and reduced final memory usage).
type SuggestionStage struct {
	// Deletes ports public Map<Integer, Entry> deletes.
	Deletes map[int]*StageEntry
	// Nodes ports public ChunkArray<Node> nodes.
	Nodes *ChunkArray[StageNode]
}

// StageNode ports SuggestionStage.Node (nested static class).
type StageNode struct {
	// Suggestion ports public String suggestion.
	Suggestion string
	// Next ports public int next.
	Next int
}

// NewStageNode ports Node(String suggestion, int next).
func NewStageNode(suggestion string, next int) StageNode {
	return StageNode{Suggestion: suggestion, Next: next}
}

// StageEntry ports SuggestionStage.Entry (nested static class).
type StageEntry struct {
	// Count ports public int count.
	Count int
	// First ports public int first.
	First int
}

// NewStageEntry ports Entry(int count, int first) (package-private in Java).
func NewStageEntry(count, first int) *StageEntry {
	return &StageEntry{Count: count, First: first}
}

// NewSuggestionStage ports SuggestionStage(int initialCapacity).
// Java: deletes = new HashMap<>(initialCapacity); nodes = new ChunkArray<>(initialCapacity * 2).
// Capacity is passed through (HashMap(0) is legal); no minimum invent.
func NewSuggestionStage(initialCapacity int) *SuggestionStage {
	return &SuggestionStage{
		Deletes: make(map[int]*StageEntry, initialCapacity),
		Nodes:   NewChunkArray[StageNode](initialCapacity * 2),
	}
}

// DeleteCount ports deleteCount(): deletes.size().
func (s *SuggestionStage) DeleteCount() int {
	return len(s.Deletes)
}

// NodeCount ports nodeCount(): nodes.count.
func (s *SuggestionStage) NodeCount() int {
	return s.Nodes.Count
}

// Clear ports clear(): deletes.clear(); nodes.clear().
func (s *SuggestionStage) Clear() {
	clear(s.Deletes)
	s.Nodes.Clear()
}

// Add ports void add(int deleteHash, String suggestion).
// Linked list is LIFO: walk from first yields most recently added first.
func (s *SuggestionStage) Add(deleteHash int, suggestion string) {
	// Entry entry = deletes.getOrDefault(deleteHash, new Entry(0, -1));
	entry, ok := s.Deletes[deleteHash]
	if !ok {
		entry = NewStageEntry(0, -1)
	}
	next := entry.First
	entry.Count++
	entry.First = s.Nodes.Count
	s.Deletes[deleteHash] = entry
	s.Nodes.Add(NewStageNode(suggestion, next))
}

// CommitTo ports void commitTo(Map<Integer, String[]> permanentDeletes).
// Existing permanent suggestions stay as prefix; staged append in LIFO walk order.
func (s *SuggestionStage) CommitTo(permanentDeletes map[int][]string) {
	for key, value := range s.Deletes {
		var i int
		var suggestions []string
		if _, ok := permanentDeletes[key]; ok {
			suggestions = permanentDeletes[key]
			i = len(suggestions)
			// Arrays.copyOf(suggestions, i + value.count)
			newSuggestion := make([]string, i+value.Count)
			copy(newSuggestion, suggestions)
			permanentDeletes[key] = newSuggestion
			suggestions = newSuggestion
		} else {
			i = 0
			suggestions = make([]string, value.Count)
			permanentDeletes[key] = suggestions
		}
		next := value.First
		for next >= 0 {
			node := s.Nodes.Get(next)
			suggestions[i] = node.Suggestion
			next = node.Next
			i++
		}
	}
}
