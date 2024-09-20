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

// Ident is an identifier for bound data, i.e. a declared variable/function.
//
// The parser will create these for identifiers in the source code.
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

// Ref is an access of an identifier namespaced with a module.
//
// The parser will create these for expressions of the form `a.b`.
type Ref struct {
	// In the module that the identifier is being accessed inside of. It
	// can be any expression but it must return an atom or an error.
	In any

	// Name is the identifier being accessed.
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

// Atom is an interned string. Atoms are comparable and are very
// efficient to compare, but slightly less efficient to create at
// runtime or to convert back to a string.
//
// The parser will automatically create these from atom literals.
type Atom struct {
	h unique.Handle[string]
}

// MakeAtom returns an atom representing the given string. The
// returned atom will be equal to all other atoms returned from this
// function when called with the same string.
func MakeAtom(str string) Atom {
	return Atom{h: unique.Make(str)}
}

// String gets the string value that the atom was created from.
func (atom Atom) String() string {
	return atom.h.Value()
}

// ArgumentNumError is returned when a function is called with the
// wrong number of arguments. If the function has a specific number of
// arguments that it expects, Expected will be >= 0.
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

// TypeError is returned by expressions that have incorrect types in
// them in some way. Val is the value that is of the wrong type. If
// there is information about types that were expected, the Expected
// field will contain it.
type TypeError struct {
	Val      any
	Expected []reflect.Type
}

// NewTypeError is a convience function that creates a new TypeError.
func NewTypeError(val any, expected ...reflect.Type) *TypeError {
	return &TypeError{
		Val:      val,
		Expected: expected,
	}
}

func (err *TypeError) Error() string {
	if len(err.Expected) == 0 {
		return fmt.Sprintf("incorrect type %T", err.Val)
	}
	return fmt.Sprintf("incorrect type %T, expected one of %v", err.Val, err.Expected)
}

// NameError is returned when an identifier was accessed but is not
// bound in the scope.
type NameError struct {
	Ident Ident
}

func (err *NameError) Error() string {
	return fmt.Sprintf("%q is not bound", string(err.Ident))
}

// UndefinedModuleError is returned when an attempt is made to access
// a module that has not been defined.
type UndefinedModuleError struct {
	Name Atom
}

func (err *UndefinedModuleError) Error() string {
	return fmt.Sprintf("module %q not found in runtime", err.Name)
}

// Eval evaluates a value, potentially passing arguments to it. If the
// value implements [Evaluator], its Eval method is called. If not and
// arguments were provided, the value is returned as the first element
// of a list containing it and the arguments provided. Otherwise, the
// value is returned unmodified.
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

// EvalAllWithContext is like [EvalAll], but also yields the context
// that results from each elements evaluation.
func EvalAllWithContext[T any](ctx context.Context, seq iter.Seq[T]) iter.Seq2[any, context.Context] {
	return func(yield func(any, context.Context) bool) {
		for v := range seq {
			var r any
			r, ctx = Eval(ctx, v, nil)
			if !yield(r, ctx) {
				return
			}
		}
	}
}

// EvalAll returns an iterator that evaluates each element in seq
// using [Eval] and yields the results. It uses ctx as the base
// context for the evaluation and updates it with the result of each
// elements evaluation.
func EvalAll[T any](ctx context.Context, seq iter.Seq[T]) iter.Seq[any] {
	return func(yield func(any) bool) {
		for v := range EvalAllWithContext(ctx, seq) {
			if !yield(v) {
				return
			}
		}
	}
}

// Evaluator is a value that can be evaluated, possibly with
// arguments, such as a function.
type Evaluator interface {
	// Eval evaluates the value in the given context with the given
	// arguments. It returns the result of the evaluation and a new
	// context representing any modifications that the evaluation has
	// made to it.
	//
	// Most implementations will simply return the context unmodified.
	Eval(ctx context.Context, args *List) (any, context.Context)
}

// EvalFunc is a func wrapper for [Evaluator].
type EvalFunc func(ctx context.Context, args *List) (any, context.Context)

func (f EvalFunc) Eval(ctx context.Context, args *List) (any, context.Context) {
	return f(ctx, args)
}
