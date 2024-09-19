package parser

import (
	"fmt"
	"io"

	"deedles.dev/extract"
	"deedles.dev/extract/scanner"
)

func Parse(r io.Reader) (*extract.List, error) {
	return ParseScanner(scanner.New(r))
}

func ParseScanner(s *scanner.Scanner) (*extract.List, error) {
	p := parser{s: s}
	return p.Parse()
}

type parser struct {
	s   *scanner.Scanner
	tok scanner.Token
}

func (p *parser) Parse() (list *extract.List, err error) {
	defer func() {
		switch r := recover().(type) {
		case nil:
		case raise:
			err = r.err
		default:
			panic(r)
		}
	}()

	return p.listInner(), nil
}

type raise struct{ err error }

func (p *parser) raise(err error) {
	panic(raise{err: err})
}

func (p *parser) raiseUnexpectedEOF() {
	p.raise(io.ErrUnexpectedEOF)
}

func (p *parser) raiseUnexpectedToken(got scanner.Token, ex any) {
	p.raise(&UnexpectedTokenError{
		Line:     got.Line,
		Col:      got.Col,
		Got:      got.Val,
		Expected: ex,
	})
}

func (p *parser) scan() scanner.Token {
	if p.tok.Val != nil {
		tok := p.tok
		p.tok.Val = nil
		return tok
	}

	if !p.s.Scan() {
		p.raiseUnexpectedEOF()
		return scanner.Token{}
	}

	return p.s.Token()
}

func (p *parser) unscan(tok scanner.Token) {
	if p.tok.Val != nil {
		panic("unscanned twice")
	}

	p.tok = tok
}

func (p *parser) peek() any {
	if p.tok.Val != nil {
		return p.tok.Val
	}

	if !p.s.Scan() {
		return nil
	}

	p.tok = p.s.Token()
	return p.tok.Val
}

func (p *parser) expect(tok any) scanner.Token {
	got := p.scan()
	if got.Val != tok {
		p.raiseUnexpectedToken(got, tok)
		return scanner.Token{}
	}
	return got
}

func expect[T any](p *parser) (tok scanner.Token, v T) {
	got := p.scan()
	if v, ok := got.Val.(T); ok {
		return got, v
	}

	p.raiseUnexpectedToken(got, nil)
	return tok, v
}

func (p *parser) list() *extract.List {
	expect[scanner.Lparen](p)
	list := p.listInner()
	expect[scanner.Rparen](p)
	return list
}

func (p *parser) listInner() *extract.List {
	var exprs []any
	for p.peek() != (scanner.Rparen{}) && p.peek() != nil {
		exprs = append(exprs, p.expr())
	}
	return extract.ListOf(exprs...)
}

func (p *parser) expr() any {
	tok := p.scan()
	switch t := tok.Val.(type) {
	case scanner.Int:
		return extract.Int(t)
	case scanner.String:
		return extract.String(t)
	case scanner.Atom:
		p.unscan(tok)
		return p.atom()
	case scanner.Lparen:
		p.unscan(tok)
		return p.list()
	}

	p.raiseUnexpectedToken(p.scan(), nil)
	return nil
}

func (p *parser) atom() any {
	_, atom := expect[scanner.Atom](p)
	if p.peek() == (scanner.Dot{}) {
		return p.moduleident(extract.Atom(atom))
	}
	return extract.Atom(atom)
}

func (p *parser) moduleident(module any) extract.ModuleIdent {
	expect[scanner.Dot](p)
	_, ident := expect[scanner.Ident](p)
	return extract.ModuleIdent{
		Module: module,
		Ident:  extract.Ident(ident),
	}
}

type UnexpectedTokenError struct {
	Line, Col int
	Got       any
	Expected  any
}

func (err *UnexpectedTokenError) Error() string {
	if err.Expected == nil {
		return fmt.Sprintf("unexpected token %q (%[1]T) at %v:%v", err.Got, err.Line, err.Col)
	}
	return fmt.Sprintf("unexpected token %q (%[1]T) at %v:%v, expected %q (%[4]T)", err.Got, err.Line, err.Col, err.Expected)
}
