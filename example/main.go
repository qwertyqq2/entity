package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qwertyqq2/entity/entity"
	"github.com/qwertyqq2/entity/process"
)

var (
	durWrite  = 20 * time.Second
	procLimit = 10
)

func writeToProcess(dur time.Duration) func(p process.Process) {
	return func(p process.Process) {
		ctx, cancel := context.WithTimeout(context.Background(), dur)
		defer cancel()
		n := 0
		for {
			select {
			case <-ctx.Done():
				return

			default:
				key := "singal " + fmt.Sprintf("proc %d , num %d", p.ID(), n)
				n++
				p.Add(key, "newdata")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func main() {
	ent := entity.New(10)

	if _, err := process.WithProcessFunc(0, writeToProcess(durWrite), ent, process.DefaultOpts()...); err != nil {
		log.Fatal(err)
	}
	if _, err := process.WithProcessFunc(1, writeToProcess(durWrite), ent, process.DefaultOpts()...); err != nil {
		log.Fatal(err)
	}

	after := time.After(durWrite)
	<-after

	fmt.Println(ent.String())
}
