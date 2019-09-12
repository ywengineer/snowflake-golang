package snowflake

import (
	"container/list"
	"errors"
	"sync"
)

type BlockingQueue struct {
	list     *list.List
	mux      sync.Mutex
	capacity int
}

func NewQueue(capacity int) *BlockingQueue {
	return &BlockingQueue{
		list:     list.New(),
		mux:      sync.Mutex{},
		capacity: capacity,
	}
}

func (q *BlockingQueue) Capacity() int {
	return q.capacity
}

func (q *BlockingQueue) RemainCapacity() int {
	q.mux.Lock()
	defer q.mux.Unlock()
	return q.capacity - q.list.Len()
}

func (q *BlockingQueue) Len() int {
	q.mux.Lock()
	defer q.mux.Unlock()
	return q.list.Len()
}

func (q *BlockingQueue) Push(v int64) {
	q.mux.Lock()
	defer q.mux.Unlock()
	q.list.PushBack(v)
}

func (q *BlockingQueue) PushAll(slice []int64) {
	q.mux.Lock()
	defer q.mux.Unlock()
	for _, v := range slice {
		q.list.PushBack(v)
	}
}

func (q *BlockingQueue) Take() (int64, error) {
	q.mux.Lock()
	defer q.mux.Unlock()
	if e := q.list.Front(); e == nil {
		return 0, errors.New("no such element")
	} else {
		return q.list.Remove(e).(int64), nil
	}
}
