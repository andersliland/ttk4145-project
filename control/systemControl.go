package control

// This module is responsible for the communication between all the elevators, organising the
// order matrix and queue, and delegating orders to spesific elevators.
//
// Orders are organised in a local and remote list. Each list contains

// The system support two types of messages
// send

import (
	"log"

	. "../utilities"
)

var debug = true

func InitSystemControl() {

}

//
func SystemControl(
	sendMessageChannel chan ElevatorOrderMessage,
	receiveOrderChannel chan ElevatorOrderMessage,
	sendBackupChannel chan ElevatorBackupMessage,
	receiveBackupChannel chan ElevatorBackupMessage,
	executeOrderChannel chan ElevatorOrderMessage,
	localIP string) {

	for {
		select {
		case order := <-receiveOrderChannel:
			//printDebug("Recieved an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)

			// calculate cost
			// broadcast answer
			// sort incomming answer, wait for all elevator to reply
			// assign order to self if AssignedTo == localIP

			order.AssignedTo = localIP
			order.Event = EvExecuteOrder
			executeOrderChannel <- order
			//log.Println("Execute order: AssignedTo", order.AssignedTo, " OriginIP:", order.OriginIP)
		}
	}
}

func updateActiveElevators() {

}

func printDebug(s string) {
	if debug {
		log.Println("[systemControl]", s)
	}
}
