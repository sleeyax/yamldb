package yamldb

type OrderFunc func(a, b string) bool

// OrderAlphabetically orders keys from a -> z.
func OrderAlphabetically(a, b string) bool { return a < b }

// OrderAlphabeticallyReversed orders keys from z -> a.
func OrderAlphabeticallyReversed(a, b string) bool { return a > b }
