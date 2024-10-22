package pkg

import "sync"

type Queue[T any] struct {
	queue []T
	mutex sync.RWMutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		queue: []T{},
		mutex: sync.RWMutex{},
	}
}

func (q *Queue[T]) Enqueue(value T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.queue = append(q.queue, value)
}

func (q *Queue[T]) Dequeue() (T, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	var value T
	if len(q.queue) == 0 {
		return value, false
	}

	value = q.queue[0]
	q.queue = q.queue[1:]
	return value, true
}

func (q *Queue[T]) Length() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.queue)
}
