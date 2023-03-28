package process

import (
	"context"
	"sync"

	mes "github.com/qwertyqq2/entity/message"
)

type Queue struct {
	ctx      context.Context
	shutdown func()
	ch       chan mes.Inside
	q        []Entry
	mu       sync.RWMutex
}

func newQueue(ctx context.Context, ch chan mes.Inside) *Queue {
	qctx, cancel := context.WithCancel(ctx)
	return &Queue{
		ctx:      qctx,
		shutdown: cancel,
		ch:       ch,
		q:        make([]Entry, 0),
	}
}

func (q *Queue) Shutdown() {
	q.shutdown()
}

func (q *Queue) Push(e Entry) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.q = append(q.q, e)

}

func (q *Queue) Can() (bool, error) {
	select {
	case <-q.ctx.Done():
		return false, q.ctx.Err()
	default:
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.Len() == 0 {
		return false, nil
	}

	return true, nil
}

func (q *Queue) Pull() (Entry, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	d := q.q[0]
	q.q = q.q[1:]
	return d, nil
}

func (q *Queue) Len() int {
	return len(q.q)
}
