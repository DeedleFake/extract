// Package literal defines types created by the parser from literals.
package literal

import (
	"deedles.dev/extract"
)

type Int = int64

type Float = float64

type String = string

type Atom = extract.Atom

type Ident = extract.Ident

type List = extract.Call

type Ref = extract.Ref
