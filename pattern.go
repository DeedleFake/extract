package extract

type Pattern struct {
	root matcher
}

func (p *Pattern) Match(env *Env, val any) (*Env, bool) {
	return p.root(env, val)
}

type matcher func(env *Env, val any) (*Env, bool)
