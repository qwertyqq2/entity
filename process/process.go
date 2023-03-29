package process

import (
	"github.com/qwertyqq2/entity/entity"
)

type Entity = entity.Entity

type ProcessFunc func(proc Process)

type Process interface {
	Shutdown()

	Registration(ent entity.Entity) error

	Add(key, data string)

	Delete(key string)

	Start(ent entity.Entity) error

	ID() int
}

func WithProcessFunc(num int, fn ProcessFunc, ent Entity, opts ...Option) (Process, error) {
	proc := newProc(num)

	for _, o := range opts {
		o(proc)
	}

	if err := proc.Start(ent); err != nil {
		return nil, err
	}

	go func() {
		fn(proc)
		proc.Shutdown()
	}()

	return proc, nil
}

func WithEntity(num int, ent Entity, opts ...Option) (Process, error) {
	proc := newProc(num)

	for _, o := range opts {
		o(proc)
	}

	if err := proc.Start(ent); err != nil {
		return nil, err
	}

	return proc, nil
}

func Shutdown(procs ...Process) {
	for _, proc := range procs {
		proc.Shutdown()
	}
}
