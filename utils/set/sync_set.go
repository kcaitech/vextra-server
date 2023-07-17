package set

import "sync"

type SyncSet struct {
	lock sync.RWMutex
	set  map[any]struct{}
}

func NewSyncSet() *SyncSet {
	return &SyncSet{
		set: make(map[any]struct{}),
	}
}

func (s *SyncSet) Add(item any) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.set[item] = struct{}{}
}

func (s *SyncSet) AddItems(items ...any) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, item := range items {
		s.set[item] = struct{}{}
	}
}

func (s *SyncSet) Remove(item any) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.set, item)
}

func (s *SyncSet) Contains(item any) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.set[item]
	return ok
}

func (s *SyncSet) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.set)
}

func (s *SyncSet) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.set = make(map[any]struct{})
}

func (s *SyncSet) Items() []any {
	s.lock.RLock()
	defer s.lock.RUnlock()
	items := make([]any, 0, len(s.set))
	for key := range s.set {
		items = append(items, key)
	}
	return items
}

func (s *SyncSet) Clone() *SyncSet {
	newSet := NewSyncSet()
	s.lock.RLock()
	defer s.lock.RUnlock()
	for key := range s.set {
		newSet.set[key] = struct{}{}
	}
	return newSet
}

func (s *SyncSet) Union(other *SyncSet) *SyncSet { // 并集
	newSet := s.Clone()
	other.lock.RLock()
	defer other.lock.RUnlock()
	for key := range other.set {
		newSet.set[key] = struct{}{}
	}
	return newSet
}

func (s *SyncSet) Intersect(other *SyncSet) *SyncSet { // 交集
	newSet := NewSyncSet()
	s.lock.RLock()
	defer s.lock.RUnlock()
	other.lock.RLock()
	defer other.lock.RUnlock()
	for key := range s.set {
		if _, ok := other.set[key]; ok {
			newSet.set[key] = struct{}{}
		}
	}
	return newSet
}

func (s *SyncSet) Difference(other *SyncSet) *SyncSet { // 差集
	newSet := NewSyncSet()
	s.lock.RLock()
	defer s.lock.RUnlock()
	other.lock.RLock()
	defer other.lock.RUnlock()
	for key := range s.set {
		if _, ok := other.set[key]; !ok {
			newSet.Add(key)
		}
	}
	return newSet
}

func (s *SyncSet) IsSubset(other *SyncSet) bool { // 判断是否为other的子集
	s.lock.RLock()
	defer s.lock.RUnlock()
	other.lock.RLock()
	defer other.lock.RUnlock()
	for key := range s.set {
		if _, ok := other.set[key]; !ok {
			return false
		}
	}
	return true
}

func (s *SyncSet) IsSuperset(other *SyncSet) bool { // 判断是否为other的超集
	return other.IsSubset(s)
}
