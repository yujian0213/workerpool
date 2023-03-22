package main

import (
	"errors"
	"fmt"
	"sync"
)

const (
	defaultCapacity = 1
	maxCapacity     = 20
)

var ErrWorkerPoolFreed = errors.New(" workerPool freed")
var ErrNoIdleWorkerInPool = errors.New(" noIdleWorker in pool")

type Pool struct {
	capacity int // workerpool大小
	active   chan struct{}
	tasks    chan Task
	wg       sync.WaitGroup // 用于在pool销毁时等待所有worker退出
	quit     chan struct{}  // 用于通知各个worker退出的信号channel
	preAlloc bool           // 是否在创建pool的时候就预创建workers，默认值为：false
	block    bool           // 当pool满的情况下，新的Schedule调用是否阻塞当前goroutine。默认值：true  如果block = false，则Schedule返回ErrNoWorkerAvailInPool
}
type Task func()

func New(capacity int, opts ...Option) *Pool {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	if capacity > maxCapacity {
		capacity = maxCapacity
	}
	p := &Pool{
		capacity: capacity,
		tasks:    make(chan Task),
		quit:     make(chan struct{}),
		active:   make(chan struct{}, capacity),
	}
	for _, opt := range opts {
		opt(p)
	}
	fmt.Printf("workerpool start(preAlloc=%t)\n", p.preAlloc)
	if p.preAlloc {
		// create all goroutines and send into works channel
		for i := 0; i < p.capacity; i++ {
			p.newWorker(i + 1)
			p.active <- struct{}{}
		}
	}
	go p.run()
	return p
}
func (p *Pool) returnTask(t Task) {
	go func() {
		p.tasks <- t
	}()
}
func (p *Pool) run() {
	idx := len(p.active)
	if !p.preAlloc {
	loop:
		for t := range p.tasks {
			p.returnTask(t)
			select {
			case <-p.quit:
				return
			case p.active <- struct{}{}:
				idx++
				p.newWorker(idx)
			default:
				break loop
			}
		}
	}
	for {
		select {
		case <-p.quit:
			return
		case p.active <- struct{}{}:
			idx++
			p.newWorker(idx)
		}
	}
}
func (p *Pool) newWorker(i int) {
	p.wg.Add(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("worker[%03d]: recover panic[%s] and exit\n", i, err)
				<-p.active
			}
			p.wg.Done()
		}()
		fmt.Printf("worker[%03d]: start\n", i)
		for {
			select {
			case <-p.quit:
				fmt.Printf("worker[%03d]: exit\n", i)
				<-p.active
				return
			case t := <-p.tasks:
				fmt.Printf("worker[%03d]: receive a task\n", i)
				t()
			}
		}
	}()
}
func (p *Pool) Schedule(t Task) error {
	select {
	case <-p.quit:
		return ErrWorkerPoolFreed
	case p.tasks <- t:
		return nil
	default:
		if p.block {
			p.tasks <- t
			return nil
		}
		return ErrNoIdleWorkerInPool
	}
}
func (p *Pool) Free() {
	close(p.quit) // make sure all worker and p.run exit and schedule return error
	p.wg.Wait()
	fmt.Printf("workerpool freed(preAlloc=%t)\n", p.preAlloc)
}
