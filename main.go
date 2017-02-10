package main

import (
	//"./cost"
	. "./config"
	"./driver"
	"./fsm"
	"./network"
	"log"
	"os"
	"os/signal"
	"time"
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

		time.Sleep(time.Second * 5)
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
	const elevatorPollDelay = 5 * time.Millisecond

	sendMessageChannel := make(chan ElevatorOrderMessage, 5)
	sendBackupChannel := make(chan ElevatorOrderMessage, 5)
	receiveMessageChannel := make(chan ElevatorOrderMessage, 5)
	buttonChannel := make(chan driver.ElevatorButton, 10)
	lightChannel := make(chan driver.ElevatorLight, 10)
	motorChannel := make(chan int, 10)
	floorChannel := make(chan int, 10)
	safeKillChannel := make(chan os.Signal)

	//go sendMessageChannelFunc(sendMessageChannel)
	go network.InitNetwork(sendMessageChannel, receiveMessageChannel, sendBackupChannel)
	go fsm.InitFSM()
	go fsm.FSM()
	go fsm.ButtonHandler(buttonChannel, lightChannel, motorChannel)

	driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay)
	//driver.SetLight(1, 2)

	signal.Notify(safeKillChannel, os.Interrupt)
	go func(){
		<-safeKillChannel
		motorChannel <- 0
		log.Fatal(ColorWhite, "\nUser terminated program\nMOTOR STOPPED\n", ColorNeutral)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
		}()


	for {
		select {
		case a := <-receiveMessageChannel:
			log.Println("Main receive: ", a)
			//log.Println("Floor:", a.Floor)
			//log.Println("OriginIP:", a.OriginIP)

		}
	}

}
