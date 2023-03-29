package process

import (
	"github.com/qwertyqq2/entity/entity"
)

type Entity = entity.Entity

// A function that is executed during the lifetime of a process
type ProcessFunc func(proc Process)

// Interface that writes data to entity
type Process interface {
	//Closing a process
	Shutdown()

	// Registration entity for process
	// Returns an error if the process is already registered or if the entity is closed
	Registration(ent entity.Entity) error

	//Adds data to the process, after which they will go to the entity
	Add(key, data string)

	//Creates a request to remove data from an entity
	Delete(key string)

	//Start a process
	Start(ent entity.Entity) error

	//Process id
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
