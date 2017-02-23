package control

// This module is responsible for the communication between all the elevators, organising the
// order matrix and queue, and delegating orders to spesific elevators.
//
// Orders are organised in a local and remote list. Each list contains

// The system support two types of messages
// send

import (
	"log"
	"time"

	. "../driver"

	. "../utilities"
)

var debug = true

func InitSystemControl() {

}

func SystemControl(
	sendMessageChannel chan ElevatorOrderMessage,
	receiveOrderChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	receiveBackupChannel chan ElevatorBackupMessage,
	executeOrderChannel chan ElevatorOrderMessage,
	buttonChannel chan ElevatorButton,
	lightChannel chan ElevatorLight,
	motorChannel chan int,
	floorChannel chan int,
	localIP string) {

	const watchdogKickTime = 100 * time.Millisecond
	const watchdogLimit = 3*watchdogKickTime + 10*time.Millisecond

	var knownElevators = make(map[string]*Elevator) // key = IPaddr
	var activeElevators = make(map[string]bool)     // key = IPaddr

	// Timers
	watchdogTimer := time.NewTicker(watchdogLimit)
	defer watchdogTimer.Stop()
	watchdogKickTimer := time.NewTicker(watchdogKickTime)
	defer watchdogKickTimer.Stop()

	goToFloorBelow(motorChannel, floorChannel)

	// init states
	log.Println("[systemControl] Send request for previous state")
	/*
		sendBackupChannel <- ElevatorBackupMessage{
			AskerIP:     localIP,
			ResponderIP: "blank",
			State:       ElevatorState{},
			Event:       EvRequestBackup,
		}
	*/

	knownElevators[localIP] = ResolveElevator(ElevatorState{LocalIP: localIP, LastFloor: 2})
	log.Println("knownElevators", knownElevators)
	updateActiveElevators(knownElevators, activeElevators, localIP, watchdogLimit)

	for {
		select {
		case <-watchdogKickTimer.C:
			sendBackupChannel <- ResolveWatchdogKickMessage(knownElevators[localIP])

		case <-watchdogTimer.C:
			updateActiveElevators(knownElevators, activeElevators, localIP, watchdogLimit)

		case order := <-receiveOrderChannel:
			//printDebug("Recieved an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)

			// calculate cost
			// broadcast answer
			// sort incomming answer, wait for all elevator to reply
			// assign order to self if AssignedTo == localIP

			order.AssignedTo = localIP
			order.Event = EvNewOrder
			executeOrderChannel <- order
			//log.Println("Execute order: AssignedTo", order.AssignedTo, " OriginIP:", order.OriginIP)
		}
	}
}

func updateActiveElevators(knownElevators map[string]*Elevator,
	activeElevators map[string]bool,
	localIP string,
	watchdogLimit time.Duration) {

	for k := range knownElevators {
		if time.Since(knownElevators[localIP].Time) > watchdogLimit {
			if activeElevators[localIP] == true {
				log.Printf("[systemControl] Removed elevator %s in activeElevators\n", knownElevators[k].State.LocalIP)
				delete(activeElevators, k)
			}
		} else {
			if activeElevators[k] != true {
				activeElevators[k] = true
				log.Printf("[systemControll] Added elevator %s in active elevators", knownElevators[k].State.LocalIP)
			}
		}
	}
}

func goToFloorBelow(motorChannel chan int, floorChannel chan int) {

	motorChannel <- MotorDown
	if <-floorChannel != FloorInvalid {
		motorChannel <- MotorStop
	}
}

func printDebug(s string) {
	if debug {
		log.Println("[systemControl]", s)
	}
}
