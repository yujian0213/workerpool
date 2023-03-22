package main

import (
	"fmt"
	"time"
)

func main() {
	p := New(5, WithPreAllocWorkers(false), WithBlock(false))
	time.Sleep(time.Second * 2)
	for i := 0; i < 10; i++ {
		err := p.Schedule(func() { time.Sleep(time.Second * 3) })
		if err != nil {
			fmt.Printf("task[%d]: error: %s\n", i, err.Error())
		}
	}
	p.Free()
}
