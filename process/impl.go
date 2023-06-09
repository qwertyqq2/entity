package process

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	opts "github.com/qwertyqq2/entity"
	"github.com/qwertyqq2/entity/entity"
	mes "github.com/qwertyqq2/entity/message"
)

var (
	ErrShutdownProcess = errors.New("process is closed")
	ErrTimeoutSend     = errors.New("timeout send")
	ErrNilMsg          = errors.New("nil mes err")
)

type recall struct {
	want      *Wantlist
	sent      *Wantlist
	undefined *Wantlist
	sentAt    map[string]time.Time
}

func (r *recall) Add(op mes.Op, data mes.Inside) {
	r.want.Add(data.Key, op, data)
}

func (r *recall) Remove(key string) {
	r.want.Remove(key)
	r.sent.Remove(key)
	delete(r.sentAt, key)
}

func (r *recall) MarkSent(key string) bool {
	e, ok := r.want.Contains(key)
	if !ok {
		return false
	}

	r.sent.Add(key, e.op, e.data)
	return true
}

func (r *recall) MarkUndefined(op mes.Op, data mes.Inside) {
	r.undefined.Add(data.Key, op, data)
}

func (r *recall) SentAt(key string, at time.Time) {
	if _, ok := r.sent.Contains(key); ok {
		if _, ok := r.sentAt[key]; !ok {
			r.sentAt[key] = at
		}
	}
}

func (r *recall) ClearSentAt(key string) {
	delete(r.sentAt, key)
}

func (r *recall) String() string {
	wnt := r.want.String()
	snt := r.sent.String()
	sntAt := ""
	for k, time := range r.sentAt {
		sntAt += fmt.Sprintf("key: %s, time: %s\n", k, time.String())
	}

	return fmt.Sprintf("Want: \n%s \nSent: \n%s \nSentAt: \n%s", wnt, snt, sntAt)
}

func newRecall() recall {
	return recall{
		want:      NewWantlist(),
		sent:      NewWantlist(),
		undefined: NewWantlist(),
		sentAt:    make(map[string]time.Time),
	}
}

type impl struct {
	id int

	queue *Queue

	ctx      context.Context
	shutdown func()

	outgoingWork chan time.Time
	disconnected chan struct{}

	ek       sync.RWMutex
	entityCh chan<- mes.Message
	ent      entity.Entity

	clock clock.Clock

	wk     sync.RWMutex
	recall recall

	lk                   sync.RWMutex
	maxMsgSize           int
	sendMsgCutoff        int
	maxWaitingConnection time.Duration
	sendMessageMaxDelay  time.Duration
	reconnectInterval    time.Duration
	resendInterval       time.Duration
	waitRespInterval     time.Duration
	waitRespTimer        *clock.Timer
	resendTimer          *clock.Timer
}

func newProc(id int) *impl {
	ctx, cancel := context.WithCancel(context.Background())

	clock := clock.New()

	resendTimer := clock.Timer(opts.ResentInterval)

	return &impl{
		id:               id,
		ctx:              ctx,
		shutdown:         cancel,
		outgoingWork:     make(chan time.Time),
		clock:            clock,
		recall:           newRecall(),
		resendInterval:   opts.ResentInterval,
		resendTimer:      resendTimer,
		queue:            newQueue(),
		waitRespInterval: opts.WaitSendInterval,
		waitRespTimer:    clock.Timer(opts.WaitSendInterval),
		disconnected:     make(chan struct{}),
	}
}

func (p *impl) Shutdown() {
	p.queue.Shutdown()
	p.shutdown()
}

func (p *impl) Registration(ent entity.Entity) error {
	p.ek.Lock()
	defer p.ek.Unlock()
	ch, err := ent.Connect(p.id)
	if err != nil {
		return err
	}
	p.entityCh = ch
	p.ent = ent
	return nil
}

func (p *impl) Start(ent entity.Entity) error {
	if err := p.Registration(ent); err != nil {
		return err
	}
	go p.queueIncomig()
	go p.run()
	return nil
}

func (p *impl) Add(key, data string) {
	p.queue.Push(Entry{op: mes.Add, data: mes.NewInside(key, data)})
}

func (p *impl) Delete(key string) {
	p.queue.Push(Entry{op: mes.Delete, data: mes.NewInside(key, "")})
}

func (p *impl) run() {
	defer p.queue.Shutdown()

	scheduleWork := p.clock.Timer(0)
	if !scheduleWork.Stop() {
		<-scheduleWork.C
	}

	var workScheduled time.Time
	for p.ctx.Err() == nil {
		select {
		case <-p.resendTimer.C:
			p.resendTimer.Reset(opts.ResentInterval)
			p.sendIfReady()

		case when := <-p.outgoingWork:
			if workScheduled.IsZero() {
				workScheduled = when
			} else if !scheduleWork.Stop() {
				<-scheduleWork.C
			}
			if p.pendingCount() > opts.SendMsgCutoff || p.clock.Since(workScheduled) >= opts.SendMessageMaxDelay {
				p.sendIfReady()
				workScheduled = time.Time{}
			} else {
				scheduleWork.Reset(opts.SendMessageMaxDelay)
			}

		case <-scheduleWork.C:
			workScheduled = time.Time{}
			p.sendIfReady()

		case <-p.ctx.Done():
			return

		}
	}

}

func (p *impl) queueIncomig() {
	p.queue.lk.Lock()
	defer p.queue.lk.Unlock()
	defer close(p.queue.done)

	for p.queue.wait() {
		ins := p.queue.pull()
		p.wk.Lock()
		if e, ok := p.recall.want.Contains(ins.key); ok {
			if e.Empty() {
				p.recall.Remove(ins.key)
			}
			continue
		}

		if e, ok := p.recall.sent.Contains(ins.key); ok && !e.Empty() {
			if e.Empty() {
				p.recall.Remove(ins.key)
			}
			continue
		}
		p.wk.Unlock()

		p.recall.Add(ins.op, ins.data)

		select {
		case <-p.ctx.Done():
			return
		default:
			p.signalWorkReady()
		}
	}

}

func (p *impl) SetResentInterval(delay time.Duration) {
	p.lk.Lock()
	defer p.lk.Unlock()

	p.resendInterval = delay
	if p.resendTimer == nil {
		return
	}
	p.resendTimer.Reset(delay)
}

func (p *impl) SetWaitRespInterval(delay time.Duration) {
	p.lk.Lock()
	defer p.lk.Unlock()

	p.waitRespInterval = delay
}

func (p *impl) resentWithTransfer() {
	p.lk.RLock()
	p.resendTimer.Reset(p.resendInterval)
	p.lk.RUnlock()

	if p.transferResent() {
		if err := p.sendMessage(); err != nil {
			p.handleError(err)
		}
	}
}

func (p *impl) sendIfReady() {
	if p.hasPendingWork() {
		if err := p.sendMessage(); err != nil {
			p.handleError(err)
		}
	}
}

func (p *impl) handleError(err error) {
	log.Println("handle error")
	select {
	case <-p.ctx.Done():
		return
	default:
	}
	switch err {
	case ErrTimeoutSend:
		after := time.After(p.maxWaitingConnection)
		reconnTimer := p.clock.Timer(0)

		if !reconnTimer.Stop() {
			<-reconnTimer.C
		}
		if err := p.Registration(p.ent); err != nil {
			reconnTimer.Reset(p.reconnectInterval)
		} else {
			goto Connection
		}
		for {
			select {
			case <-p.ctx.Done():
				return

			case <-after:
				p.Shutdown()

			case <-reconnTimer.C:
				if err := p.Registration(p.ent); err != nil {
					reconnTimer.Reset(p.reconnectInterval)
					continue
				}
				goto Connection
			}
		}
	Connection:
		p.sendIfReady()

	default:
	}
}

func (p *impl) sendMessage() error {
	p.wk.Lock()
	entires := p.recall.want.Entries()
	p.wk.Unlock()

	var (
		msgSize     = 0
		sentEntries = 0
	)

	msgs := make([]mes.Message, 0)

	for _, e := range entires {
		p.wk.Lock()
		if !p.recall.MarkSent(e.key) {
			p.recall.want.Remove(e.key)
			continue
		}
		p.wk.Unlock()

		msg := mes.Message{
			Inside: e.data,
			Op:     e.op,
		}
		msgSize += msg.Size()
		sentEntries++
		msgs = append(msgs, msg)

		if msgSize > p.maxMsgSize {
			break
		}

		p.wk.Lock()
		p.recall.want.Remove(msg.Inside.Key)
		p.recall.sent.Add(msg.Inside.Key, msg.Op, msg.Inside)
		p.wk.Unlock()
	}

	if len(msgs) == 0 {
		return ErrNilMsg
	}

	p.wk.Lock()
	defer p.wk.Unlock()
	log.Println("send")
	for _, m := range msgs {
		select {
		case <-p.ctx.Done():
			return ErrShutdownProcess

		case p.entityCh <- m:
			done := make(chan struct{})
			p.recall.SentAt(m.Inside.Key, p.clock.Now())

			if !p.waitRespTimer.Stop() {
				<-p.waitRespTimer.C
			}

			go func() {
				doneFn := func() {
					select {
					case done <- struct{}{}:
					}
				}
				resp := p.ent.Resp(p.id)
				p.recall.ClearSentAt(m.Inside.Key)
				switch resp.Op {
				case mes.Success:
					p.recall.Remove(m.Inside.Key)
					doneFn()
				case mes.Fail:
					p.recall.MarkUndefined(m.Op, m.Inside)
					p.recall.sent.Remove(m.Inside.Key)
					doneFn()
				default:
					p.recall.want.Add(m.Inside.Key, m.Op, m.Inside)
					p.recall.sent.Remove(m.Inside.Key)
				}
			}()

			p.waitRespTimer.Reset(p.waitRespInterval)

			select {
			case <-done:
				p.waitRespTimer.Reset(p.waitRespInterval)

			// entity not work with you
			case <-p.waitRespTimer.C:
				p.recall.want.Add(m.Inside.Key, m.Op, m.Inside)
				p.recall.sent.Remove(m.Inside.Key)
				return ErrTimeoutSend
			}
		}
	}
	p.logSendingMessage(msgs)
	return nil
}

func (p *impl) hasPendingWork() bool {
	return p.recall.want.Len() > 0
}

func (p *impl) transferResent() bool {
	p.wk.Lock()
	defer p.wk.Unlock()

	if p.recall.sent.Len() == 0 {
		return false
	}
	p.recall.want.Absorf(p.recall.sent)
	return true
}

func (p *impl) pendingCount() int {
	return p.recall.want.Len()
}

func (p *impl) signalWorkReady() {
	select {
	case p.outgoingWork <- p.clock.Now():
	default:
	}
}

func (p *impl) ID() int {
	return p.id
}

func (p *impl) logSendingMessage(msgs []entity.Message) {
	res := fmt.Sprintf("sending: %d, want: %d, undefined: %d\n", len(msgs), p.recall.want.Len(), p.recall.undefined.Len())
	log.Println(res)
}
