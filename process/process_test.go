package process

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/qwertyqq2/entity/entity"
	mes "github.com/qwertyqq2/entity/message"
)

const (
	procLimit     = 10
	durationWrite = 5 * time.Second
)

var (
	keys  = []string{"k1", "k2", "k3", "k4"}
	datas = []string{"Between a Rock and a Hard Place",
		"Mouth-watering",
		"Put a Sock In It",
		"Money Doesn't Grow On Trees"}

	msgs = make([]msg, len(keys))
)

type msg struct {
	op  mes.Op
	ent Entry
}

func newMsg(op mes.Op, key, data string) msg {
	return msg{
		op: op,
		ent: Entry{
			key:  key,
			data: mes.NewInside(key, data),
		},
	}
}

func constructor() {
	for i, k := range keys {
		msgs[i] = newMsg(mes.Add, k, datas[i])
	}
}

func init() {
	constructor()
}

func TestRecallAdd(t *testing.T) {
	rec := newRecall()
	for _, m := range msgs {
		rec.Add(m.op, m.ent.data)
	}

	fmt.Println(rec.String())
}

func TestRecallDelete(t *testing.T) {
	rec := newRecall()
	for _, m := range msgs {
		rec.Add(m.op, m.ent.data)
	}

	keyForDelete := keys[0]
	rec.Remove(keyForDelete)

	fmt.Println(rec.String())
}

func TestMarkSent(t *testing.T) {
	rec := newRecall()
	for _, m := range msgs {
		rec.Add(m.op, m.ent.data)
	}

	keyForMarkSent := keys[1]
	if !rec.MarkSent(keyForMarkSent) {
		t.Fatal("cant mark to sent")
	}

	fmt.Println(rec.String())
}

func TestSentAt(t *testing.T) {
	rec := newRecall()
	for _, m := range msgs {
		rec.Add(m.op, m.ent.data)
	}

	keyForMarkSent := keys[1]
	if !rec.MarkSent(keyForMarkSent) {
		t.Fatal("cant mark to sent")
	}

	rec.SentAt(keyForMarkSent, time.Now())

	fmt.Println(rec.String())
}

func writeToProcess(p *impl, dur time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()
	num := 0

	for {
		select {
		case <-ctx.Done():
			return

		default:
			key := "singal " + fmt.Sprintf("num %d", num)
			p.Add(key, "newdata")
			num++
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// uploading data to a process
func unloading(proc *impl) {
	after := time.After(durationWrite)
	go proc.queueIncomig()
	go writeToProcess(proc, durationWrite)
	<-after
}

func TestQueue(t *testing.T) {
	proc := New(1)

	unloading(proc)

	proc.Shutdown()

	fmt.Println(proc.recall.String())

}

func TestRegestration(t *testing.T) {
	proc := New(1)
	ent := entity.New(procLimit)
	if err := proc.Registration(ent); err != nil {
		t.Fatal(err)
	}
}

func TestSendOnce(t *testing.T) {
	proc := New(1)
	go proc.queueIncomig()
	key := "snap"
	data := "is a snap"
	proc.Add(key, data)
	key = "snap2"
	data = "is a snap2"
	proc.Add(key, data)

	ent := entity.New(procLimit)
	if err := proc.Registration(ent); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)

	if !proc.hasPendingWork() {
		t.Fatal("its not true")
	}

	proc.sendIfReady()

	fmt.Println(ent.String())

	fmt.Println(proc.recall.String())
}

func TestSend(t *testing.T) {
	proc := New(1)
	unloading(proc)

	ent := entity.New(procLimit)

	if err := proc.Registration(ent); err != nil {
		t.Fatal(err)
	}

	for proc.hasPendingWork() {
		proc.sendIfReady()
	}

	fmt.Println(ent.String())

}

func TestSendTimeout(t *testing.T) {
	proc := New(1)

	unloading(proc)

	ent := entity.New(procLimit)

	if err := proc.Registration(ent); err != nil {
		t.Fatal(err)
	}
	for proc.hasPendingWork() {
		proc.sendIfReady()
	}

	len := ent.Len()

	time.Sleep(3 * time.Second)

	if proc.hasPendingWork() {
		proc.sendMessage()
	}

	if ent.Len() != len {
		t.Fatal("its not true")
	}
}

func TestRunSendMsgCutoff(t *testing.T) {
	proc := New(1)
	ent := entity.New(procLimit)

	var (
		n = 1
	)

	addFn := func() {
		for {
			key := fmt.Sprintf("key%d", n)
			proc.Add(key, " ")
			n++
			time.Sleep(50 * time.Millisecond)
		}
	}

	if err := proc.Start(ent); err != nil {
		t.Fatal(err)
	}

	go addFn()

	after := time.After(300 * time.Second)
	<-after
	fmt.Println("n", n)
	fmt.Println(ent.Len())
}

func distribution(processes ...*impl) []chan struct{} {
	var (
		n      = len(processes)
		delay  = 100 * time.Millisecond
		breaks = make([]chan struct{}, n)
	)

	handleProc := func(num int, proc *impl) {
		breaks[num] = make(chan struct{})
		for {
			select {
			case <-breaks[num]:
			default:
				continue
			}
			key, data := fmt.Sprintf("key%d", num), fmt.Sprintf("data for %d process", num)
			proc.Add(key, data)
			time.Sleep(delay)
		}
	}

	for i, proc := range processes {
		go handleProc(i, proc)
	}
	return breaks
}
