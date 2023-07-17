package set

type Set[T comparable] struct {
	set map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	set := &Set[T]{
		set: make(map[T]struct{}),
	}
	set.AddItems(items...)
	return set
}

func (s *Set[T]) Add(item T) {
	s.set[item] = struct{}{}
}

func (s *Set[T]) AddItems(items ...T) {
	for _, item := range items {
		s.set[item] = struct{}{}
	}
}

func (s *Set[T]) Remove(item T) {
	delete(s.set, item)
}

func (s *Set[T]) Contains(item T) bool {
	_, ok := s.set[item]
	return ok
}

func (s *Set[T]) Size() int {
	return len(s.set)
}

func (s *Set[T]) Clear() {
	s.set = make(map[T]struct{})
}

func (s *Set[T]) Items() []T {
	items := make([]T, 0, len(s.set))
	for key := range s.set {
		items = append(items, key)
	}
	return items
}

func (s *Set[T]) Clone() *Set[T] {
	newSet := NewSet[T]()
	for key := range s.set {
		newSet.set[key] = struct{}{}
	}
	return newSet
}

func (s *Set[T]) Union(other *Set[T]) *Set[T] { // 并集
	newSet := s.Clone()
	for key := range other.set {
		newSet.set[key] = struct{}{}
	}
	return newSet
}

func (s *Set[T]) Intersect(other *Set[T]) *Set[T] { // 交集
	newSet := NewSet[T]()
	for key := range s.set {
		if _, ok := other.set[key]; ok {
			newSet.set[key] = struct{}{}
		}
	}
	return newSet
}

func (s *Set[T]) Difference(other *Set[T]) *Set[T] { // 差集
	newSet := NewSet[T]()
	for key := range s.set {
		if _, ok := other.set[key]; !ok {
			newSet.Add(key)
		}
	}
	return newSet
}

func (s *Set[T]) IsSubset(other *Set[T]) bool { // 判断是否为other的子集
	for key := range s.set {
		if _, ok := other.set[key]; !ok {
			return false
		}
	}
	return true
}

func (s *Set[T]) IsSuperset(other *Set[T]) bool { // 判断是否为other的超集
	return other.IsSubset(s)
}
