package entity

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	mes "github.com/qwertyqq2/entity/message"
)

const (
	delayRequest = 3 * time.Second
)

type proc struct {
	id   int
	conn chan<- Message
}

func newProc(id int) *proc {
	return &proc{
		id: id,
	}
}

func (p *proc) Connect(e Entity) error {
	conn, err := e.Connect(p.id)
	if err != nil {
		return err
	}
	p.conn = conn
	return nil
}

func (p *proc) Disconnect(e Entity) error {
	return e.Disconnect(p.id)
}

func (p *proc) Send(m Message, e Entity) error {
	ctx, cancel := context.WithTimeout(context.Background(), delayRequest)
	defer cancel()
	select {
	case p.conn <- m:
		resp := e.Resp(p.id)
		if resp.IsNil() {
			return errors.New("nil response")
		}
		return nil
	case <-ctx.Done():
		return errors.New("timeout")
	}
}

const (
	procLimit = 2
)

var (
	ent   Entity
	proc1 *proc
	proc2 *proc
	proc3 *proc
	proc4 *proc
)

func constructor() {
	ent = New(procLimit)
	proc1 = newProc(1)
	proc2 = newProc(2)
	proc3 = newProc(3)
	proc4 = newProc(4)

}

func TestConnect(t *testing.T) {
	constructor()

	if err := proc1.Connect(ent); err != nil {
		t.Fatal(err)
	}
	if err := proc2.Connect(ent); err != nil {
		t.Fatal(err)
	}

	if err := proc3.Connect(ent); err == nil {
		t.Fatal("connect over proc limit")
	}

}

func TestDisconnect(t *testing.T) {
	constructor()

	if err := proc1.Connect(ent); err != nil {
		t.Fatal(err)
	}
	if err := proc2.Connect(ent); err != nil {
		t.Fatal(err)
	}

	if err := proc2.Disconnect(ent); err != nil {
		t.Fatal(err)
	}

	if err := proc3.Connect(ent); err != nil {
		t.Fatal(err)
	}

}

func TestSendAdd(t *testing.T) {
	constructor()

	if err := proc1.Connect(ent); err != nil {
		t.Fatal(err)
	}

	msg1 := mes.AddMessage("n1", "is n1")
	if err := proc1.Send(msg1, ent); err != nil {
		t.Fatal(err)
	}

	msg2 := mes.AddMessage("n2", "is n2")
	if err := proc1.Send(msg2, ent); err != nil {
		t.Fatal(err)
	}
	msg3 := mes.AddMessage("n3", "is n3")
	if err := proc1.Send(msg3, ent); err != nil {
		t.Fatal(err)
	}

	fmt.Println(ent.String())
}

func TestSendDelete(t *testing.T) {
	constructor()

	if err := proc1.Connect(ent); err != nil {
		t.Fatal(err)
	}

	if err := proc2.Connect(ent); err != nil {
		t.Fatal(err)
	}

	msg1 := mes.AddMessage("n1", "is n1")
	if err := proc1.Send(msg1, ent); err != nil {
		t.Fatal(err)
	}

	msg2 := mes.AddMessage("n2", "is n2")
	if err := proc2.Send(msg2, ent); err != nil {
		t.Fatal(err)
	}

	msg3 := mes.AddMessage("n3", "is n3")
	if err := proc1.Send(msg3, ent); err != nil {
		t.Fatal(err)
	}

	msgdel1 := mes.DeleteMessage("n2")
	msgdel2 := mes.DeleteMessage("n3")

	if err := proc1.Send(msgdel1, ent); err != nil {
		t.Fatal(err)
	}
	if err := proc1.Send(msgdel2, ent); err != nil {
		t.Fatal(err)
	}
	fmt.Println(ent.String())

}

func TestTimeout(t *testing.T) {
	constructor()

	if err := proc1.Connect(ent); err != nil {
		t.Fatal(err)
	}

	time.Sleep(delayDisconnect)
	msg1 := mes.AddMessage("n1", "is n1")
	if err := proc1.Send(msg1, ent); err == nil {
		t.Fatal("timeout not working")
	}

}

func TestEntry(t *testing.T) {
	es := []int{1, 2, 3, 4, 2}
	ess := es[0:len(es):len(es)]

	fmt.Println(ess, len(ess), cap(ess))

}
