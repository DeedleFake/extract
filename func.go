package extract

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"deedles.dev/xiter"
)

var ErrPatternMatch = errors.New("arguments did not match defined patterns")

type funcVariant struct {
	Pattern *Pattern
	Body    *List
}

type Func struct {
	m        sync.RWMutex
	env      *Env
	name     Ident
	variants []funcVariant
}

func NewFunc(env *Env, name Ident, pattern *Pattern, body *List) *Func {
	f := Func{
		name:     name,
		variants: []funcVariant{{Pattern: pattern, Body: body}},
	}
	f.env = env.Let(name, &f)
	return &f
}

func (f *Func) Eval(env *Env, args *List) (*Env, any) {
	f.m.RLock()
	defer f.m.RUnlock()

	eargs := CollectList(EvalAll(env, args.All()))
	for _, variant := range f.variants {
		if fenv, ok := variant.Pattern.Match(f.env, eargs); ok {
			_, r := Run(fenv, variant.Body.All())
			return env, r
		}
	}
	return env, ErrPatternMatch
}

func (f *Func) AddVariant(pattern *Pattern, body *List) {
	f.m.Lock()
	defer f.m.Unlock()

	f.variants = append(f.variants, funcVariant{Pattern: pattern, Body: body})
}

func compileFuncPattern(pattern any) (name Ident, cpattern *Pattern, err error) {
	switch pattern := pattern.(type) {
	case Call:
		if pattern.Len() == 0 {
			return Ident{}, nil, errors.New("function pattern list must contain at least one element")
		}

		name, ok := pattern.Head().(Ident)
		if !ok {
			return Ident{}, nil, NewTypeError(name, reflect.TypeFor[Ident]())
		}

		cpattern, err := CompilePattern(pattern.Tail())
		if err != nil {
			return name, nil, err
		}

		return name, cpattern, nil

	default:
		return Ident{}, nil, NewTypeError(pattern, reflect.TypeFor[*List](), reflect.TypeFor[Ident]())
	}
}

type Pattern struct {
	root matcher
}

func (p *Pattern) Match(env *Env, val any) (*Env, bool) {
	return p.root(env, val)
}

type matcher func(env *Env, val any) (*Env, bool)

func CompilePattern(format any) (*Pattern, error) {
	root, err := compilePattern(format)
	return &Pattern{root: root}, err
}

func compilePattern(format any) (matcher, error) {
	switch format := format.(type) {
	case Atom, int64, float64, string:
		return equalityMatcher(format), nil
	case Ident:
		return assignMatcher(format), nil
	case Call:
		return listMatcher(format.List)
	case *List:
		return listMatcher(format)
	default:
		return nil, fmt.Errorf("unexpected type %T in pattern", format)
	}
}

func equalityMatcher[T comparable](val T) matcher {
	return func(env *Env, v any) (*Env, bool) {
		return env, val == v
	}
}

func assignMatcher(name Ident) matcher {
	return func(env *Env, val any) (*Env, bool) {
		return env.Let(name, val), true
	}
}

func listMatcher(list *List) (matcher, error) {
	matchers := make([]matcher, 0, list.Len())
	for part := range list.All() {
		matcher, err := compilePattern(part)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, matcher)
	}

	return func(env *Env, val any) (_ *Env, ok bool) {
		vlist, ok := val.(*List)
		if !ok || vlist.Len() != len(matchers) {
			return env, false
		}

		for i, v := range xiter.Enumerate(vlist.All()) {
			env, ok = matchers[i](env, v)
			if !ok {
				return env, false
			}
		}
		return env, true
	}, nil
}
