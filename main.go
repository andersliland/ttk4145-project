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
	"./network"
	"./simulator/simulatorCore"
	. "./utilities"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	const elevatorPollDelay = 5 * time.Millisecond // Move to config?

	sendMessageChannel := make(chan ElevatorOrderMessage, 5)
	receiveOrderChannel := make(chan ElevatorOrderMessage, 5)
	sendBackupChannel := make(chan ElevatorBackupMessage, 5)
	receiveBackupChannel := make(chan ElevatorBackupMessage, 5)

	buttonChannel := make(chan ElevatorButton)
	lightChannel := make(chan ElevatorLight)
	motorChannel := make(chan int)
	floorChannel := make(chan int)

	safeKillChannel := make(chan os.Signal, 10)
	executeOrderChannel := make(chan ElevatorOrderMessage, 10)

	var localIP string
	var err error
	localIP, err = network.Init(sendMessageChannel, receiveOrderChannel, sendBackupChannel, receiveBackupChannel)
	CheckError("ERROR [main]: Could not initiate network", err)

	simulatorCore.IOInit()                                                                         //Simulator init
	simulatorCore.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay) // elevator init
	//driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay) // driver init
	//log.Println("[main] Ready with IP:", localIP)

	go control.SystemControl(sendMessageChannel, receiveOrderChannel, sendBackupChannel, receiveBackupChannel, executeOrderChannel, localIP)

	go control.MessageLoop(buttonChannel,
		lightChannel,
		motorChannel,
		floorChannel,
		sendMessageChannel,
		receiveOrderChannel,
		sendBackupChannel,
		receiveBackupChannel,
		WorkingElevators,
		RegisteredElevators,
		OrderMatrix,
		localIP)

	// Kill motor when user terminates program
	signal.Notify(safeKillChannel, os.Interrupt)
	go safeKill(safeKillChannel, motorChannel)

	select {} // Block main loop indefinitely

}

func safeKill(safeKillChannel <-chan os.Signal, motorChannel chan int) {
	<-safeKillChannel
	//motorChannel <- MotorStop
	time.Sleep(10 * time.Millisecond) // wait for motor stop too be processed
	log.Fatal(ColorWhite, "\nUser terminated program\nMOTOR STOPPED\n", ColorNeutral)
	os.Exit(1)
}
