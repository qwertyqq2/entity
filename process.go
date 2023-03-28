package process

import (
	"time"

	"github.com/qwertyqq2/entity/entity"
	"github.com/qwertyqq2/entity/process"
)

type Entity = entity.Entity

type ProcessFunc func(proc Process)

type Process interface {
	Shutdown()

	Registration(ent entity.Entity) error

	Add(key, data string)

	Delete(key string)

	Start(ent entity.Entity) error
}

func WithProcessFunc(num int, fn ProcessFunc, ent Entity) (Process, error) {
	proc := process.New(num)

	if err := proc.Start(ent); err != nil {
		return nil, err
	}

	go func() {
		fn(proc)
		proc.Shutdown()
	}()

	return proc, nil
}

func WithEntity(num int, ent Entity) (Process, error) {
	proc := process.New(num)

	if err := proc.Start(ent); err != nil {
		return nil, err
	}

	return proc, nil
}

func WithIntervals(num int, sendTimeoutInterval time.Duration, respTimeoutInterval time.Duration) Process {
	proc := process.New(num)

	proc.SetResentInterval(sendTimeoutInterval)
	proc.SetWaitRespInterval(respTimeoutInterval)

	return proc
}

func Shutdown(procs ...Process) {
	for _, proc := range procs {
		proc.Shutdown()
	}
}
