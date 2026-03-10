package components

type ObjectList[T any] struct {
	Items    []T
	Number   int
	NumPages int
	Total    int64
}
