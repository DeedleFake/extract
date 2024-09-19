// Package extract implements the core of the Extract language.
package extract

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"unique"
)

type Ident string

func (ident Ident) Eval(ctx context.Context, args *List) (any, context.Context) {
	c := ctx.Value(ident)
	if c == nil {
		return &NameError{Ident: ident}, ctx
	}
	if c, ok := c.(Ident); ok && c == ident {
		panic(fmt.Errorf("name %q is bound to itself", string(ident)))
	}
	return Eval(ctx, c, args)
}

type Ref struct {
	In   any
	Name Ident
}

func (r Ref) Eval(ctx context.Context, args *List) (any, context.Context) {
	in, ctx := Eval(ctx, r.In, nil)
	switch in := in.(type) {
	case Atom:
		runtime := GetRuntime(ctx)
		if runtime == nil {
			panic(errors.New("no runtime in context"))
		}
		m := runtime.GetModule(in)
		if m == nil {
			return &UndefinedModuleError{Name: in}, ctx
		}
		v, ok := m.Lookup(r.Name)
		if !ok {
			return &NameError{Ident: r.Name}, ctx
		}
		return Eval(ctx, v, args)

	case error:
		return in, ctx

	default:
		return NewTypeError(in, reflect.TypeFor[Atom]()), ctx
	}
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

type ArgumentNumError struct {
	Num      int
	Expected int
}

func (err *ArgumentNumError) Error() string {
	if err.Expected < 0 {
		return fmt.Sprintf("incorrect number of arguments %v", err.Num)
	}
	return fmt.Sprintf("incorrect number of arguments %v, expected %v", err.Num, err.Expected)
}

type TypeError struct {
	Val      any
	Expected []reflect.Type
}

func NewTypeError(val any, expected ...reflect.Type) *TypeError {
	return &TypeError{
		Val:      val,
		Expected: expected,
	}
}

func (err *TypeError) Error() string {
	return fmt.Sprintf("incorrect type %T, expected one of %v", err.Val, err.Expected)
}

type NameError struct {
	Ident Ident
}

func (err *NameError) Error() string {
	return fmt.Sprintf("%q is not bound", string(err.Ident))
}

type UndefinedModuleError struct {
	Name Atom
}

func (err *UndefinedModuleError) Error() string {
	return fmt.Sprintf("module %q not found in runtime", err.Name)
}
