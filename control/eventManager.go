package control

import (
	"log"
	"strconv"
	"time"

	. "../utilities"

	"../orders"
)

const debugEventManager = true

// Need three functions in an orders.go file to work:
// ShouldStop(floor, direction, localIP)
// ChooseDirection(floor, direction, localIP)
// RemoveFloorOrders(floor, localIP)

func eventManager(
	newOrder chan bool,
	broadcastOrderChannel chan OrderMessage,
	floorReached chan int,
	lightChannel chan ElevatorLight,
	motorChannel chan int, localIP string) {

	var state int = Idle
	var floor int = FloorInvalid // to initialize or not to initialize?
	var direction int

	// if restore order from file do ..., else:
	const pollDelay = 5 * time.Millisecond
	doorTimeout := make(chan bool)
	doorTimerReset := make(chan bool)

	go doorTimer(doorTimeout, doorTimerReset)

	for {
		select {
		case <-newOrder:
			// if floor == FloorInvalid - goToFloorBelow
			switch state {
			case Idle:
				direction = setDirection(orders.ChooseDirection(floor, direction, localIP), localIP)

				if orders.ShouldStop(floor, direction, localIP) {
					printEventManager("Stopped at floor " + strconv.Itoa(floor+1))
					doorTimerReset <- true
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					state = setState(DoorOpen, localIP)

				} else {
					motorChannel <- direction
					state = setState(Moving, localIP)
					//newState <- Moving
				}
			case Moving: // Ignore
			case DoorOpen:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
				}
			default:
				// Insert error handling
			}
		case floor = <-floorReached:
			ElevatorStatus[localIP].Floor = floor //  TODO: Confirm functionality of this assignment
			switch state {
			case Idle:
				printEventManager("Elevator reached floor " + strconv.Itoa(floor+1) + " in state IDLE")

			case Moving:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					motorChannel <- Stop
					state = setState(DoorOpen, localIP)
				}
			case DoorOpen:
				// not applicable
			default:
				// Insert error handling
			}
		case <-doorTimeout:
			switch state {
			case Idle:
				// not applicable
			case Moving:
				// not applicable
			case DoorOpen:
				lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: false}
				orders.RemoveFloorOrders(floor, direction, localIP)
				printEventManager("eventDoorTimeout, Idle: direction: " + MotorStatus[direction+1])
				direction = setDirection(orders.ChooseDirection(floor, direction, localIP), localIP)
				printEventManager("Door closing, new direction is " + MotorStatus[direction+1] + ".  Elevator " + localIP)
				if direction == Stop {
					state = setState(Idle, localIP)
				} else {
					motorChannel <- direction
					state = setState(Moving, localIP)
				}
			default:
				// Insert error handling here - elevator might possibly need to be restarted ()
			}
			//case <-newState:
			//broadcastOrderChannel <- OrderMessage{}

		}
	}
}

func doorTimer(timeout chan<- bool, reset <-chan bool) {
	const doorOpenTime = 3 * time.Second
	timer := time.NewTimer(0)
	timer.Stop()
	for {
		select {
		case <-reset:
			timer.Reset(doorOpenTime)
		case <-timer.C:
			timer.Stop()
			timeout <- true
		}
	}
}

func setDirection(direction int, localIP string) int {
	ElevatorStatus[localIP].Direction = direction
	return direction

}

func setState(state int, localIP string) int {
	ElevatorStatus[localIP].State = state
	return state
}

func printEventManager(s string) {
	if debugEventManager {
		log.Println("[eventManager]\t", s)
	}
}
