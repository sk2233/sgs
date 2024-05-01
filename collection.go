/*
@author: sk
@date: 2024/5/1
*/
package main

type Stack[T any] struct {
	data []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Push(data T) {
	s.data = append(s.data, data)
}

func (s *Stack[T]) Pop() T {
	l := len(s.data)
	data := s.data[l-1]
	s.data = s.data[:l-1]
	return data
}

func (s *Stack[T]) Peek() T {
	l := len(s.data)
	return s.data[l-1]
}

func (s *Stack[T]) Len() int {
	return len(s.data)
}
