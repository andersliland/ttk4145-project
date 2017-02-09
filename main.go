package main

import (
	//"./cost"
	"./driver"
	//"./network"
	"log"
	"time"
)

func main() {
	log.Println("ACTIVATE: Elevator")
	/*
		driver.Elevator()
		driver.Io()
		network.Network()
		cost.Cost()
		network.Udp()

		network.InitUDP()
	*/

	const elevatorPollDelay = 5 * time.Millisecond

	buttonChannel := make(chan driver.ElevButton, 10)
	lightChannel := make(chan driver.ElevLight, 10)
	motorChannel := make(chan int, 10)
	floorChannel := make(chan int, 10)

	driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay)

	//driver.SetLight(1, 2)

	for {
		select {
		case a := <-buttonChannel:
			log.Println(a)
		case a := <-floorChannel:
			log.Println(a)
		}
	}

}
