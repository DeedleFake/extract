package scanner

var (
	opers = setOf(
		"(",
		")",
		"+",
		"-",
		"*",
		"/",
		"%",
		"^",
		".",
	)
)

func maybeOper(c rune) bool {
	s := string(c)
	if _, ok := opers[s]; ok {
		return true
	}

	return false
}

func oper(s string) (string, string) {
	if len(s) >= 1 {
		if _, ok := opers[s[:1]]; ok {
			return s[:1], s[1:]
		}
	}

	return "", s
}

func setOf[T comparable](vals ...T) map[T]struct{} {
	m := make(map[T]struct{}, len(vals))
	for _, val := range vals {
		m[val] = struct{}{}
	}
	return m
}
