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
	m.decls = map[Ident]any{
		MakeIdent("to_upper"): EvalFunc(func(env *Env, args *List) (*Env, any) {
			if args.Len() != 1 {
				return env, &ArgumentNumError{Num: args.Len(), Expected: 1}
			}

			_, head := Eval(env, args.Head(), nil)
			str, ok := head.(string)
			if !ok {
				return env, NewTypeError(head, reflect.TypeFor[string]())
			}

			return env, strings.ToUpper(str)
		}),
		MakeIdent("to_lower"): EvalFunc(func(env *Env, args *List) (*Env, any) {
			if args.Len() != 1 {
				return env, &ArgumentNumError{Num: args.Len(), Expected: 1}
			}

			_, head := Eval(env, args.Head(), nil)
			str, ok := head.(string)
			if !ok {
				return env, NewTypeError(head, reflect.TypeFor[string]())
			}

			return env, strings.ToLower(str)
		}),
		MakeIdent("format"): EvalFunc(func(env *Env, args *List) (*Env, any) {
			if args.Len() == 0 {
				return env, &ArgumentNumError{Num: args.Len(), Expected: -1}
			}

			_, head := Eval(env, args.Head(), nil)
			str, ok := head.(string)
			if !ok {
				return env, NewTypeError(head, reflect.TypeFor[string]())
			}

			verbs := slices.Collect(EvalAll(env, args.Tail().All()))
			return env, fmt.Sprintf(str, verbs...)
		}),
	}

	return &m
}
