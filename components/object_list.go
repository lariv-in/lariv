package components

type ObjectList[T any] struct {
	Items    []T
	Number   uint
	NumPages uint
	Total    uint64
}
