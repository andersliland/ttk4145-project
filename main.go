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


func main() {
	log.Println("ACTIVATE: Elevator")

	const elevatorPollDelay = 5 * time.Millisecond

	sendMessageChannel := make(chan ElevatorOrderMessage, 5)
	sendBackupChannel := make(chan ElevatorOrderMessage, 5)
	receiveMessageChannel := make(chan ElevatorOrderMessage, 5)
	buttonChannel := make(chan driver.ElevatorButton, 10)
	lightChannel := make(chan driver.ElevatorLight, 10)
	motorChannel := make(chan int, 10)
	floorChannel := make(chan int, 10)
	safeKillChannel := make(chan os.Signal, 10)

	var localIP string
	var err error
	localIP, err = network.InitNetwork(sendMessageChannel, receiveMessageChannel, sendBackupChannel)
	CheckError("ERROR [main] Could not initiate network", err)

	driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay)
	//driver.SetLight(1, 2)

	//go sendMessageChannelFunc(sendMessageChannel)
	go fsm.InitFSM()
	go fsm.FSM(buttonChannel, lightChannel, motorChannel, floorChannel, sendMessageChannel, receiveMessageChannel, localIP)


	// Kill motor when user terinates program
	signal.Notify(safeKillChannel, os.Interrupt)
	go func(){
		<-safeKillChannel
		motorChannel <- 0
		log.Fatal(ColorWhite, "\nUser terminated program\nMOTOR STOPPED\n", ColorNeutral)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
		}()

	log.Println("SUCCESS [main] Ready to run!!")
	select{}

}
