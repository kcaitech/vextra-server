/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package my_map

import (
	"encoding/json"
	"fmt"
	"sync"
)

type SyncMap[K comparable, V any] struct {
	lock sync.RWMutex
	m    map[K]V
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		m: make(map[K]V),
	}
}

func (sm *SyncMap[K, V]) Get(key K) (V, bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	v, ok := sm.m[key]
	return v, ok
}

func (sm *SyncMap[K, V]) Set(key K, value V) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.m[key] = value
}

func (sm *SyncMap[K, V]) Delete(key K) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	delete(sm.m, key)
}

func (sm *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	for key, value := range sm.m {
		if !f(key, value) {
			break
		}
	}
}

func (sm *SyncMap[K, V]) Len() int {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return len(sm.m)
}

func (sm *SyncMap[K, V]) Clear() {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.m = make(map[K]V)
}

func (sm *SyncMap[K, V]) Keys() []K {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	keys := make([]K, 0, len(sm.m))
	for key := range sm.m {
		keys = append(keys, key)
	}
	return keys
}

func (sm *SyncMap[K, V]) Values() []V {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	values := make([]V, 0, len(sm.m))
	for _, value := range sm.m {
		values = append(values, value)
	}
	return values
}

func (sm *SyncMap[K, V]) Items() map[K]V {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	items := make(map[K]V, len(sm.m))
	for key, value := range sm.m {
		items[key] = value
	}
	return items
}

func (sm *SyncMap[K, V]) Contains(key K) bool {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	_, ok := sm.m[key]
	return ok
}

func (sm *SyncMap[K, V]) Clone() *SyncMap[K, V] {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	newMap := NewSyncMap[K, V]()
	for key, value := range sm.m {
		newMap.Set(key, value)
	}
	return newMap
}

func (sm *SyncMap[K, V]) Merge(other *SyncMap[K, V]) {
	other.lock.RLock()
	defer other.lock.RUnlock()
	for key, value := range other.m {
		sm.Set(key, value)
	}
}

func (sm *SyncMap[K, V]) Equal(other *SyncMap[K, V], isEqual func(V, V) bool) bool {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	other.lock.RLock()
	defer other.lock.RUnlock()
	if len(sm.m) != len(other.m) {
		return false
	}
	for key, value := range sm.m {
		otherValue, ok := other.m[key]
		if !ok || !isEqual(value, otherValue) {
			return false
		}
	}
	return true
}

func (sm *SyncMap[K, V]) Any(f func(key K, value V) bool) bool {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	for key, value := range sm.m {
		if f(key, value) {
			return true
		}
	}
	return false
}

func (sm *SyncMap[K, V]) All(f func(key K, value V) bool) bool {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	for key, value := range sm.m {
		if !f(key, value) {
			return false
		}
	}
	return true
}

func (sm *SyncMap[K, V]) Filter(f func(key K, value V) bool) *SyncMap[K, V] {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	newMap := NewSyncMap[K, V]()
	for key, value := range sm.m {
		if f(key, value) {
			newMap.Set(key, value)
		}
	}
	return newMap
}

func (sm *SyncMap[K, V]) Map(f func(key K, value V) (K, V)) *SyncMap[K, V] {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	newMap := NewSyncMap[K, V]()
	for key, value := range sm.m {
		newKey, newValue := f(key, value)
		newMap.Set(newKey, newValue)
	}
	return newMap
}

func (sm *SyncMap[K, V]) Reduce(f func(key K, value V, accum V) V, accum V) V {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	for key, value := range sm.m {
		accum = f(key, value, accum)
	}
	return accum
}

func (sm *SyncMap[K, V]) String() string {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return fmt.Sprintf("%v", sm.m)
}

func (sm *SyncMap[K, V]) MarshalJSON() ([]byte, error) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return json.Marshal(sm.m)
}

func (sm *SyncMap[K, V]) UnmarshalJSON(b []byte) error {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	return json.Unmarshal(b, &sm.m)
}
