// Package extract implements the core of the Extract language.
package extract

import "unique"

type Ident string

type Ref struct {
	In   any
	Name any
}

type Atom struct {
	h unique.Handle[string]
}

func NewAtom(str string) Atom {
	return Atom{h: unique.Make(str)}
}

func (atom Atom) String() string {
	return atom.h.Value()
}
