/*
@author: sk
@date: 2024/5/1
*/
package main

//=================Stack====================

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

//================Set=================

type Set[T comparable] struct {
	data map[T]struct{}
}

func NewSet[T comparable](data ...T) *Set[T] {
	res := &Set[T]{data: make(map[T]struct{}, len(data))}
	for _, item := range data {
		res.Add(item)
	}
	return res
}

func (s *Set[T]) Add(data T) {
	s.data[data] = struct{}{}
}

func (s *Set[T]) Contain(data T) bool {
	_, ok := s.data[data]
	return ok
}
