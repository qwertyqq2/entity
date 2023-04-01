package process

import (
	"fmt"
	"testing"

	mes "github.com/qwertyqq2/entity/message"
)

func TestQueuePush(t *testing.T) {
	queue := newQueue()

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

	q1 := queue.pull()
	fmt.Println(q1.String())

	q2 := queue.pull()
	fmt.Println(q2.String())

	q3 := queue.pull()
	fmt.Println(q3.String())

	if queue.len() != 0 {
		t.Fatal("not zero length")
	}
}

func TestQueuePull(t *testing.T) {

	queue := newQueue()

	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("first", "data first"),
	})

	q1 := queue.pull()
	fmt.Println(q1.String())

	queue.Push(Entry{
		op:   mes.Add,
		data: mes.NewInside("second", "data second"),
	})

	q2 := queue.pull()
	fmt.Println(q2.String())

	if queue.len() != 0 {
		t.Fatal("not zero length")
	}
}

func TestQueueEmptyPull(t *testing.T) {
	queue := newQueue()
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

	if queue.len() == 0 {
		t.Fatal("not zero length")
	}
}
