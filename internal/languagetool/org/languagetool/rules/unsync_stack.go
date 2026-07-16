package rules

import "errors"

// ErrEmptyStack mirrors java.util.EmptyStackException for UnsyncStack.
var ErrEmptyStack = errors.New("empty stack")

// UnsyncStack ports org.languagetool.rules.UnsyncStack.
type UnsyncStack[E any] struct {
	data []E
}

func NewUnsyncStack[E any]() *UnsyncStack[E] {
	return &UnsyncStack[E]{}
}

func (s *UnsyncStack[E]) Push(item E) E {
	s.data = append(s.data, item)
	return item
}

func (s *UnsyncStack[E]) Empty() bool {
	return len(s.data) == 0
}

func (s *UnsyncStack[E]) Peek() E {
	if len(s.data) == 0 {
		panic(ErrEmptyStack)
	}
	return s.data[len(s.data)-1]
}

func (s *UnsyncStack[E]) Pop() E {
	obj := s.Peek()
	s.data = s.data[:len(s.data)-1]
	return obj
}

func (s *UnsyncStack[E]) Search(o E, eq func(a, b E) bool) int {
	for i := len(s.data) - 1; i >= 0; i-- {
		if eq(s.data[i], o) {
			return len(s.data) - i
		}
	}
	return -1
}

func (s *UnsyncStack[E]) Data() []E { return s.data }
func (s *UnsyncStack[E]) Len() int { return len(s.data) }
func (s *UnsyncStack[E]) At(i int) E { return s.data[i] }
