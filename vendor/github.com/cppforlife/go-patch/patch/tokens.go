package patch

type Token interface{}

type RootToken struct{}

type IndexToken struct {
	Index int
}

type AfterLastIndexToken struct{}

type MatchingIndexToken struct {
	Key   string
	Value string

	Optional bool
}

type KeyToken struct {
	Key string

	Optional bool
}
