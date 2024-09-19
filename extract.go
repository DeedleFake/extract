// Package extract implements the core of the Extract language.
package extract

type Atom string

type Ident string

type Ref struct {
	In   any
	Name any
}
