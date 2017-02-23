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
			log.Printf("[systemControl] Watchdog send IAmAlive from %v \n", localIP)

		case <-watchdogTimer.C:
			updateActiveElevators(knownElevators, activeElevators, localIP, watchdogLimit)
			log.Println("[systemControl] Active Elevators", activeElevators)

			// Network
		case f := <-receiveBackupChannel:
			switch f.Event {
			case EvIAmAlive:
				if _, ok := knownElevators[f.ResponderIP]; ok { // check if a value exsist for ResponderIP
					knownElevators[f.ResponderIP].Time = time.Now() //update time for known elevator
				} else {
					log.Println("[systemControl] Received EvIAmAlive from a new elevator with IP ", f.ResponderIP)
					knownElevators[f.ResponderIP] = ResolveElevator(f.State)
				}
				updateActiveElevators(knownElevators, activeElevators, localIP, watchdogLimit)

			case EvBackupState:

			case EvRequestBackupState:
				log.Println("[systemControl] From EvRequestBackupState")

			case EvBackupStateReturned:

			default:
				log.Println("Received invalid ElevatorBackupMessage from", f.ResponderIP)
			}

		// Order
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

// checks
func updateActiveElevators(knownElevators map[string]*Elevator, activeElevators map[string]bool, localIP string, watchdogLimit time.Duration) {
	for k := range knownElevators {
		//log.Println(time.Since(knownElevators[localIP].Time), watchdogLimit)
		if time.Since(knownElevators[localIP].Time) > watchdogLimit { //watchdog timeout
			log.Println("[systemControl] watchdog timeout")
			if activeElevators[localIP] == true {
				log.Printf("[systemControl] Removed elevator %s in activeElevators\n", knownElevators[k].State.LocalIP)
				delete(activeElevators, k)
			}
		} else { // watchdog not timed out
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
