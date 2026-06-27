// Package concurrent @Author larry
// @Date 2025/6/18 11:21
// @Desc

package concurrent

import (
	"sync"
)

type CSlice[T any] struct {
	mu    sync.Mutex
	slice []T
}

// NewCSlice creates a new CSlice instance.
func NewCSlice[T any]() *CSlice[T] {
	return &CSlice[T]{
		slice: make([]T, 0),
	}
}

func (s *CSlice[T]) Append(val T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slice = append(s.slice, val)
}

func (s *CSlice[T]) Get(index int) (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.slice) {
		var zero T
		return zero, false
	}
	return s.slice[index], true
}

func (s *CSlice[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.slice)
}

func (s *CSlice[T]) Slice() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]T(nil), s.slice...) // Return a copy of the slice
}

func (s *CSlice[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slice = make([]T, 0) // Reset the slice to an empty state
}
