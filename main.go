package main

import (
	//"./cost"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	. "./config"
	"./driver"
	"./fsm"
	"./network"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	const elevatorPollDelay = 5 * time.Millisecond

	sendMessageChannel := make(chan ElevatorOrderMessage, 5)
	receiveOrderChannel := make(chan ElevatorOrderMessage, 5)
	sendBackupChannel := make(chan ElevatorBackupMessage, 5)
	receiveBackupChannel := make(chan ElevatorBackupMessage, 5)

	buttonChannel := make(chan driver.ElevatorButton, 10)
	lightChannel := make(chan driver.ElevatorLight, 10)
	motorChannel := make(chan int, 10)
	floorChannel := make(chan int, 10)
	safeKillChannel := make(chan os.Signal, 10)

	var localIP string
	var err error
	localIP, err = network.InitNetwork(sendMessageChannel, receiveOrderChannel, sendBackupChannel, receiveBackupChannel)
	CheckError("ERROR [main]: Could not initiate network", err)

	driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay)

	go fsm.InitFSM()
	go fsm.FSM(buttonChannel, lightChannel, motorChannel, floorChannel, sendMessageChannel, receiveOrderChannel, sendBackupChannel, receiveBackupChannel, localIP)

	// Kill motor when user terminates program
	signal.Notify(safeKillChannel, os.Interrupt)
	go func() {
		<-safeKillChannel
		motorChannel <- MotorStop
		time.Sleep(10 * time.Millisecond) // wait for motor stop too be processed
		log.Fatal(ColorWhite, "\nUser terminated program\nMOTOR STOPPED\n", ColorNeutral)
		os.Exit(1)
	}()

	log.Println("SUCCESS [main]: Elevator ready!")
	select {}

}
