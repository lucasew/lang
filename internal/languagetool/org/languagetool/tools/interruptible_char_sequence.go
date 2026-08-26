package tools

import "context"

// InterruptibleCharSequence ports org.languagetool.tools.InterruptibleCharSequence.
// Char access checks ctx; if cancelled/done, CharAt panics with ctx.Err().
// Use with regexp matching of untrusted input to allow cooperative cancel.
type InterruptibleCharSequence struct {
	inner string
	ctx   context.Context
}

// NewInterruptibleCharSequence wraps s; ctx may be nil (no interrupt checks).
func NewInterruptibleCharSequence(s string, ctx context.Context) InterruptibleCharSequence {
	return InterruptibleCharSequence{inner: s, ctx: ctx}
}

func (i InterruptibleCharSequence) Len() int { return len(i.inner) }

// CharAt returns the byte at index (Java charAt is UTF-16; we use byte index for Go strings
// used as regexp input via String()). Prefer String() for full matching.
func (i InterruptibleCharSequence) CharAt(index int) byte {
	i.check()
	return i.inner[index]
}

func (i InterruptibleCharSequence) SubSequence(start, end int) InterruptibleCharSequence {
	i.check()
	return InterruptibleCharSequence{inner: i.inner[start:end], ctx: i.ctx}
}

func (i InterruptibleCharSequence) String() string {
	i.check()
	return i.inner
}

func (i InterruptibleCharSequence) check() {
	if i.ctx == nil {
		return
	}
	select {
	case <-i.ctx.Done():
		panic(i.ctx.Err())
	default:
	}
}
