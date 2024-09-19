// Package extract implements the core of the Extract language.
package extract

import (
	"context"
	"fmt"
	"unique"
)

type Ident string

func (ident Ident) Value(ctx context.Context) (any, error) {
	c := ctx.Value(ident)
	if c == nil {
		return nil, &NameError{Ident: ident}
	}
	return c, nil
}

type NameError struct {
	Ident Ident
}

func (err *NameError) Error() string {
	return fmt.Sprintf("%q is not bound", string(err.Ident))
}

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
