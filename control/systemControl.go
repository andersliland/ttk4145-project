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

	. "../utilities"
)

var debugSystemControl = false

func InitSystemControl() {

}

func SystemControl(
	sendMessageChannel chan<- ElevatorOrderMessage,
	receiveOrderChannel <-chan ElevatorOrderMessage,
	sendBackupChannel chan<- ElevatorBackupMessage,
	receiveBackupChannel <-chan ElevatorBackupMessage,
	executeOrderChannel chan<- ElevatorOrderMessage,
	localIP string) {

	//var CabOrderMatrix [NumFloors]ElevatorOrder

	const watchdogKickTime = 100 * time.Millisecond
	const watchdogLimit = 3*watchdogKickTime + 10*time.Millisecond

	// Timers
	watchdogTimer := time.NewTicker(watchdogLimit)
	defer watchdogTimer.Stop()
	watchdogKickTimer := time.NewTicker(watchdogKickTime)
	defer watchdogKickTimer.Stop()

	// init states
	sendBackupChannel <- ElevatorBackupMessage{
		AskerIP:     localIP,
		Event:       EvRequestBackupState,
		ResponderIP: "",
		State:       ElevatorState{},
	}

	RegisteredElevators[localIP] = ResolveElevator(ElevatorState{LocalIP: localIP, LastFloor: 2})
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
		case msg := <-receiveBackupChannel:
			//log.Printf("[systemControl] receivedBackupChannel with event %v from %v]", EventType[msg.Event], msg.AskerIP)
			switch msg.Event {
			// resolved incomming alive, if timeout remove elevator
			case EvIAmAlive:
				if _, ok := RegisteredElevators[msg.ResponderIP]; ok { // check if a value exsist for ResponderIP
					RegisteredElevators[msg.ResponderIP].Time = time.Now() //update time for known elevator
				} else {
					printDebug("[systemControl] Received EvIAmAlive from a new elevator with IP" + msg.ResponderIP)
					RegisteredElevators[msg.ResponderIP] = ResolveElevator(msg.State)
				}
				updateWorkingElevators(RegisteredElevators, WorkingElevators, localIP, watchdogLimit)

				// inncomming backup state,
			case EvBackupState:
				printDebug("[systemControl] Received an EvBackupState from" + msg.AskerIP)

				// send out 'ElevatorBackupMessage' when receiving msg
			case EvRequestBackupState:
				printDebug("Received an EvRequestBackupState from" + msg.AskerIP)
				if msg.AskerIP != localIP {
					sendBackupChannel <- ElevatorBackupMessage{
						AskerIP:     msg.AskerIP,
						ResponderIP: localIP,
						Event:       EvBackupStateReturned,
						State:       ElevatorState{},
					}

				} else {
					printDebug("[systemControl] No stored state for elevator at selv " + localIP)
					/*
						sendBackupChannel <- ElevatorBackupMessage{
							AskerIP:     msg.AskerIP,
							ResponderIP: localIP,
							Event:       EvBackupStateReturned,
							State:       ElevatorState{},
						}
					*/

				}

			case EvBackupStateReturned:
				log.Printf("[systemControl] Received EvBackupStateReturned requested by me @ %v", localIP)

				if msg.AskerIP == localIP {

				} else {
					log.Printf("[systemControl] Received EvBackupStateReturned not requested by me")

				}
			case EvCabOrder:
				log.Printf("[systemControl] Received EvCabOrder from %v", msg.AskerIP)
				//save order in map map[string] bool

			default:
				log.Println("Received invalid ElevatorBackupMessage from", msg.ResponderIP)
			}

		// Order
		case order := <-receiveOrderChannel:
			//log.Printf("[systemControl] receivedBackupChannel with event %v from %v]", EventType[msg.Event], msg.AskerIP)
			printDebug("Recieved an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)
			// calculate cost
			// broadcast answer
			// sort incomming answer, wait for all elevator to reply
			// assign order to self if AssignedTo == localIP
			switch order.Event {
			case EvNewOrder:
				log.Printf("[systemControl] Received a new order from %v", order.OriginIP)

				switch HallOrderMatrix[order.Floor][order.ButtonType].Status {
				case NotActive:
					HallOrderMatrix[order.Floor][order.ButtonType].AssignedTo = order.AssignedTo
					HallOrderMatrix[order.Floor][order.ButtonType].Status = Awaiting

				case Awaiting:
					//empty

				case UnderExecution:
					//empty
				}
			case EvAckNewOrder:
				// received AckNewOrder, send out final assigned order

			case EvOrderConfirmed:
				// the order is confirmed, start executing

			case EvAckOrderConfirmed:

			case EvOrderDone:

			case EvAckOrderDone:
				// delete order from matrix and timer functions

			case EvReassignOrder:
				// error handling, if a order timer timeout, give the order to someone else

			default:
				printDebug("Received an invalid ElevatorOrderMessage from" + order.SenderIP)

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
				printDebug("[systemControl] Removed elevator " + RegisteredElevators[k].State.LocalIP + "in WorkingElevators")
				log.Printf("[systemControl] All active elevators %v", WorkingElevators)

			}
		} else { // watchdog not timed out
			if WorkingElevators[k] != true {
				WorkingElevators[k] = true
				printDebug("[systemControl] Added elevator " + RegisteredElevators[k].State.LocalIP + "in active elevators")
				log.Printf("[systemControl] All active elevators %v", WorkingElevators)
			}
		}
	}
}
func printDebug(s string) {
	if debugSystemControl {
		log.Println("[systemControl]", s)
	}
	if debugElevatorControl {
		log.Println("[elevatorControl]", s)
	}

}
