package process

import (
	"sync"
)

type Queue struct {
	lk    sync.RWMutex
	queue []Entry
	cond  sync.Cond
	stop  bool
	done  chan struct{}
	out   chan Entry
}

func newQueue() *Queue {
	q := &Queue{
		queue: make([]Entry, 0),
		done:  make(chan struct{}),
	}
	q.cond = sync.Cond{L: &q.lk}
	return q
}

func (q *Queue) Shutdown() {
	q.lk.Lock()
	q.stop = true
	q.lk.Unlock()
	q.cond.Broadcast()
	<-q.done
}

func (q *Queue) wait() bool {
	for !q.stop && len(q.queue) == 0 {
		q.cond.Wait()
	}
	return !q.stop
}

func (q *Queue) pull() Entry {
	if q.len() == 0 {
		panic("empty queue")
	}
	e := q.queue[0]
	q.queue = q.queue[1:]
	return e
}

func (q *Queue) Push(e Entry) {
	q.lk.Lock()
	defer q.lk.Unlock()
	q.queue = append(q.queue, e)
	q.cond.Broadcast()
}

func (q *Queue) len() int {
	return len(q.queue)
}
