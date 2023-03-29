package entity

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	mes "github.com/qwertyqq2/entity/message"
)

const (
	delayDisconnect = 10 * time.Second
	delayResponse   = 5 * time.Second
)

type Message = mes.Message
type Inside = mes.Inside

// Key-value store
type Entity interface {
	//Closeing entity
	Shutdown()

	//Connect process to entity
	Connect(id int) (chan<- Message, error)

	//Disconnect a process from an entity
	Disconnect(id int) error

	//Getting response on request asynchronously
	Resp(id int) Message

	//Display data in entities
	String() string

	// Len of entity
	Len() int
}

type entity struct {
	ctx      context.Context
	shutdown func()

	vals map[string]Inside
	lk   sync.RWMutex

	sessions  map[int]struct{}
	inside    map[int]chan Message
	proc      int
	procLimit int
	slk       sync.RWMutex
}

func New(procLimit int) Entity {
	ctx, cancel := context.WithCancel(context.Background())
	return &entity{
		ctx:       ctx,
		shutdown:  cancel,
		vals:      make(map[string]Inside),
		sessions:  make(map[int]struct{}),
		inside:    make(map[int]chan Message),
		proc:      0,
		procLimit: procLimit,
	}
}

func (e *entity) Shutdown() {
	e.shutdown()
}

func (e *entity) Connect(id int) (chan<- Message, error) {
	e.slk.Lock()
	_, ok := e.sessions[id]
	if !ok && e.proc < e.procLimit {
		e.sessions[id] = struct{}{}
		e.proc++
		e.slk.Unlock()
		in := e.startSession(id)
		time.Sleep(100 * time.Millisecond)
		return in, nil
	}
	e.slk.Unlock()
	if ok {
		return nil, fmt.Errorf("conn already exist")
	}
	if e.proc >= e.procLimit {
		return nil, fmt.Errorf("limit processes")
	}
	return nil, nil
}

func (e *entity) Disconnect(id int) error {
	e.slk.Lock()
	defer e.slk.Unlock()

	if _, ok := e.sessions[id]; ok {
		delete(e.sessions, id)
		delete(e.inside, id)
		e.proc--
		return nil
	}
	return errors.New("disconnect err")

}

func (e *entity) isConnected(id int) bool {
	e.lk.RLock()
	defer e.lk.RUnlock()

	if _, ok := e.sessions[id]; ok {
		return true
	}
	return false
}

func (e *entity) startSession(id int) chan<- Message {
	out := make(chan Message, 100)

	howlong := clock.New().Timer(0)
	if !howlong.Stop() {
		<-howlong.C
	}

	in := make(chan Message, 100)
	e.slk.Lock()
	if _, ok := e.inside[id]; ok {
		panic("")
	}
	e.inside[id] = in
	e.slk.Unlock()

	howlong.Reset(delayDisconnect)

	go func() {
		defer e.closeConn(id)
		for {
			select {
			case m, ok := <-out:
				if !ok {
					return
				}
				if err := e.handle(m); err != nil {
					e.sendResp(mes.FailMessage(err), id)
					return
				}
				e.sendResp(mes.SuccessMessage(), id)
				howlong.Reset(delayDisconnect)

			case <-howlong.C:
				return

			case <-e.ctx.Done():
				return
			}
		}
	}()
	return out
}

func (e *entity) Len() int {
	e.lk.RLock()
	defer e.lk.RUnlock()

	return len(e.vals)
}

func (e *entity) closeConn(id int) {
	e.slk.Lock()
	defer e.slk.Unlock()

	e.proc--
	delete(e.sessions, id)
	delete(e.inside, id)
}

func (e *entity) handle(m Message) error {
	if m.IsNil() {
		return errors.New("nil mes")
	}
	switch m.Op {
	case mes.Add:
		e.handleAdd(m)
		return nil

	case mes.Delete:
		e.handleDelete(m)
		return nil

	case mes.Ping:
		return nil
	default:
		return errors.New("undefined mes")
	}
}

func (e *entity) handleAdd(m Message) {
	e.add(m.Inside)
}

func (e *entity) handleDelete(m Message) {
	e.delete(m.Inside.Key)
}

func (e *entity) sendResp(m Message, id int) {
	e.slk.RLock()

	if _, ok := e.inside[id]; !ok {
		return
	}
	e.slk.RUnlock()

	select {
	case e.inside[id] <- m:
	}
}

func (e *entity) Resp(id int) Message {
	ctx, cancel := context.WithTimeout(context.Background(), delayResponse)
	defer cancel()
	select {
	case m := <-e.inside[id]:
		return m
	case <-ctx.Done():
		return mes.Message{}
	}

}

func (e *entity) add(in Inside) {
	e.lk.Lock()
	defer e.lk.Unlock()

	e.vals[in.Key] = in
}

func (e *entity) get(key string) Inside {
	e.lk.RLock()
	defer e.lk.RUnlock()

	if val, ok := e.vals[key]; ok {
		return val
	}

	return Inside{}
}

func (e *entity) has(key string) bool {
	e.lk.RLock()
	defer e.lk.RUnlock()

	if _, ok := e.vals[key]; ok {
		return true
	}
	return false
}

func (e *entity) delete(key string) {
	e.lk.Lock()
	defer e.lk.Unlock()

	delete(e.vals, key)
}

func (e *entity) String() string {
	e.lk.RLock()
	defer e.lk.RUnlock()

	res := ""
	for key, in := range e.vals {
		val := fmt.Sprintf("key: %s, data: %s \n", key, in.Data)
		res += val
	}
	return res
}
