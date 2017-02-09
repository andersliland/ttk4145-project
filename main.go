package main

import (
	//"./cost"
	"log"
	"time"

	. "./config"
	"./driver"
	"./network"
)

func sendMessageChannelFunc(sendMessageChannel chan ElevatorOrderMessage) {

	for {
		// struct with JSON formatting
		sendMessageChannel <- ElevatorOrderMessage{
			Floor:      3,
			ButtonType: 3,
			AssignedTo: "one",
			OriginIP:   "I.P",
			SenderIP:   "sender IP",
			Event:      23,
		}

		time.Sleep(time.Second * 1)
		log.Println("Main send = ")
	}
}

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
	sendMessageChannel := make(chan ElevatorOrderMessage, 5)
	sendBackupChannel := make(chan ElevatorOrderMessage, 5)
	receiveMessageChannel := make(chan ElevatorOrderMessage, 5)

	go sendMessageChannelFunc(sendMessageChannel)

	go network.InitNetwork(sendMessageChannel, receiveMessageChannel, sendBackupChannel)

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
		case a := <-receiveMessageChannel:
			log.Println("Main receive: ", a)
			log.Println("Floor:", a.Floor)
			log.Println("OriginIP:", a.OriginIP)

		}
	}

}
