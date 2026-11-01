package hookfacet

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParentRead(t *testing.T) {
	root := NewSessionState()
	Set(root, 42)
	child := NewChildState(root)
	v, ok := Get[int](child)
	assert.True(t, ok)
	assert.Equal(t, 42, v)
}

func TestReadPriority(t *testing.T) {
	root := NewSessionState()
	Set(root, 1)
	child := NewChildState(root)
	Set(child, 2)
	grandchild := NewChildState(child)
	Set(grandchild, 3)

	v, ok := Get[int](grandchild)
	assert.True(t, ok)
	assert.Equal(t, 3, v)

	v2, ok2 := Get[int](child)
	assert.True(t, ok2)
	assert.Equal(t, 2, v2)

	v3, ok3 := Get[int](root)
	assert.True(t, ok3)
	assert.Equal(t, 1, v3)
}

type MyType struct {
	Value string
}

func TestTypeAsKey(t *testing.T) {
	root := NewSessionState()
	Set(root, 42)
	Set(root, "hi")
	Set(root, MyType{Value: "hello"})

	v, ok := Get[int](root)
	assert.True(t, ok)
	assert.Equal(t, 42, v)

	v2, ok2 := Get[string](root)
	assert.True(t, ok2)
	assert.Equal(t, "hi", v2)

	v3, ok3 := Get[bool](root)
	assert.False(t, ok3)
	assert.False(t, v3)

	v4, ok4 := Get[MyType](root)
	assert.True(t, ok4)
	assert.Equal(t, MyType{Value: "hello"}, v4)
}

func TestNilSafe(t *testing.T) {
	v, ok := Get[int](nil)
	assert.False(t, ok)
	assert.Equal(t, 0, v)

	Set[int](nil, 1) // no panic
}

func TestRaceSafe(t *testing.T) {
	root := NewSessionState()
	Set(root, 42)
	child := NewChildState(root)
	Set(child, "hi")

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			for range 10 {
				_, ok := Get[int](child)
				assert.True(t, ok)
			}
		})
	}
	for range 10 {
		wg.Go(func() {
			for range 10 {
				next := rand.Int()
				Set[int](root, next)
			}
		})
	}
	for range 10 {
		wg.Go(func() {
			for range 10 {
				_, ok := Get[string](child)
				assert.True(t, ok)
			}
		})
	}
	for range 10 {
		wg.Go(func() {
			for range 10 {
				next := fmt.Sprintf("hi%d", rand.Int())
				Set[string](child, next)
			}
		})
	}

	wg.Wait()
}
