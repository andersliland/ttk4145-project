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

	"../cost"

	. "../utilities"
)

var debugSystemControl = true

func InitSystemControl() {

}

func SystemControl(
	sendMessageChannel chan<- ElevatorOrderMessage,
	receiveOrderChannel <-chan ElevatorOrderMessage,
	sendBackupChannel chan<- ElevatorBackupMessage,
	receiveBackupChannel <-chan ElevatorBackupMessage,
	executeOrderChannel chan<- ElevatorOrderMessage,
	localIP string) {

	const watchdogKickTime = 100 * time.Millisecond
	const watchdogLimit = 3*watchdogKickTime + 10*time.Millisecond

	// Timers
	watchdogTimer := time.NewTicker(watchdogLimit)
	defer watchdogTimer.Stop()
	watchdogKickTimer := time.NewTicker(watchdogKickTime)
	defer watchdogKickTimer.Stop()

	// init states
	sendBackupChannel <- ElevatorBackupMessage{
		AskerIP: localIP,
		Event:   EventRequestBackup,
		//ResponderIP: "",
		//State:       ElevatorState{},
	}

	RegisteredElevators[localIP] = ResolveElevator(Elevator{LocalIP: localIP, LastFloor: 2})
	updateWorkingElevators(RegisteredElevators, WorkingElevators, localIP, watchdogLimit)

	for {
		select {
		// Watchdog
		case <-watchdogKickTimer.C:
			sendBackupChannel <- ResolveWatchdogKickMessage(RegisteredElevators[localIP])
			//log.Printf("[systemControl] Watchdog send IAmAlive from %v \n", localIP)

		case <-watchdogTimer.C:
			updateWorkingElevators(RegisteredElevators, WorkingElevators, localIP, watchdogLimit)
			//log.Println("[systemControl] Active Elevators", WorkingElevators)

			// Network
		case backup := <-receiveBackupChannel:
			//log.Printf("[systemControl] receivedBackupChannel with event %v from %v]", EventType[backup.Event], backup.AskerIP)
			switch backup.Event {
			// resolved incomming alive, if timeout remove elevator
			case EventElevatorAlive:
				if _, ok := RegisteredElevators[backup.ResponderIP]; ok { // check if a value exsist for ResponderIP
					RegisteredElevators[backup.ResponderIP].Time = time.Now() //update time for known elevator
				} else {
					printSystemControl(" Received EventElevatorAlive from a new elevator with IP" + backup.ResponderIP)
					RegisteredElevators[backup.ResponderIP] =
						ResolveElevator(backup.State)
				}
				updateWorkingElevators(RegisteredElevators, WorkingElevators, localIP, watchdogLimit)

			case EventElevatorBackup:
				printSystemControl(" Received an EventElevatorBackup from" + backup.AskerIP)

			case EventRequestBackup:
				printSystemControl("Received an EventRequestBackup from" + backup.AskerIP)
				if backup.AskerIP != localIP {
					sendBackupChannel <- ElevatorBackupMessage{
						AskerIP:     backup.AskerIP,
						ResponderIP: localIP,
						Event:       EventElevatorBackupReturned,
						State:       Elevator{},
					}

				} else {
					printSystemControl(" No stored state for elevator at selv " + localIP)
					/*
						sendBackupChannel <- ElevatorBackupMessage{
							AskerIP:     backup.AskerIP,
							ResponderIP: localIP,
							Event:       EventElevatorBackupReturned,
							State:       ElevatorState{},
						}
					*/

				}

			case EventElevatorBackupReturned:
				log.Printf("[systemControl] Received EventElevatorBackupReturned from %v", backup.ResponderIP)
				if backup.AskerIP == localIP {
					// i requested this backup, update
					// OrderMatrix
					// CabOrderMatrix

				} else {
					log.Printf("[systemControl] Received EventElevatorBackupReturned not requested by me")

				}
			case EventCabOrder:
				printSystemControl(" Received EventCabOrder from " + backup.AskerIP)
				if backup.AskerIP == localIP {
					printSystemControl("Received an EventCabOrder from selv, ignoring")
				} else {
					//CabOrderMatrix[backup.State.LastFloor].Status = Awaiting

				}

				//save order in map map[string] bool

			case EventAckCabOrder:

			default:
				log.Println("Received invalid ElevatorBackupMessage from", backup.ResponderIP)
			}

		// Order
		case order := <-receiveOrderChannel:
			//log.Printf("[systemControl] receivedBackupChannel with event %v from %v]", EventType[order.Event], order.AskerIP)
			printSystemControl("Received an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)
			// calculate cost
			// broadcast answer
			// sort incomming answer, wait for all elevator to reply
			// assign order to self if AssignedTo == localIP
			switch order.Event {
			case EventNewOrder:
				log.Printf("[systemControl] Received a new order from %v", order.OriginIP)

				switch OrderMatrix[order.Floor][order.ButtonType].Status {
				case NotActive:
					OrderMatrix[order.Floor][order.ButtonType].AssignedTo = order.AssignedTo
					OrderMatrix[order.Floor][order.ButtonType].Status = Awaiting

				case Awaiting:
					//empty

				case UnderExecution:
					//empty
				}
			case EventAckNewOrder:
				// received AckNewOrder, send out final assigned order

			case EventOrderConfirmed:
				// the order is confirmed, start executing

			case EventAckOrderConfirmed:

			case EventOrderDone:

			case EventAckOrderDone:
				// delete order from matrix and timer functions

			default:
				printSystemControl("Received an invalid ElevatorOrderMessage from" + order.SenderIP)

			}

		} // select
	} // for
} //function

// removes elevator from 'WorkingElevators' if watchdog timeout
// adds elevator to 'WorkingElevators' if watchdog not timeout
func updateWorkingElevators(RegisteredElevators map[string]*Elevator, WorkingElevators map[string]bool, localIP string, watchdogLimit time.Duration) {
	for k := range RegisteredElevators {
		if time.Since(RegisteredElevators[k].Time) > watchdogLimit { //watchdog timeout
			if WorkingElevators[k] == true {
				delete(WorkingElevators, k)
				printSystemControl("[systemControl] Removed elevator " + RegisteredElevators[k].LocalIP + " in WorkingElevators")
				log.Printf("[systemControl] All Working elevators %v", WorkingElevators)

			}
		} else { // watchdog not timed out
			if WorkingElevators[k] != true {
				WorkingElevators[k] = true
				printSystemControl("[systemControl] Added elevator " + RegisteredElevators[k].LocalIP + " in WorkingElevators")
				log.Printf("[systemControl] All WorkingElevators %v", WorkingElevators)
			}
		}
	}
}
func printSystemControl(s string) {
	if debugSystemControl {
		log.Println("[systemControl]", s)
	}
}
