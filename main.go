// The main problem of the orignial code is that it doesn't really finish run for me.
// It hogs my CPU and if I've waited, I guess it would end up with channel RW dealock anyway.
// The changes here are similar to the fix/naive branch in terms of making it run well:
// 	1) workers should process jobs synchronisely, if you want more concurrency, increase the numebr of workers
//  2) as a job processing pipeline, the job dispatcher should be run synchronisely in the most case since the 
//     bottle neck should normally be the process step
//  3) better use of chananels in general.
// What's in additon to the fix/naive branch is that this packs the processing into a Pool type to make
// the code more strat forward, instead firing goroutines left and right. 
package main

import (
	"sync"
	"fmt"
)

// Pool is a simpler container of jobs
type Pool struct{
	stop chan struct{}
	Jobs chan int
	Results chan int
	Size int
}

// NewPool gives you a new Pool. Common design pattern would also call "Start", 
// but this is just a simple demo.
func NewPool(size int) *Pool {
	return &Pool{
		stop: make(chan struct{}),
		Jobs: make(chan int, size),
		Results: make(chan int),
		Size: size,
	}
}

// Start start the pool, it will listen on Jobs channel.
// Inside, we start a goroutine because we don't want to block the caller
// Then we start multiple goroutines that represent a worker. For a more 
// comperhensive system, we would track the workers in various ways, i.e. 
// give each worker a control channel, or make workers it's own type. 
// But for the purpose of this demo, we just use a weit group to get 
// notified when all the workers finished.
func (p *Pool) Start() {
	go func() {
		wg := sync.WaitGroup{}
		for i := 0; i < p.Size; i++ {
			wg.Add(1)
			go func() {
				p.WorkerDo(i+1)
				wg.Done()
			}()
		}
		wg.Wait()
		close(p.Jobs)
		close(p.Results)
	}()
}

// WorkerDo do a job, id is the "worker" id
// Didn't really understand the point of assigning to j in the original algorithm.
// So I removed all the that, hopefully it still does what was intented.
func (p *Pool) WorkerDo(id int) {
	done := false
	for !done {
		select {
		case j := <- p.Jobs:
			switch j % 3 {
			case 0:
				continue
			case 1:
				p.Results <- j * 2 * 2
			case 2:
				p.Results <- j * 3
			}
		case <- p.stop:
			done = true
		}
	}
}

// Close close the pool, call this when you finished submitting all the jobs
func (p *Pool) Close() {
	p.stop <- struct{}{}
	close(p.stop)
}

func main() {

	pool := NewPool(1000)
	pool.Start()

	// Add jobs
	// The change here is instead of firing each job in it's seperate goroutine,
	// we run the whole loop in a goroutine since firing that many goroutines in
	// this senario hurts (we too much overhead w/ no benefit since our jobs
	// channel was blocking almost immediately.
	// However, the produedure has to be run in a goroutine to aviable deadlock.
	go func() {
		jobCount := 0
		fmt.Println("Creating jobs...")
		for i := 1; i <= 1000000000; i ++ {
			jobCount ++
			if i % 2 == 0 {
				i += 99
			}
			pool.Jobs <- i
		}
		fmt.Println(jobCount, "jobs created")
		pool.Close()
	}()
	
	var sum int32
	for r := range pool.Results {
		sum += int32(r)
	}
	fmt.Println(sum)
}