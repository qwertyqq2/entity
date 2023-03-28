package mes

import "errors"

type Op int

const (
	Add Op = iota + 1
	Get
	Has
	Delete
	Ping
	Success
	Fail
)

type Inside struct {
	Key  string
	Data string
}

func NewInside(key, data string) Inside {
	return Inside{
		Key:  key,
		Data: data,
	}
}

func (in Inside) isNil() bool {
	return in.Key == ""
}

type Message struct {
	Inside Inside

	Op Op
}

func (m *Message) IsNil() bool {
	switch m.Op {
	case Add, Get, Has, Delete, Fail, Success, Ping:
		return false
	default:
		return true
	}
}

func AddMessage(key, data string) Message {
	return Message{
		Inside: NewInside(key, data),
		Op:     Add,
	}
}

func DeleteMessage(key string) Message {
	return Message{
		Inside: NewInside(key, ""),
		Op:     Delete,
	}
}

func PingMessage() Message {
	return Message{
		Inside: Inside{},
		Op:     Ping,
	}
}

func SuccessMessage() Message {
	return Message{
		Inside: Inside{},
		Op:     Success,
	}
}

func FailMessage(err error) Message {
	errstr := err.Error()
	return Message{
		Inside: Inside{Data: errstr},
		Op:     Fail,
	}
}

func (m Message) IsFail() bool {
	return m.Op == Fail
}

func (m Message) IsSuccess() bool {
	return m.Op == Success
}

func (m Message) Size() int {
	return len(m.Inside.Data)
}

func (m Message) Error() error {
	if m.Op != Fail {
		return nil
	}
	err := errors.New(m.Inside.Data)
	return err
}
