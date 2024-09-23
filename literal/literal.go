// Package literal defines types created by the parser from literals.
package literal

import (
	"deedles.dev/extract"
)

// Int is created from integer literal expressions such as 2 or -5.
type Int = int64

// Float is created from float literal expressions such as 2.0 or
// -1.3.
type Float = float64

// String is created from string literal expressions such as
// "example".
type String = string

// Atom is created from atom literal expressions such as :example,
// :"example with spaces", or Example.
type Atom = extract.Atom

// Ident is created from identifiers.
type Ident = extract.Ident

// List is created from list literal expressions such as (a b c). The
// elements of the list will be other types in this package.
type List = extract.Call

// Ref is created from module references such as Example.function.
type Ref = extract.Ref

// Pin is created from usages of the pin operator before an
// identifier. It looks like \ident.
type Pin = extract.Pinned
