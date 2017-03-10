package main

import (
	//"./cost"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"./control"
	"./driver"
	"./network"
	. "./utilities"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	const elevatorPollDelay = 50 * time.Millisecond // Move to config?

	broadcastOrderChannel := make(chan OrderMessage, 5)
	receiveOrderChannel := make(chan OrderMessage, 5)
	broadcastBackupChannel := make(chan BackupMessage, 5)
	receiveBackupChannel := make(chan BackupMessage, 5)

	buttonChannel := make(chan ElevatorButton)
	lightChannel := make(chan ElevatorLight)
	motorChannel := make(chan int)
	floorChannel := make(chan int)

	newOrder := make(chan bool)

	safeKillChannel := make(chan os.Signal, 10)
	executeOrderChannel := make(chan OrderMessage, 10)

	var localIP string
	var err error
	localIP, err = network.Init(broadcastOrderChannel, receiveOrderChannel, broadcastBackupChannel, receiveBackupChannel)
	CheckError("ERROR [main]: Could not initiate network", err)

	// SIMULATOR Uncomment simulatorCore lines and Comment driver lines
	//simulatorCore.IOInit()                                                                         //Simulator init
	//simulatorCore.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay) // elevator init
	driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, elevatorPollDelay) // driver init

	log.Println("[main]\t\t New Elevator ready with IP:", localIP)
	control.Init(localIP)
	go control.SystemControl(newOrder, broadcastOrderChannel, receiveOrderChannel, broadcastBackupChannel, receiveBackupChannel, executeOrderChannel, localIP)
	go control.MessageLoop(newOrder,
		buttonChannel,
		lightChannel,
		motorChannel,
		floorChannel,
		broadcastOrderChannel,
		receiveOrderChannel,
		broadcastBackupChannel,
		receiveBackupChannel,
		OnlineElevators,
		ElevatorStatus,
		HallOrderMatrix,
		localIP)

	broadcastBackupChannel <- BackupMessage{
		AskerIP: localIP,
		Event:   EventElevatorBackup,
		State: Elevator{
			LocalIP: localIP,
		},
	}

	// Kill motor when user terminates program
	signal.Notify(safeKillChannel, os.Interrupt)
	go safeKill(safeKillChannel, motorChannel)

	select {} // Block main loop indefinitely

}

func safeKill(safeKillChannel <-chan os.Signal, motorChannel chan int) {
	<-safeKillChannel
	motorChannel <- Stop
	time.Sleep(10 * time.Millisecond) // wait for motor stop too be processed
	log.Fatal(ColorWhite, "\nUser terminated program\nMOTOR STOPPED\n", ColorNeutral)
	os.Exit(1)
}
