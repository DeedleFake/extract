// Package extract implements the core of the Extract language.
package extract

import (
	"fmt"
	"iter"
	"reflect"
	"unique"
)

// Pinned is an identifier that has been pinned. This is used to
// signal during pattern matching that the value of an identifier
// should be matched against instead of simply binding the identifier
// to a new value.
type Pinned struct {
	Ident Ident
}

// Eval returns an error every time because a Pinned should never
// actually be used as an expression.
func (p Pinned) Eval(env *Env, args *List) (*Env, any) {
	return env, fmt.Errorf("pinned ident %q used as expression", p.Ident)
}

// Call is a function call. It calls the first element of the
// underlying list with the remainder of the list as arguments. If the
// list is empty, it just returns the list.
type Call struct {
	*List
}

func (call Call) Eval(env *Env, args *List) (*Env, any) {
	if call.Len() == 0 {
		return env, call
	}

	env, r := Eval(env, call.Head(), call.Tail())
	if args.Len() == 0 {
		return env, r
	}
	return Eval(env, r, args)
}

// Ident is an identifier for bound data, i.e. a declared
// variable/function.
type Ident struct {
	h unique.Handle[string]
}

// MakeIdent returns a new Ident for the given string. It has the
// exact same semantics as [MakeAtom].
func MakeIdent(str string) Ident {
	return Ident{
		h: unique.Make(str),
	}
}

func (ident Ident) Eval(env *Env, args *List) (*Env, any) {
	c, ok := env.Lookup(ident)
	if !ok {
		return env, &NameError{Ident: ident}
	}
	if c, ok := c.(Ident); ok && c == ident {
		panic(fmt.Errorf("name %q is bound to itself", ident))
	}
	return Eval(env, c, args)
}

func (ident Ident) String() string {
	return ident.h.Value()
}

// Ref is an access of an identifier namespaced with a module.
type Ref struct {
	// In the module that the identifier is being accessed inside of. It
	// can be any expression but it must return an atom or an error.
	In any

	// Name is the identifier being accessed.
	Name Ident
}

func (ref Ref) Eval(env *Env, args *List) (*Env, any) {
	env, in := Eval(env, ref.In, nil)
	switch in := in.(type) {
	case Atom:
		m := env.GetModule(in)
		if m == nil {
			return env, &UndefinedModuleError{Name: in}
		}
		v, ok := m.Lookup(ref.Name)
		if !ok {
			return env, &NameError{Ident: ref.Name}
		}
		return Eval(env, v, args)

	case error:
		return env, in

	default:
		return env, NewTypeError(in, reflect.TypeFor[Atom]())
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
	return fmt.Sprintf("%q is not bound", err.Ident)
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
func Eval(env *Env, expr any, args *List) (*Env, any) {
	switch expr := expr.(type) {
	case Evaluator:
		return expr.Eval(env, args)
	default:
		if args.Len() > 0 {
			expr = args.Push(expr)
		}
		return env, expr
	}
}

// EvalAllWithRuntime is like [EvalAll], but also yields the [Env]
// that results from each elements evaluation.
func EvalAllWithRuntime[T any](env *Env, seq iter.Seq[T]) iter.Seq2[*Env, any] {
	return func(yield func(*Env, any) bool) {
		for v := range seq {
			var ret any
			env, ret = Eval(env, v, nil)
			if !yield(env, ret) {
				return
			}
		}
	}
}

// EvalAll returns an iterator that evaluates each element in seq
// using [Eval] and yields the results. It uses r as the base
// [Env] for the evaluation and updates it with the result of each
// elements evaluation.
func EvalAll[T any](env *Env, seq iter.Seq[T]) iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, v := range EvalAllWithRuntime(env, seq) {
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
	Eval(env *Env, args *List) (*Env, any)
}

// EvalFunc is a func wrapper for [Evaluator].
type EvalFunc func(env *Env, args *List) (*Env, any)

func (f EvalFunc) Eval(env *Env, args *List) (*Env, any) {
	return f(env, args)
}

// Run runs a list like it's the body of a function. If any elements
// of the list return an error when evaluated, this function returns
// early with that error. Otherwise, it returns the result of the
// evaluation of the last element of the list.
func Run[T any](env *Env, seq iter.Seq[T]) (e *Env, ret any) {
	for v := range seq {
		env, ret = Eval(env, v, nil)
		if err, ok := ret.(error); ok {
			return env, err
		}
	}
	return env, ret
}

// Equaler is implemented by types that want to define custom
// equality.
type Equaler interface {
	Equal(any) bool
}

// IsEquatable returns true if val is capable of being equated.
func IsEquatable(val any) bool {
	if _, ok := val.(Equaler); ok {
		return true
	}
	return reflect.TypeOf(val).Comparable()
}

// Equal returns true if one of the following is true, in order:
//
// * v1 is an Equaler and v1.Equal(v2)
// * v2 is an Equaler and v2.Equal(v1)
// * v1 == v2
//
// If the last step is reached and either type is not comparable, the
// result is false.
func Equal(v1, v2 any) bool {
	if v1, ok := v1.(Equaler); ok {
		return v1.Equal(v2)
	}
	if v2, ok := v2.(Equaler); ok {
		return v2.Equal(v1)
	}

	return reflect.TypeOf(v1).Comparable() && v1 == v2
}
