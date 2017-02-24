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

	var externalOrderMatrix [NumFloors][NumButtons]ElevatorOrder

	const watchdogKickTime = 100 * time.Millisecond
	const watchdogLimit = 3*watchdogKickTime + 10*time.Millisecond

	// key = IPaddr
	var knownElevators = make(map[string]*Elevator) // containing last known state
	var activeElevators = make(map[string]bool)

	// Timers
	watchdogTimer := time.NewTicker(watchdogLimit)
	defer watchdogTimer.Stop()
	watchdogKickTimer := time.NewTicker(watchdogKickTime)
	defer watchdogKickTimer.Stop()

	goToFloorBelow(motorChannel, floorChannel)

	// init states
	sendBackupChannel <- ElevatorBackupMessage{
		AskerIP: localIP,
		State:   ElevatorState{},
		Event:   EvRequestBackupState,
	}
	knownElevators[localIP] = ResolveElevator(ElevatorState{LocalIP: localIP, LastFloor: 2})
	updateActiveElevators(knownElevators, activeElevators, localIP, watchdogLimit)

	for {
		select {
		// Watchdog
		case <-watchdogKickTimer.C:
			sendBackupChannel <- ResolveWatchdogKickMessage(knownElevators[localIP])
			//log.Printf("[systemControl] Watchdog send IAmAlive from %v \n", localIP)

		case <-watchdogTimer.C:
			updateActiveElevators(knownElevators, activeElevators, localIP, watchdogLimit)
			//log.Println("[systemControl] Active Elevators", activeElevators)

			// Network
		case msg := <-receiveBackupChannel:
			switch msg.Event {
			// resolved incomming alive, if timeout remove elevator
			case EvIAmAlive:
				if _, ok := knownElevators[msg.ResponderIP]; ok { // check if a value exsist for ResponderIP
					knownElevators[msg.ResponderIP].Time = time.Now() //update time for known elevator
				} else {
					log.Println("[systemControl] Received EvIAmAlive from a new elevator with IP ", msg.ResponderIP)
					knownElevators[msg.ResponderIP] = ResolveElevator(msg.State)
				}
				updateActiveElevators(knownElevators, activeElevators, localIP, watchdogLimit)

				// inncomming backup state,
			case EvBackupState:

				// send out 'ElevatorBackupMessage' when receiving msg
			case EvRequestBackupState:
				if msg.AskerIP == localIP {
					log.Printf("[systemControl] Received an EvRequestBackupState from %v", msg.AskerIP)
					sendBackupChannel <- ElevatorBackupMessage{
						AskerIP:     msg.AskerIP,
						ResponderIP: localIP,
						Event:       EvBackupStateReturned,
						State:       ElevatorState{},
					}

				} else {
					log.Println("[systemControl] No stored state for elevator ", localIP)
				}

			case EvBackupStateReturned:
				if msg.AskerIP == localIP {
					log.Printf("[systemControl] Received EvBackupStateReturned requested by me", localIP)

				} else {
					log.Printf("[systemControl] REceived EvBackupStateReturned not requested by me")

				}

			default:
				log.Println("Received invalid ElevatorBackupMessage from", msg.ResponderIP)
			}

		// Order
		case order := <-receiveOrderChannel:
			//printDebug("Recieved an " + EventType[order.Event] + " from " + order.SenderIP + " with OriginIP " + order.OriginIP)
			// calculate cost
			// broadcast answer
			// sort incomming answer, wait for all elevator to reply
			// assign order to self if AssignedTo == localIP
			switch order.Event {
			case EvNewOrder:
				switch externalOrderMatrix[order.Floor][order.ButtonType].Status {
				case NotActive:

				case Awaiting:

				case UnderExecution:

				}
			case EvAckNewOrder:
			case EvOrderConfirmed:
			case EvAckOrderConfirmed:
			case EvOrderDone:
			case EvAckOrderDone:
			case EvReassignOrder:
			default:
				printDebug("Received an invalid ElevatorOrderMessage from" + order.SenderIP)

			}

			order.AssignedTo = localIP
			order.Event = EvNewOrder
			executeOrderChannel <- order
			//log.Println("Execute order: AssignedTo", order.AssignedTo, " OriginIP:", order.OriginIP)
		}
	}
}

// checks
func updateActiveElevators(knownElevators map[string]*Elevator, activeElevators map[string]bool, localIP string, watchdogLimit time.Duration) {
	for k := range knownElevators {
		if time.Since(knownElevators[k].Time) > watchdogLimit { //watchdog timeout
			if activeElevators[k] == true {
				delete(activeElevators, k)
				log.Printf("[systemControl] Removed elevator %s in activeElevators\n", knownElevators[k].State.LocalIP)
				log.Printf("[systemControl] All active elevators %v", activeElevators)

			}
		} else { // watchdog not timed out
			if activeElevators[k] != true {
				activeElevators[k] = true
				log.Printf("[systemControl] Added elevator %s in active elevators", knownElevators[k].State.LocalIP)
				log.Printf("[systemControl] All active elevators %v", activeElevators)
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
