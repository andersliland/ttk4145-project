package main

import (
	//"./cost"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"./control"
	//"./driver"
	. "./simulator/simulatorCore"
	"./network"
	. "./utilities"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	const elevatorPollDelay = 5 * time.Millisecond // Move to config?

	sendMessageChannel := make(chan ElevatorOrderMessage, 5)
	receiveOrderChannel := make(chan ElevatorOrderMessage, 5)
	sendBackupChannel := make(chan ElevatorBackupMessage, 5)
	receiveBackupChannel := make(chan ElevatorBackupMessage, 5)

	buttonChannel := make(chan ElevatorButton, 10)
	lightChannel := make(chan ElevatorLight, 10)
	motorChannel := make(chan int, 10)
	floorChannel := make(chan int, 10)

	safeKillChannel := make(chan os.Signal, 10)

	executeOrderChannel := make(chan ElevatorOrderMessage, 10)

	var localIP string
	var err error
	localIP, err = network.Init(sendMessageChannel, receiveOrderChannel, sendBackupChannel, receiveBackupChannel)
	CheckError("ERROR [main]: Could not initiate network", err)

	IOInit()
	// Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay) // driver init

	control.InitElevatorControl()
	//go control.MessageLoop(buttonChannel, lightChannel, motorChannel, floorChannel, sendMessageChannel, receiveOrderChannel, sendBackupChannel, receiveBackupChannel, localIP)
	//go control.FSM(buttonChannel, lightChannel, motorChannel, floorChannel, sendMessageChannel, receiveOrderChannel, sendBackupChannel, receiveBackupChannel, executeOrderChannel, localIP)

	go control.SystemControl(sendMessageChannel, receiveOrderChannel, sendBackupChannel, receiveBackupChannel, executeOrderChannel, buttonChannel, lightChannel, motorChannel, floorChannel, localIP)

	// Kill motor when user terminates program
	signal.Notify(safeKillChannel, os.Interrupt)
	go safeKill(safeKillChannel, motorChannel)

	log.Println("[main] SUCCESS Elevator ready!")

	// initialise this elevator
	// send out request for previous state
	sendBackupChannel <- ElevatorBackupMessage{
		AskerIP: localIP,
		State:   ElevatorState{},
		Event:   EvRequestBackupState,
	}

	select {}

}

func safeKill(safeKillChannel <-chan os.Signal, motorChannel chan int) {
	<-safeKillChannel
	motorChannel <- MotorStop
	time.Sleep(10 * time.Millisecond) // wait for motor stop too be processed
	log.Fatal(ColorWhite, "\nUser terminated program\nMOTOR STOPPED\n", ColorNeutral)
	os.Exit(1)
}
