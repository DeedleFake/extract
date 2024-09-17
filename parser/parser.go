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

func (p *parser) expect(tok any) {
	got := p.scan()
	if got != tok {
		p.raiseUnexpectedToken(got, tok)
		return
	}
}

func (p *parser) list() *extract.List {
	p.expect(scanner.Oper("("))
	list := p.listInner()
	p.expect(scanner.Oper(")"))
	return list
}

func (p *parser) listInner() *extract.List {
	var exprs []any
	for p.peek() != scanner.Oper(")") && p.peek() != nil {
		exprs = append(exprs, p.expr())
	}
	return extract.ListFrom(exprs...)
}

func (p *parser) expr() any {
	tok := p.peek()
	switch tok := tok.(type) {
	case scanner.Int:
		return extract.Int(tok)
	case scanner.String:
		return extract.String(tok)
	case scanner.Atom:
		// TODO: Implement uniqification.
		return extract.Atom(tok)
	}

	switch tok {
	case scanner.Oper("("):
		return p.list()
	}

	p.raiseUnexpectedToken(p.scan(), nil)
	return nil
}

type UnexpectedTokenError struct {
	Line, Col int
	Got       any
	Expected  any
}

func (err *UnexpectedTokenError) Error() string {
	if err.Expected == nil {
		return fmt.Sprintf("unexpected token %q at %v:%v", err.Got, err.Line, err.Col)
	}
	return fmt.Sprintf("unexpected token %q at %v:%v, expected %q", err.Got, err.Line, err.Col, err.Expected)
}
