package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan interface{})
	terminated := doWork(done, nil)

	go func() {
		time.Sleep(time.Second * 1)
		fmt.Println("Canceling doWork...")
		close(done)
	}()

	<-terminated
	fmt.Println("done")
}

func doWork(done <-chan interface{}, s <-chan string) <-chan interface{} {
	terminated := make(chan interface{})
	go func() {
		defer fmt.Println("doWork exited!")
		defer close(terminated)
		fmt.Println("do Work running")
		for {
			select {
			case <-done:
				return
			case s := <-s:
				fmt.Println(s)
			}
		}
	}()
	return terminated
}
