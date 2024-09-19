// Package extract implements the core of the Extract language.
package extract

import (
	"context"
	"errors"
	"fmt"
	"iter"
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

func Eval(ctx context.Context, expr any, args *List) (any, context.Context) {
	switch expr := expr.(type) {
	case Evaluator:
		return expr.Eval(ctx, args)
	default:
		if args.Len() > 0 {
			expr = args.Push(expr)
		}
		return expr, ctx
	}
}

func EvalAllWithContext[T any](ctx context.Context, seq iter.Seq[T]) iter.Seq2[any, context.Context] {
	return func(yield func(any, context.Context) bool) {
		for v := range seq {
			var r any
			r, ctx = Eval(ctx, v, nil)
			if !yield(r, ctx) {
				return
			}
			if _, ok := r.(error); ok {
				return
			}
		}
	}
}

func EvalAll[T any](ctx context.Context, seq iter.Seq[T]) iter.Seq[any] {
	return func(yield func(any) bool) {
		for v := range EvalAllWithContext(ctx, seq) {
			if !yield(v) {
				return
			}
		}
	}
}

type Evaluator interface {
	Eval(ctx context.Context, args *List) (any, context.Context)
}

type EvalFunc func(ctx context.Context, args *List) (any, context.Context)

func (f EvalFunc) Eval(ctx context.Context, args *List) (any, context.Context) {
	return f(ctx, args)
}
