package implementation

// ChunkArray ports org.languagetool.rules.spelling.symspell.implementation.ChunkArray.
//
// A growable list optimized for adds (not deletes), storing data in GC-friendly
// chunks rather than one monolithic array. Generic over T for Java ChunkArray<T>
// (used with SuggestionStage.Node).
//
// chunkSize must be a power of 2 so row/col can use shifts and masks.
const (
	chunkArrayChunkSize = 4096 // power of 2
	chunkArrayDivShift  = 12   // bit position of chunkSize (index >> 12 == index / 4096)
)

// ChunkArray is the Go twin of Java ChunkArray<T>.
type ChunkArray[T any] struct {
	// Values is public in Java (values). Outer length is number of chunks;
	// each inner slice has length chunkSize.
	Values [][]T
	// Count is the number of elements added (public count in Java).
	Count int
}

// NewChunkArray ports ChunkArray(int initialCapacity).
// chunks = (initialCapacity + chunkSize - 1) / chunkSize — may be 0 when initialCapacity is 0.
func NewChunkArray[T any](initialCapacity int) *ChunkArray[T] {
	chunks := (initialCapacity + chunkArrayChunkSize - 1) / chunkArrayChunkSize
	values := make([][]T, chunks)
	for i := range values {
		values[i] = make([]T, chunkArrayChunkSize)
	}
	return &ChunkArray[T]{Values: values}
}

// capacity ports private capacity(): values.length * chunkSize.
func (a *ChunkArray[T]) capacity() int {
	return len(a.Values) * chunkArrayChunkSize
}

// row ports private row(index): index >> divShift.
func (a *ChunkArray[T]) row(index int) int {
	return index >> chunkArrayDivShift
}

// col ports private col(index): index & (chunkSize - 1).
func (a *ChunkArray[T]) col(index int) int {
	return index & (chunkArrayChunkSize - 1)
}

// Add ports add(Node value): grow by one chunk if full, store at count, return index.
func (a *ChunkArray[T]) Add(value T) int {
	if a.Count == a.capacity() {
		// Java: Arrays.copyOf(values, values.length + 1); newValues[values.length] = new Node[chunkSize]
		a.Values = append(a.Values, make([]T, chunkArrayChunkSize))
	}
	a.Values[a.row(a.Count)][a.col(a.Count)] = value
	a.Count++
	return a.Count - 1
}

// Clear ports clear(): only resets count; does not shrink or zero residual storage.
func (a *ChunkArray[T]) Clear() {
	a.Count = 0
}

// Get ports getValues(int index).
func (a *ChunkArray[T]) Get(index int) T {
	return a.Values[a.row(index)][a.col(index)]
}

// Set ports setValues(int index, Node value).
func (a *ChunkArray[T]) Set(index int, value T) {
	a.Values[a.row(index)][a.col(index)] = value
}

// SetOn ports setValues(int index, Node value, Node[][] list) writing into an external list.
// No Java call sites in-tree; kept for line-by-line twin of the public overload.
func (a *ChunkArray[T]) SetOn(list [][]T, index int, value T) {
	list[a.row(index)][a.col(index)] = value
}
