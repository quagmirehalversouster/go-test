package main

import (
	"sync"
	"fmt"
)

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		switch j % 3 {
		case 0:
			continue
		case 1:
			results <- j * 2 * 2
		case 2:
			results <- j * 3
		}
	}
}

func main() {
	jobs := make(chan int)
	results := make(chan int)

	go func() {
		fmt.Println("Creating jobs...")
		k := 0
		for i := 1; i <= 1000000000; i++ {
			k += 1
			if i%2 == 0 {
				i += 99
			}
			jobs <- i
		}
		fmt.Println(k, "jobs created")
		close(jobs)
	}()

	go func() {
		poolSize := 1000
		wg := sync.WaitGroup{}
		for id := 1; id <= poolSize; id++ {
			wg.Add(1)
			go func() {
				worker(id, jobs, results)
				wg.Done()
			}()
		}
		wg.Wait()
		close(results)	
	}()

	var sum int32
	for r := range results {
		sum += int32(r)
	}
	fmt.Println(sum)
}