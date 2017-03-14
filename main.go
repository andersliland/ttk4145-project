package main

import (
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

	broadcastOrderChannel := make(chan OrderMessage, 5)
	receiveOrderChannel := make(chan OrderMessage, 5)
	broadcastBackupChannel := make(chan BackupMessage, 5)
	receiveBackupChannel := make(chan BackupMessage, 5)

	orderCompleteChannel := make(chan OrderMessage, 5) // send OrderComplete from RemoveOrders to SystemControl

	buttonChannel := make(chan ElevatorButton, 10)
	lightChannel := make(chan ElevatorLight)
	motorChannel := make(chan int)
	floorChannel := make(chan int)

	newOrder := make(chan bool, 5)
	timeoutChannel := make(chan ExtendedHallOrder)

	safeKillChannel := make(chan os.Signal, 10)

	var onlineElevators = make(map[string]bool)

	var localIP string
	var err error
	localIP, err = network.Init(broadcastOrderChannel, receiveOrderChannel, broadcastBackupChannel, receiveBackupChannel)
	CheckError("ERROR [main]: Could not initiate network", err)

	driver.Init(buttonChannel, lightChannel, motorChannel, floorChannel, ElevatorPollDelay) // driver init

	//log.Println("[main]\t\t New Elevator ready with IP:", localIP)
	control.Init(localIP)
	go control.SystemControl(
		onlineElevators,
		motorChannel,
		newOrder,
		timeoutChannel,
		broadcastOrderChannel,
		receiveOrderChannel,
		broadcastBackupChannel,
		receiveBackupChannel,
		orderCompleteChannel,
		localIP)
	go control.MessageLoop(newOrder,
		buttonChannel,
		lightChannel,
		motorChannel,
		floorChannel,
		broadcastOrderChannel,
		receiveOrderChannel,
		broadcastBackupChannel,
		receiveBackupChannel,
		orderCompleteChannel,
		onlineElevators,
		ElevatorStatus,
		HallOrderMatrix,
		localIP)

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
