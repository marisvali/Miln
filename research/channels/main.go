package main

import (
	"fmt"
	"time"
)

// Question: does a channel act like a queue, with a producer putting things in
// and a consumer taking things out? If yes, does and I can keep adding things to
// it even if the one taking things out of the channel is slower than ?
// Answer: not really. A channel by default acts like a queue with a maximum
// size of 1. But we can specify a larger maximum size when creating it and then
// it will have the effect the questions mentions (with blocking behavior when
// the queue gets full).

// Typical output
// UseChannelDefault
// producer: before pushing package 1
// consumer: before receiving package
// consumer: after receiving package 1
// producer: after pushing package 1
// producer: before pushing package 2
// consumer: after processing package 1
// consumer: before receiving package
// producer: after pushing package 2
// producer: before pushing package 3
// consumer: after receiving package 2
// consumer: after processing package 2
// consumer: before receiving package
// consumer: after receiving package 4
// producer: after pushing package 4
// producer: before pushing package 5
// consumer: after processing package 4
// consumer: before receiving package
// consumer: after receiving package 5
// producer: after pushing package 5
// consumer: after processing package 5
// consumer: before receiving package
// UseChannelWithCustomSize
// producer: before pushing package 1
// producer: after pushing package 1
// producer: before pushing package 2
// producer: after pushing package 2
// producer: before pushing package 3
// producer: after pushing package 3
// producer: before pushing package 4
// producer: after pushing package 4
// producer: before pushing package 5
// producer: after pushing package 5
// consumer: before receiving package
// consumer: after receiving package 1
// consumer: after processing package 1
// consumer: before receiving package
// consumer: after receiving package 2
// consumer: after processing package 2
// consumer: before receiving package
// consumer: after receiving package 3
// consumer: after processing package 3
// consumer: before receiving package
// consumer: after receiving package 4
// consumer: after processing package 4
// consumer: before receiving package
// consumer: after receiving package 5
// consumer: after processing package 5
// consumer: before receiving package

type data struct {
	packetIndex int
	x2          string
}

func ConsumeIncomingData(ch chan data) {
	for {
		// Receive data from the channel.
		// Blocks until data is received.
		fmt.Println("consumer: before receiving package")
		d := <-ch
		fmt.Println("consumer: after receiving package", d.packetIndex)

		// Process the data.
		time.Sleep(1 * time.Second)
		fmt.Println("consumer: after processing package", d.packetIndex)
	}
}

func UseChannelDefault() {
	fmt.Println("UseChannelDefault")
	ch := make(chan data)
	go ConsumeIncomingData(ch)

	// Produce and send data. This happens instantly.
	for i := 1; i <= 5; i++ {
		fmt.Println("producer: before pushing package", i)
		ch <- data{i, "i"}
		fmt.Println("producer: after pushing package", i)
	}
	// Wait for ConsumeIncomingData to do its job.
	time.Sleep(6 * time.Second)
}

func UseChannelWithCustomSize() {
	fmt.Println("UseChannelWithCustomSize")
	ch := make(chan data, 10)
	go ConsumeIncomingData(ch)

	// Produce and send data. This happens instantly.
	for i := 1; i <= 5; i++ {
		fmt.Println("producer: before pushing package", i)
		ch <- data{i, "i"}
		fmt.Println("producer: after pushing package", i)
	}
	// Wait for ConsumeIncomingData to do its job.
	time.Sleep(6 * time.Second)
}

func main() {
	UseChannelDefault()
	UseChannelWithCustomSize()
}
