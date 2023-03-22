package main

import "time"

func main() {
	p := New(5)
	for i := 0; i < 10; i++ {
		err := p.Schedule(func() {
			time.Sleep(1 * time.Second)
		})
		if err != nil {
			println("task: ", i, "err:", err)
		}
	}
	//p.Free()
}
