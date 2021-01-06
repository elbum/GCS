package main

import (
	"log"
	"sync"
	"time"
)

const tasksPerSecond = 1

func main() {
	totalTasks := 100
	concurrency := 5

	var wg sync.WaitGroup
	wg.Add(concurrency)

	log.Println("Starting ...")

	for i := 0; i < concurrency; i++ {
		go func(count int) {
			runWorker(count, totalTasks/concurrency)
			wg.Done()
		}(i)
	}
	wg.Wait()
	log.Println("... Done")
}

func runWorker(tn, n int) {
	log.Printf("task: %d n: %d\n", tn, n)
	var throttle <-chan time.Time
	if tasksPerSecond > 0 {
		throttle = time.Tick(time.Duration(1e6/(tasksPerSecond)) * time.Microsecond)
	}

	for i := 0; i < n; i++ {
		if tasksPerSecond > 0 {
			<-throttle
		}
		log.Printf("doing task %d\n", tn)
	}
}
