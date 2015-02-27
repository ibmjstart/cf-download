package main

import (
"fmt"
"time"
)

var messages = make(chan string)

func

func main() {

	go func() { messages <- "ping" }()
	for i := 0; i < 20; t++ {

	}
	msg := <-messages
	fmt.Println(msg)
}
