package scanner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
	"strconv"
	"strings"
	"unicode"
)

type scanner struct {
	r         *bufio.Reader
	line, col int
	c         rune
	err       error

	buf strings.Builder
	tok Token
}

func Scan(r io.Reader) iter.Seq2[Token, error] {
	s := scanner{
		r:    bufio.NewReader(r),
		line: 1, col: 1,
	}
	state := s.start

	return func(yield func(Token, error) bool) {
		for s.err != nil {
			if s.tok.Val != nil {
				if !yield(s.tok, nil) {
					return
				}
				s.tok = Token{}
			}

			state = state()
		}

		if !errors.Is(s.err, io.EOF) {
			yield(Token{}, s.err)
		}
	}
}

func (s *scanner) raiseToken(err error) {
	s.err = &TokenError{
		Line: s.tok.Line,
		Col:  s.tok.Col,
		Err:  err,
	}
}

func (s *scanner) raiseUnexpectedRune() {
	s.err = &UnexpectedRuneError{
		Line: s.line,
		Col:  s.col - 1,
		Rune: s.c,
	}
}

func (s *scanner) read() bool {
	s.c, _, s.err = s.r.ReadRune()
	if s.err != nil {
		return false
	}

	switch s.c {
	case '\n':
		s.col = 1
		s.line++
	default:
		s.col++
	}
	return true
}

func (s *scanner) unread() {
	err := s.r.UnreadRune()
	if err != nil {
		panic(err) // If this happens, there's a bug.
	}
}

type stateFunc func() stateFunc

func (s *scanner) start() stateFunc {
	s.tok.Line = s.line
	s.tok.Col = s.col
	if !s.read() {
		return nil
	}
	s.buf.Reset()

	switch s.c {
	case '(':
		s.tok.Val = Lparen{}
		return s.start
	case ')':
		s.tok.Val = Rparen{}
		return s.start
	case '"':
		return s.string
	case '\'':
		return s.rune
	case '_':
		s.buf.WriteByte('_')
		return s.ident
	}

	if unicode.IsSpace(s.c) {
		return s.start
	}
	if s.c >= '0' && s.c <= '9' {
		s.buf.WriteRune(s.c)
		return s.int
	}
	if (s.c >= 'a' && s.c <= 'z') || (s.c >= 'A' && s.c <= 'Z') {
		s.buf.WriteRune(s.c)
		return s.ident
	}

	s.raiseUnexpectedRune()
	return nil
}

func (s *scanner) int() stateFunc {
	if !s.read() {
		return nil
	}

	if s.c == '.' {
		s.buf.WriteByte('.')
		return s.float
	}
	if s.c >= '0' && s.c <= '9' {
		s.buf.WriteRune(s.c)
		return s.int
	}

	str := s.buf.String()
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		s.raiseToken(fmt.Errorf("parse integer literal: %w", err))
	}
	s.tok.Val = Int(v)

	s.unread()
	return s.start
}

func (s *scanner) float() stateFunc {
	if !s.read() {
		return nil
	}

	if s.c >= '0' && s.c <= '9' {
		s.buf.WriteRune(s.c)
		return s.float
	}

	str := s.buf.String()
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		s.raiseToken(fmt.Errorf("parse float literal: %w", err))
	}
	s.tok.Val = Float(v)

	s.unread()
	return s.start
}

func (s *scanner) string() stateFunc {
	if !s.read() {
		if errors.Is(s.err, io.EOF) {
			s.raiseToken(errors.New("EOF in string literal"))
		}
		return nil
	}

	switch s.c {
	case '\\':
		if !s.read() {
			if errors.Is(s.err, io.EOF) {
				s.raiseToken(errors.New("EOF in string literal"))
			}
			return nil
		}
		v, ok := escape(s.c, '"')
		if !ok {
			s.raiseToken(fmt.Errorf("invalid escape sequence %q", s.c))
			return nil
		}
		s.buf.WriteRune(v)
		return s.string

	case '"':
		return s.start

	default:
		s.buf.WriteRune(s.c)
		return s.string
	}
}

func (s *scanner) rune() stateFunc {
	var val rune

	if !s.read() {
		if errors.Is(s.err, io.EOF) {
			s.raiseToken(errors.New("EOF in rune literal"))
		}
		return nil
	}

	switch s.c {
	case '\\':
		if !s.read() {
			if errors.Is(s.err, io.EOF) {
				s.raiseToken(errors.New("EOF in rune literal"))
			}
			return nil
		}
		v, ok := escape(s.c, '\'')
		if !ok {
			s.raiseToken(fmt.Errorf("invalid escape sequence %q", s.c))
			return nil
		}
		val = v

	case '\'':
		s.raiseToken(errors.New("empty rune literal"))
		return nil

	default:
		val = s.c
	}

	if !s.read() {
		if errors.Is(s.err, io.EOF) {
			s.raiseToken(errors.New("EOF in rune literal"))
		}
		return nil
	}
	if s.c != '\'' {
		s.raiseToken(errors.New("rune literal contains more than one rune"))
		return nil
	}

	s.tok.Val = Int(val)
	return s.start
}

func (s *scanner) ident() stateFunc {
	if !s.read() {
		return nil
	}

	switch s.c {
	case '_':
		s.buf.WriteByte('_')
		return s.ident
	case '?', '!':
		s.buf.WriteRune(s.c)
		return s.start
	}

	if (s.c >= 'a' && s.c <= 'z') || (s.c >= 'A' && s.c <= 'Z') {
		s.buf.WriteRune(s.c)
		return s.ident
	}

	s.unread()
	s.tok.Val = Ident(s.buf.String())
	return s.start
}

type Token struct {
	Line, Col int
	Val       any
}

type Lparen struct{}
type Rparen struct{}
type Int int64
type Float float64
type tring string
type Ident string
type Oper string

type UnexpectedRuneError struct {
	Line, Col int
	Rune      rune
}

func (err *UnexpectedRuneError) Error() string {
	return fmt.Sprintf("unexpected rune %q (%v:%v)", err.Rune, err.Line, err.Col)
}

type TokenError struct {
	Line, Col int
	Err       error
}

func (err *TokenError) Error() string {
	return fmt.Sprintf("error in token (%v:%v): %v", err.Line, err.Col, err.Err)
}

func (err *TokenError) Unwrap() error {
	return err.Err
}

func escape(c rune, q rune) (rune, bool) {
	switch c {
	case q, '\\':
		return c, true
	case 'n':
		return '\n', true
	case 't':
		return '\t', true
	default:
		return 0, false
	}
}
