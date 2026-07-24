package implementation

// ChunkArray is a growable list optimized for adds (ports ChunkArray<Node>).
// Generic over any pointer/value type for reuse beyond SuggestionStage.Node.
type ChunkArray[T any] struct {
	chunkSize int
	divShift  int
	values    [][]T
	Count     int
}

func NewChunkArray[T any](initialCapacity int) *ChunkArray[T] {
	const chunkSize = 4096
	const divShift = 12
	chunks := (initialCapacity + chunkSize - 1) / chunkSize
	if chunks < 1 {
		chunks = 1
	}
	values := make([][]T, chunks)
	for i := range values {
		values[i] = make([]T, chunkSize)
	}
	return &ChunkArray[T]{chunkSize: chunkSize, divShift: divShift, values: values}
}

func (a *ChunkArray[T]) capacity() int { return len(a.values) * a.chunkSize }

func (a *ChunkArray[T]) row(index int) int { return index >> a.divShift }
func (a *ChunkArray[T]) col(index int) int { return index & (a.chunkSize - 1) }

// Add appends value and returns its index.
func (a *ChunkArray[T]) Add(value T) int {
	if a.Count == a.capacity() {
		a.values = append(a.values, make([]T, a.chunkSize))
	}
	a.values[a.row(a.Count)][a.col(a.Count)] = value
	a.Count++
	return a.Count - 1
}

func (a *ChunkArray[T]) Clear() { a.Count = 0 }

func (a *ChunkArray[T]) Get(index int) T {
	return a.values[a.row(index)][a.col(index)]
}

func (a *ChunkArray[T]) Set(index int, value T) {
	a.values[a.row(index)][a.col(index)] = value
}
