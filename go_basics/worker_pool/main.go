package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 2. Пул воркеров с контекстом
// Пул горутин, читающих задания из канала, с возможностью остановки по контексту.

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	wg := &sync.WaitGroup{}
	workers := 5

	jobsNum := 100000

	jobs := make(chan int, jobsNum)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case job, ok := <-jobs:
					if !ok {
						return
					}
					fmt.Println(job)
				case <-ctx.Done():
					fmt.Println("context was done")
					return
				}
			}
		}()
	}

	go func() {
		for i := range jobsNum {
			select {
			case jobs <- i * i:
			case <-ctx.Done():
				fmt.Println("producer остановлен досрочно")
				close(jobs)
				return
			}
		}
		close(jobs)
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigchan:
		cancel()
		fmt.Println("stopped by ctrl+c")
	case <-ctx.Done():
		fmt.Println("stopped by done context")
	}
	wg.Wait()
}
