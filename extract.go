// Package extract implements the core of the Extract language.
package extract

import (
	"fmt"
	"iter"
	"reflect"
	"unique"
)

// Ident is an identifier for bound data, i.e. a declared variable/function.
//
// The parser will create these for identifiers in the source code.
type Ident string

func (ident Ident) Eval(r *Runtime, args *List) (*Runtime, any) {
	c, ok := r.Lookup(ident)
	if !ok {
		return r, &NameError{Ident: ident}
	}
	if c, ok := c.(Ident); ok && c == ident {
		panic(fmt.Errorf("name %q is bound to itself", string(ident)))
	}
	return Eval(r, c, args)
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

func (ref Ref) Eval(r *Runtime, args *List) (*Runtime, any) {
	r, in := Eval(r, ref.In, nil)
	switch in := in.(type) {
	case Atom:
		m := r.GetModule(in)
		if m == nil {
			return r, &UndefinedModuleError{Name: in}
		}
		v, ok := m.Lookup(ref.Name)
		if !ok {
			return r, &NameError{Ident: ref.Name}
		}
		return Eval(r, v, args)

	case error:
		return r, in

	default:
		return r, NewTypeError(in, reflect.TypeFor[Atom]())
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
func Eval(r *Runtime, expr any, args *List) (*Runtime, any) {
	switch expr := expr.(type) {
	case Evaluator:
		return expr.Eval(r, args)
	default:
		if args.Len() > 0 {
			expr = args.Push(expr)
		}
		return r, expr
	}
}

// EvalAllWithRuntime is like [EvalAll], but also yields the [Runtime]
// that results from each elements evaluation.
func EvalAllWithRuntime[T any](r *Runtime, seq iter.Seq[T]) iter.Seq2[*Runtime, any] {
	return func(yield func(*Runtime, any) bool) {
		for v := range seq {
			var ret any
			r, ret = Eval(r, v, nil)
			if !yield(r, ret) {
				return
			}
		}
	}
}

// EvalAll returns an iterator that evaluates each element in seq
// using [Eval] and yields the results. It uses r as the base
// [Runtime] for the evaluation and updates it with the result of each
// elements evaluation.
func EvalAll[T any](r *Runtime, seq iter.Seq[T]) iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, v := range EvalAllWithRuntime(r, seq) {
			if !yield(v) {
				return
			}
		}
	}
}

// Evaluator is a value that can be evaluated, possibly with
// arguments, such as a function.
type Evaluator interface {
	// Eval evaluates the value in the given [Runtime] with the given
	// arguments. It returns the result of the evaluation and a new
	// Runtime representing any modifications that the evaluation has
	// made to it.
	//
	// Most implementations will simply return the Runtime unmodified.
	Eval(r *Runtime, args *List) (*Runtime, any)
}

// EvalFunc is a func wrapper for [Evaluator].
type EvalFunc func(r *Runtime, args *List) (*Runtime, any)

func (f EvalFunc) Eval(r *Runtime, args *List) (*Runtime, any) {
	return f(r, args)
}
