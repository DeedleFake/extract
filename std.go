package extract

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// std is the Extract standard library in the form of a map of module
// names to modules.
var std = map[Atom]*Module{
	MakeAtom("String"): stdString(),
}

func stdString() *Module {
	m := Module{name: MakeAtom("String")}
	m.decls.Store(Ident("to_upper"), EvalFunc(func(r *Runtime, args *List) (*Runtime, any) {
		if args.Len() != 1 {
			return r, &ArgumentNumError{Num: args.Len(), Expected: 1}
		}

		_, head := Eval(r, args.Head(), nil)
		str, ok := head.(string)
		if !ok {
			return r, NewTypeError(head, reflect.TypeFor[string]())
		}

		return r, strings.ToUpper(str)
	}))
	m.decls.Store(Ident("to_lower"), EvalFunc(func(r *Runtime, args *List) (*Runtime, any) {
		if args.Len() != 1 {
			return r, &ArgumentNumError{Num: args.Len(), Expected: 1}
		}

		_, head := Eval(r, args.Head(), nil)
		str, ok := head.(string)
		if !ok {
			return r, NewTypeError(head, reflect.TypeFor[string]())
		}

		return r, strings.ToLower(str)
	}))
	m.decls.Store(Ident("format"), EvalFunc(func(r *Runtime, args *List) (*Runtime, any) {
		if args.Len() == 0 {
			return r, &ArgumentNumError{Num: args.Len(), Expected: -1}
		}

		_, head := Eval(r, args.Head(), nil)
		str, ok := head.(string)
		if !ok {
			return r, NewTypeError(head, reflect.TypeFor[string]())
		}

		verbs := slices.Collect(EvalAll(r, args.Tail().All()))
		return r, fmt.Sprintf(str, verbs...)
	}))

	return &m
}
