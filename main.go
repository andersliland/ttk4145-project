package main

import (
	//"./cost"
	"log"
	"os"
	"os/signal"
	"time"

	. "./config"
	"./driver"
	"./fsm"
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

func safeKill(safeKillChannel chan os.Signal, motorChannel chan int) {
	signal.Notify(safeKillChannel, os.Interrupt)
	<-safeKillChannel
	motorChannel <- MotorStop
	log.Fatal(ColorWhite, "User terminated program", ColorNeutral)

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
	const elevatorPollDelay = 5 * time.Millisecond

	sendMessageChannel := make(chan ElevatorOrderMessage, 5)
	sendBackupChannel := make(chan ElevatorOrderMessage, 5)
	receiveMessageChannel := make(chan ElevatorOrderMessage, 5)
	buttonChannel := make(chan driver.ElevButton, 10)
	lightChannel := make(chan driver.ElevLight, 10)
	motorChannel := make(chan int, 10)
	floorChannel := make(chan int, 10)
	safeKillChannel := make(chan os.Signal)

	go safeKill(safeKillChannel, motorChannel)
	go sendMessageChannelFunc(sendMessageChannel)
	go network.InitNetwork(sendMessageChannel, receiveMessageChannel, sendBackupChannel)
	go fsm.FSM()

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
			//log.Println("Floor:", a.Floor)
			//log.Println("OriginIP:", a.OriginIP)

		}
	}

}
