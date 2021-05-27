package render

import "sync"

// rwLock represents an interface for sync.RWMutex
type rwLock interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

var (
	_ rwLock = &sync.RWMutex{}
	_ rwLock = emptyLock{}
)

// emptyLock is a dummy RWLock implementation
type emptyLock struct{}

func (emptyLock) Lock()    {}
func (emptyLock) Unlock()  {}
func (emptyLock) RLock()   {}
func (emptyLock) RUnlock() {}
