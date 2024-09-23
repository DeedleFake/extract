package extract

import (
	"fmt"

	"deedles.dev/xiter"
)

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
