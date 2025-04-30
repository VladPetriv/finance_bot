package worker

import (
	"context"
	"fmt"
	"sync"
)

type job[T any] struct {
	ID   string
	Data T
}

// Func is a function that handles a worker job.
type Func[T any] func(ctx context.Context, id string, data T) error

// Pool is a worker pool.
type Pool[T any] struct {
	workersCount int
	handlerFunc  Func[T]
	jobs         chan job[T]
	wg           *sync.WaitGroup
	dedup        map[string]struct{}
	mu           *sync.Mutex
}

// NewPool creates a new worker pool.
func NewPool[T any](workersCount int, handlerFunc Func[T]) *Pool[T] {
	return &Pool[T]{
		workersCount: workersCount,
		handlerFunc:  handlerFunc,
		jobs:         make(chan job[T]),
		wg:           &sync.WaitGroup{},
		dedup:        make(map[string]struct{}),
		mu:           &sync.Mutex{},
	}
}

// Start starts the number of workers that were passed in constructor.
func (p *Pool[T]) Start(ctx context.Context) {
	for range p.workersCount {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

func (p *Pool[T]) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker stopping due to context cancellation: %v\n", ctx.Err())
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}

			err := p.handlerFunc(ctx, job.ID, job.Data)
			if err != nil {
				fmt.Printf("handle job error: %v\n", err)
			}
			p.mu.Lock()
			delete(p.dedup, job.ID)
			p.mu.Unlock()
		}
	}
}

// Stop stops the worker pool.
func (p *Pool[T]) Stop() {
	close(p.jobs)
	p.wg.Wait()
}

// AddJob adds a new job to the worker pool.
func (p *Pool[T]) AddJob(id string, data T) {
	p.mu.Lock()
	_, ok := p.dedup[id]
	if ok {
		p.mu.Unlock()
		return
	}
	p.dedup[id] = struct{}{}
	p.mu.Unlock()

	p.jobs <- job[T]{ID: id, Data: data}
}
