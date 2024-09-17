// Package scanner implements a scanner for Extract tokens.
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

type Scanner struct {
	r         *bufio.Reader
	line, col int
	c         rune
	err       error

	buf strings.Builder
	tok Token
}

func New(r io.Reader) *Scanner {
	return &Scanner{
		r:    bufio.NewReader(r),
		line: 1, col: 1,
	}
}

func (s *Scanner) Scan() bool {
	s.start()
	return s.err == nil
}

func (s *Scanner) Token() Token {
	return s.tok
}

func (s *Scanner) Err() error {
	if errors.Is(s.err, io.EOF) {
		return nil
	}
	return s.err
}

func (s *Scanner) All() iter.Seq[Token] {
	return func(yield func(Token) bool) {
		for s.Scan() {
			if !yield(s.Token()) {
				return
			}
		}
	}
}

type raise struct{ err error }

func (s *Scanner) raise(err error) {
	panic(raise{err: err})
}

func (s *Scanner) raiseToken(err error) {
	s.raise(&TokenError{
		Line: s.tok.Line,
		Col:  s.tok.Col,
		Err:  err,
	})
}

func (s *Scanner) raiseUnexpectedRune() {
	s.raise(&UnexpectedRuneError{
		Line: s.line,
		Col:  s.col - 1,
		Rune: s.c,
	})
}

func (s *Scanner) raiseUnexpectedEOF(literal string) {
	if errors.Is(s.err, io.EOF) {
		s.raiseToken(fmt.Errorf("%w in %v literal", io.ErrUnexpectedEOF, literal))
	}
}

func (s *Scanner) read() bool {
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

func (s *Scanner) unread() {
	err := s.r.UnreadRune()
	if err != nil {
		panic(err) // If this happens, there's a bug.
	}
}

func (s *Scanner) start() {
	defer func() {
		switch r := recover().(type) {
		case nil:
		case raise:
			s.err = r.err
		default:
			panic(r)
		}
	}()

	defer s.buf.Reset()

	s.tok.Line = s.line
	s.tok.Col = s.col

	for {
		if !s.read() {
			return
		}
		if !unicode.IsSpace(s.c) {
			break
		}
	}

	switch s.c {
	case '"':
		s.string()
		return
	case ':':
		s.atomcolon()
		return
	case '\'':
		s.rune()
		return
	case '_':
		s.buf.WriteByte('_')
		s.ident()
		return
	}

	if s.c >= '0' && s.c <= '9' {
		s.buf.WriteRune(s.c)
		s.int()
		return
	}
	if s.c >= 'a' && s.c <= 'z' {
		s.buf.WriteRune(s.c)
		s.ident()
		return
	}
	if s.c >= 'A' && s.c <= 'Z' {
		s.buf.WriteRune(s.c)
		s.atom()
		return
	}
	if maybeOper(s.c) {
		s.buf.WriteRune(s.c)
		s.oper()
		return
	}

	s.raiseUnexpectedRune()
}

func (s *Scanner) atomcolon() {
	if !s.read() {
		s.raiseUnexpectedEOF("atom")
		return
	}

	switch s.c {
	case '"':
		s.string()
		s.tok.Val = Atom(s.tok.Val.(String))
		return

	default:
		s.unread()
		s.atom()
		return
	}
}

func (s *Scanner) atom() {
	s.ident()
	s.tok.Val = Atom(s.tok.Val.(Ident))
}

func (s *Scanner) int() {
	for {
		if !s.read() {
			break
		}

		if s.c == '.' {
			s.buf.WriteByte('.')
			s.float()
			return
		}
		if s.c >= '0' && s.c <= '9' {
			s.buf.WriteRune(s.c)
			continue
		}

		s.unread()
		break
	}

	str := s.buf.String()
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		s.raiseToken(fmt.Errorf("parse integer literal: %w", err))
	}
	s.tok.Val = Int(v)
}

func (s *Scanner) float() {
	for {
		if !s.read() {
			return
		}

		if s.c >= '0' && s.c <= '9' {
			s.buf.WriteRune(s.c)
			continue
		}

		s.unread()
		break
	}

	str := s.buf.String()
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		s.raiseToken(fmt.Errorf("parse float literal: %w", err))
	}
	s.tok.Val = Float(v)
}

func (s *Scanner) string() {
	for {
		if !s.read() {
			s.raiseUnexpectedEOF("string")
			return
		}

		switch s.c {
		case '\\':
			if !s.read() {
				s.raiseUnexpectedEOF("string")
				return
			}
			s.escape('"')
			s.buf.WriteRune(s.c)

		case '"':
			s.tok.Val = String(s.buf.String())
			return

		default:
			s.buf.WriteRune(s.c)
		}
	}
}

func (s *Scanner) rune() {
	if !s.read() {
		s.raiseUnexpectedEOF("rune")
		return
	}

	var val rune
	switch s.c {
	case '\\':
		if !s.read() {
			s.raiseUnexpectedEOF("rune")
			return
		}
		s.escape('\'')
		val = s.c

	case '\'':
		s.raiseToken(errors.New("empty rune literal"))
		return

	default:
		val = s.c
	}

	if !s.read() {
		s.raiseUnexpectedEOF("rune")
		return
	}
	if s.c != '\'' {
		s.raiseToken(errors.New("rune literal contains more than one rune"))
		return
	}

	s.tok.Val = Int(val)
}

func (s *Scanner) ident() {
	for {
		if !s.read() {
			return
		}

		switch s.c {
		case '_':
			s.buf.WriteRune(s.c)
			continue
		case '?', '!':
			s.buf.WriteRune(s.c)
			break
		}

		if (s.c >= 'a' && s.c <= 'z') || (s.c >= 'A' && s.c <= 'Z') || (s.c >= '0' && s.c <= '9') {
			s.buf.WriteRune(s.c)
			continue
		}

		s.unread()
		break
	}

	s.tok.Val = Ident(s.buf.String())
}

func (s *Scanner) oper() {
	// This has its own state to make it easier to potentially support
	// longer operators later.
	s.tok.Val = Oper(s.buf.String())
}

func (s *Scanner) escape(q rune) {
	switch s.c {
	case q, '\\':
	case 'n':
		s.c = '\n'
	case 't':
		s.c = '\t'
	default:
		s.raiseToken(fmt.Errorf("invalid escape sequence %q", s.c))
	}
}

// Token is an Extract language parser token. If the token is valid,
// Val will be one of the token types defined in this package.
type Token struct {
	Line, Col int
	Val       any
}

// Token value type.
type (
	Int    int64
	Float  float64
	String string
	Ident  string
	Oper   string
	Atom   string
)

// UnexpectedRuneError is yielded when an unexpected rune is found
// during the course of scanning.
type UnexpectedRuneError struct {
	Line, Col int
	Rune      rune
}

func (err *UnexpectedRuneError) Error() string {
	return fmt.Sprintf("unexpected rune %q (%v:%v)", err.Rune, err.Line, err.Col)
}

// TokenError is yielded when an unexpected error occurs during the
// scanning of a token. Line and Col are for the beginning of the
// token, not the exact location of the error.
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
