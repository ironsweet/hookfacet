package hookfacet

import (
	"reflect"
	"sync"
)

type State struct {
	parent *State
	mu     sync.RWMutex
	values map[reflect.Type]any
}

func NewSessionState() *State {
	return &State{
		values: make(map[reflect.Type]any),
	}
}

func NewChildState(parent *State) *State {
	s := NewSessionState()
	s.parent = parent
	return s
}

func Get[T any](st *State) (T, bool) {
	key := reflect.TypeFor[T]()
	for current := st; current != nil; current = current.parent {
		current.mu.RLock()
		if value, ok := current.values[key]; ok {
			current.mu.RUnlock()
			return value.(T), true
		}
		current.mu.RUnlock()
	}
	var zero T
	return zero, false
}

func Set[T any](st *State, value T) {
	if st == nil {
		return
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	key := reflect.TypeFor[T]()
	st.values[key] = value
}
