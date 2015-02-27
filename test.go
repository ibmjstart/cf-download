package main

import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup

/*func routine(i int) {
	defer wg.Done() // 3
	fmt.Printf("routine %v finished\n", i)
}

func main() {
	for i := 0; i < 10; i++ {
		wg.Add(1)     // 2
		go routine(i) // *
	}
	wg.Wait() // 4
	fmt.Println("main finished")
}*/

func listener(c chan int, quit chan int) {
	defer wg.Done()
	for {
		select {
		case x := <-c:
			fmt.Println(x)
		case <-quit:
			fmt.Println("quit")

			return
		}
	}

}

func main() {
	c := make(chan int)
	quit := make(chan int)
	wg.Add(1)
	go listener(c, quit)
	for i := 0; i < 10; i++ {
		c <- i
	}
	quit <- 0
	wg.Wait()
	fmt.Println("main finished")
}
