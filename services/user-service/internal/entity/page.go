package entity

type Page[T any] struct {
	Items      []T
	TotalCount int
}
