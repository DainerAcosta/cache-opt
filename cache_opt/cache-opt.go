package cache_opt

import (
	"sync"
)

type Memory struct {
	f          Function
	InProgress map[InputKey]bool
	IsPending  map[InputKey][]chan OutputResponse
	cache      map[InputKey]FunctionResult
	lock       sync.RWMutex
}

type Function func(key interface{}) (OutputResponse, error)

type InputKey interface{}

type OutputResponse interface{}

type FunctionResult struct {
	value interface{}
	err   error
}

func NewCache(f Function) *Memory {
	return &Memory{
		f:          f,
		InProgress: make(map[InputKey]bool),
		IsPending:  make(map[InputKey][]chan OutputResponse),
		cache:      make(map[InputKey]FunctionResult),
	}
}

func (m *Memory) Get(key InputKey) (interface{}, error) {
	m.lock.Lock()
	result, exists := m.cache[key]
	m.lock.Unlock()
	if !exists {
		result.value, result.err = m.Work(key)
		m.lock.Lock()
		m.cache[key] = result
		m.lock.Unlock()
	}
	return result.value, result.err
}

func (m *Memory) Work(job InputKey) (OutputResponse, error) {
	m.lock.RLock()
	exists := m.InProgress[job]
	if exists {
		m.lock.RUnlock()
		response := make(chan OutputResponse)
		defer close(response)

		m.lock.Lock()
		m.IsPending[job] = append(m.IsPending[job], response)
		m.lock.Unlock()
		resp := <-response
		return resp, nil
	}
	m.lock.RUnlock()

	m.lock.Lock()
	m.InProgress[job] = true
	m.lock.Unlock()

	result, _ := m.f(job)

	m.lock.RLock()
	pendingWorkers, exists := m.IsPending[job]
	m.lock.RUnlock()

	if exists {
		for _, pendingWorker := range pendingWorkers {
			pendingWorker <- result
		}
	}
	m.lock.Lock()
	m.InProgress[job] = false
	m.IsPending[job] = make([]chan OutputResponse, 0)
	m.lock.Unlock()
	return result, nil
}
