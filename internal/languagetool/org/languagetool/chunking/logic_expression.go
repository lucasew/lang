package chunking

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// LogicExpression ports edu.washington.cs.knowitall.logic.LogicExpression for
// TokenExpressionFactory (AND &, OR |, NOT !, parentheses).
// Dependency of org.languagetool.chunking.TokenExpressionFactory / GermanChunker.

type LogicExpression[T any] struct {
	apply func(T) bool
	empty bool
}

// CompileLogicExpression compiles an infix logic string; atomFactory builds
// predicates for argument tokens (e.g. TokenPredicate descriptions).
func CompileLogicExpression[T any](input string, atomFactory func(string) func(T) bool) *LogicExpression[T] {
	input = tools.JavaStringTrim(input)
	if input == "" {
		return &LogicExpression[T]{empty: true, apply: func(T) bool { return true }}
	}
	tokens := tokenizeLogic(input)
	rpn := logicToRPN(tokens)
	fn := buildLogicAST(rpn, atomFactory)
	return &LogicExpression[T]{apply: fn}
}

func (e *LogicExpression[T]) Apply(v T) bool {
	if e == nil || e.empty {
		return true
	}
	return e.apply(v)
}

type logicTokKind int

const (
	ltAtom logicTokKind = iota
	ltAnd
	ltOr
	ltNot
	ltLParen
	ltRParen
)

type logicTok struct {
	kind logicTokKind
	atom string // for ltAtom
}

func tokenizeLogic(input string) []logicTok {
	var tokens []logicTok
	i := 0
	for i < len(input) {
		c := input[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		switch c {
		case '(':
			tokens = append(tokens, logicTok{kind: ltLParen})
			i++
		case ')':
			tokens = append(tokens, logicTok{kind: ltRParen})
			i++
		case '!':
			tokens = append(tokens, logicTok{kind: ltNot})
			i++
		case '&':
			tokens = append(tokens, logicTok{kind: ltAnd})
			i++
		case '|':
			tokens = append(tokens, logicTok{kind: ltOr})
			i++
		default:
			tok := readLogicToken(input[i:])
			tokens = append(tokens, logicTok{kind: ltAtom, atom: tok})
			i += len(tok)
			// trim was applied inside read; account for leading spaces already skipped
			// readLogicToken returns trimmed content; advance by raw length including
			// any trailing content until operator — Java uses token.length() on trimmed
			// but advances by token.length() only; spaces after token are skipped next loop.
			// If remainder had trailing spaces inside token trim, length shrinks — match
			// Java: token = remainder.substring(0, nextExpression).trim(); i += token.length()
			// which can desync if leading spaces in remainder (we skip spaces first).
		}
	}
	return tokens
}

// readLogicToken ports LogicExpressionParser.readToken.
func readLogicToken(remainder string) string {
	depth := 0
	next := 0
	for next < len(remainder) {
		c := remainder[next]
		// single-quoted literal
		if c == '\'' {
			end := next + 1
			for end < len(remainder) && remainder[end] != '\'' {
				end++
			}
			if end < len(remainder) {
				next = end + 1
				continue
			}
		}
		// double-quoted
		if c == '"' {
			end := next + 1
			for end < len(remainder) && remainder[end] != '"' {
				if remainder[end] == '\\' && end+1 < len(remainder) {
					end += 2
					continue
				}
				end++
			}
			if end < len(remainder) {
				next = end + 1
				continue
			}
		}
		if c == '(' {
			depth++
		} else if c == ')' {
			if depth == 0 {
				break
			}
			depth--
		} else if c == '&' || c == '|' {
			break
		}
		next++
	}
	token := tools.JavaStringTrim(remainder[:next])
	if token == "" {
		panic(fmt.Sprintf("logic: zero-length token in %q", remainder))
	}
	return token
}

// shunting-yard: NOT right-assoc high prec, AND mid, OR low.
func logicToRPN(tokens []logicTok) []logicTok {
	var out []logicTok
	var opStack []logicTok
	prec := func(k logicTokKind) int {
		switch k {
		case ltNot:
			return 3
		case ltAnd:
			return 2
		case ltOr:
			return 1
		default:
			return 0
		}
	}
	isRight := func(k logicTokKind) bool { return k == ltNot }

	for _, t := range tokens {
		switch t.kind {
		case ltAtom:
			out = append(out, t)
		case ltLParen:
			opStack = append(opStack, t)
		case ltRParen:
			for len(opStack) > 0 && opStack[len(opStack)-1].kind != ltLParen {
				out = append(out, opStack[len(opStack)-1])
				opStack = opStack[:len(opStack)-1]
			}
			if len(opStack) == 0 {
				panic("logic: mismatched parentheses")
			}
			opStack = opStack[:len(opStack)-1] // pop (
		case ltNot, ltAnd, ltOr:
			for len(opStack) > 0 {
				top := opStack[len(opStack)-1]
				if top.kind == ltLParen {
					break
				}
				if isRight(t.kind) {
					if prec(top.kind) > prec(t.kind) {
						out = append(out, top)
						opStack = opStack[:len(opStack)-1]
						continue
					}
				} else {
					if prec(top.kind) >= prec(t.kind) {
						out = append(out, top)
						opStack = opStack[:len(opStack)-1]
						continue
					}
				}
				break
			}
			opStack = append(opStack, t)
		}
	}
	for len(opStack) > 0 {
		out = append(out, opStack[len(opStack)-1])
		opStack = opStack[:len(opStack)-1]
	}
	return out
}

func buildLogicAST[T any](rpn []logicTok, atomFactory func(string) func(T) bool) func(T) bool {
	if len(rpn) == 0 {
		return func(T) bool { return true }
	}
	var stack []func(T) bool
	for _, t := range rpn {
		switch t.kind {
		case ltAtom:
			stack = append(stack, atomFactory(t.atom))
		case ltNot:
			if len(stack) < 1 {
				panic("logic: NOT without operand")
			}
			sub := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			stack = append(stack, func(v T) bool { return !sub(v) })
		case ltAnd:
			if len(stack) < 2 {
				panic("logic: AND needs 2 operands")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, func(v T) bool { return a(v) && b(v) })
		case ltOr:
			if len(stack) < 2 {
				panic("logic: OR needs 2 operands")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, func(v T) bool { return a(v) || b(v) })
		default:
			panic(fmt.Sprintf("logic: unexpected rpn token %v", t.kind))
		}
	}
	if len(stack) != 1 {
		panic(fmt.Sprintf("logic: stack size %d after apply", len(stack)))
	}
	return stack[0]
}

// ensure unicode imported for possible future whitespace — use strings only
