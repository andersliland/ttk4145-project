package control

import (
	"log"
	"strconv"
	"time"

	. "../utilities"

	"../orders"
)

const debugEventManager = true

const (
	idle = iota
	moving
	doorOpen
)

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

	var state int = idle
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
			switch state {
			case idle:
				printEventManager("Received new order: floor " + strconv.Itoa(floor+1))
				direction = orders.ChooseDirection(floor, direction, localIP)
				if orders.ShouldStop(floor, direction, localIP) {
					printEventManager("Stopped at floor " + strconv.Itoa(floor+1))
					doorTimerReset <- true
					orders.RemoveFloorOrders(floor, direction, localIP)
					//orders.RemoveFloorOrders(floor, direction, localIP broadcastOrderChannel) // change the above function with this later
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					state = doorOpen
				} else {
					motorChannel <- direction
					state = moving
					//newState <- moving
				}
			case moving: // Ignore
			case doorOpen:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
					orders.RemoveFloorOrders(floor, direction, localIP)
				}
			default:
				// Insert error handling
			}
		case floor = <-floorReached:
			ElevatorStatus[localIP].Floor = floor //  TODO: Confirm functionality of this assignment
			switch state {
			case idle:
				printEventManager("Elevator reached floor " + strconv.Itoa(floor+1) + " in state IDLE")

			case moving:
				if orders.ShouldStop(floor, direction, localIP) {
					doorTimerReset <- true
					orders.RemoveFloorOrders(floor, direction, localIP)
					lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: true}
					//direction = MotorStop
					motorChannel <- MotorStop
					state = doorOpen
				}
			case doorOpen:
				// not applicable

			default:
				// Insert error handling
			}
		case <-doorTimeout:
			switch state {
			case idle:
				// not applicable

			case moving:
				// not applicable
			case doorOpen:
				lightChannel <- ElevatorLight{Kind: DoorIndicator, Active: false}
				printEventManager("eventDoorTimeout, idle: direction: " + MotorStatus[direction])
				direction = orders.ChooseDirection(floor, direction, localIP)
				printEventManager("Door closing, new direction is " + MotorStatus[direction] + ".  Elevator " + localIP)
				if direction == MotorStop {
					state = idle
				} else {
					motorChannel <- direction
					state = moving
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

func printEventManager(s string) {
	if debugEventManager {
		log.Println("[eventManager]\t", s)
	}
}
