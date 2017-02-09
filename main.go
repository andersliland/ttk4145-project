package main

import (
	//"./cost"
	"./driver"
	//"./network"
	"log"
	"time"
)

func main() {
	log.Println("Activate: Elevator")
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
	err := driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay)
	if err != nil {
		log.Fatal(err)
	}

	driver.SetLight(1, 2)
}
