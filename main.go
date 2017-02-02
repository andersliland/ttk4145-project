package main

import (
	//"./cost"
	"./driver"
	//"./network"
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("Hello World from main")

	/*
		driver.Elevator()
		driver.Io()
		network.Network()
		cost.Cost()
		network.Udp()

		network.InitUDP()
	*/

	const elevatorPollDelay = 50 * time.Millisecond

	buttonChannel := make(chan driver.ElevButton, 10)
	lightChannel := make(chan driver.ElevLight)
	motorChannel := make(chan int)
	floorChannel := make(chan int)

	if err := driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Success: Driver initialization")
	}
}
