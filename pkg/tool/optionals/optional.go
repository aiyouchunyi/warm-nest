// Package optionals @Author larry
// File optionals.go
// @Date 2024/8/15 09:37:00
// @Desc 结构体工具
package optionals

type Optional[T any] struct {
	value *T
}

func Of[T any](value T) Optional[T] {
	return Optional[T]{value: &value}
}

func OfPtr[T any](value *T) Optional[T] {
	return Optional[T]{value: value}
}

func Empty[T any]() Optional[T] {
	return Optional[T]{value: nil}
}

func (o Optional[T]) IsPresent() bool {
	return o.value != nil
}

func (o Optional[T]) IsEmpty() bool {
	return o.value == nil
}

func (o Optional[T]) Get() T {
	if o.value == nil {
		var zero T
		return zero
	}
	return *o.value
}

func (o Optional[T]) GetPtr() *T {
	if o.value == nil {
		var zero T
		return &zero
	}
	return o.value
}

func (o Optional[T]) OrElse(defaultValue T) T {
	if o.value == nil {
		return defaultValue
	}
	return *o.value
}

func (o Optional[T]) OrElseGet(supplier func() T) T {
	if o.value == nil {
		return supplier()
	}
	return *o.value
}
