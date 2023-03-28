package process

import (
	"context"
	"fmt"
	"testing"

	mes "github.com/qwertyqq2/entity/message"
)

func TestQueuePush(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan mes.Inside)

	queue := newQueue(ctx, ch)

	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("first", "data first"),
	})

	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("second", "data second"),
	})

	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("third", "data third"),
	})

	q1, err := queue.Pull()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(q1.String())

	q2, err := queue.Pull()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(q2.String())

	q3, err := queue.Pull()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(q3.String())

	if queue.Len() != 0 {
		t.Fatal("not zero length")
	}
}

func TestQueuePull(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan mes.Inside)

	queue := newQueue(ctx, ch)

	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("first", "data first"),
	})

	q1, err := queue.Pull()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(q1.String())

	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("second", "data second"),
	})

	q2, err := queue.Pull()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(q2.String())
	if queue.Len() != 0 {
		t.Fatal("not zero length")
	}
}

func TestQueueEmptyPull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan mes.Inside)

	queue := newQueue(ctx, ch)
	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("first", "data first"),
	})
	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("second", "data second"),
	})
	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("third", "data third"),
	})
	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("four", "data four"),
	})

	for can, err := queue.Can(); err == nil; can, err = queue.Can() {
		if !can {
			return
		}
		e, err := queue.Pull()
		if err != nil {
			break
		}
		fmt.Println(e.String())
	}

	if queue.Len() == 0 {
		t.Fatal("not zero length")
	}
}
